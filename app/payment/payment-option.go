package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/flarehotspot/sdk/api/connmgr"
	"github.com/flarehotspot/sdk/api/http/flash"
	"github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/plugin"
	mdls "github.com/flarehotspot/wired-coinslot/app/models"
)

type pmtEvt struct {
	PaymentAmount  float64
	PaymentTotal   float64
	WalletDebit    float64
	WalletBal      float64
	WalletAvailBal float64
}

type PaymentOption struct {
	mu       sync.RWMutex
	api      plugin.IPluginApi
	provider *PaymentProvider
	coinslot *mdls.WiredCoinslot
	client   connmgr.IClientDevice
	purchase models.IPurchase
}

func (self *PaymentOption) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentOption) ErrResp(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, client connmgr.IClientDevice, purchase models.IPurchase) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.client != nil && self.client.Device().Id() != client.Device().Id() {
		err := purchase.Cancel(r.Context())
		if err != nil {
			self.ErrResp(w, err)
			return
		}
		self.api.HttpApi().Respond().SetFlashMsg(w, flash.FlashTypeError, "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	self.client = client
	self.purchase = purchase

	stat, err := purchase.Stat(r.Context())
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	data := map[string]interface{}{
		"PaymentAmount":  fmt.Sprintf("%0.2f", stat.PaymentAmount),
		"PaymentTotal":   fmt.Sprintf("%0.2f", stat.PaymentTotal),
		"WalletDebit":    fmt.Sprintf("%0.2f", stat.WalletDebit),
		"WalletBal":      fmt.Sprintf("%0.2f", stat.WalletBal),
		"WalletAvailBal": fmt.Sprintf("%0.2f", stat.WalletAvailBal),
		"UseWallet":      stat.WalletDebit > 0,
	}

	self.api.HttpApi().Respond().PortalView(w, r, "insert-coin.html", data)
}

func (self *PaymentOption) PaymentReceived(ctx context.Context, amount float64) {
	err := self.purchase.IncPayment(ctx, amount, nil, nil)
	if err != nil {
		log.Printf("Error while updating payment: %+v\n", err)
		return
	}
	data := map[string]any{"amount": amount}
	self.client.Emit("payment:received", data)
}

func (self *PaymentOption) UseWalletBal(w http.ResponseWriter, r *http.Request, debit float64) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	ctx := r.Context()
	tx, err := self.api.Db().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	payment, err := self.purchase.PaymentTx(tx, ctx)

	err = payment.UseWalletBalTx(tx, ctx, debit)
	if err != nil {
		log.Println(err)
		return err
	}

	stat, err := self.purchase.StatTx(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "applicatoin/json")
	json, err := json.Marshal(stat)
	if err != nil {
		return err
	}

	w.Write(json)

	return nil
}

func (self *PaymentOption) Done(w http.ResponseWriter, r *http.Request) {
	self.api.PaymentsApi().ExecCallback(w, r, self.purchase)
}

func NewPaymentOpt(api plugin.IPluginApi, prvdr *PaymentProvider, c *mdls.WiredCoinslot) *PaymentOption {
	return &PaymentOption{
		api:      api,
		provider: prvdr,
		coinslot: c,
		client:   nil,
	}
}
