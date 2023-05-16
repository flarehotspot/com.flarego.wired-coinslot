package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/flarehotspot/sdk/api/devices"
	"github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/plugin"
	mdls "github.com/flarehotspot/wired-coinslot/app/models"
)

type pmtEvt struct {
	PaymentAmount  float64 `json:"amount"`
	WalletDebit    float64 `json:"wallet_debit"`
	TotalAmount    float64 `json:"total_amount"`
	WalletBal      float64 `json:"wallet_bal"`
	WalletAvailBal float64 `json:"wallet_avail_bal"`
}

type PaymentOption struct {
	mu       sync.RWMutex
	api      plugin.IPluginApi
	provider *PaymentProvider
	coinslot *mdls.WiredCoinslot
	client   devices.IClientDevice
	purchase models.IPurchase
}

func (self *PaymentOption) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentOption) ErrResp(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, client devices.IClientDevice, purchase models.IPurchase) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.client != nil && self.client.Device().Id() != client.Device().Id() {
		err := purchase.Cancel(r.Context())
		if err != nil {
			self.ErrResp(w, err)
			return
		}
		self.api.HttpApi().Respond().SetFlashMsg(w, "error", "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	self.client = client
	self.purchase = purchase

	wallet, err := self.client.Device().Wallet(r.Context())
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	data := map[string]interface{}{
		"walletBal": fmt.Sprintf("%0.2f", wallet.Balance()),
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
	if err != nil {
		log.Println(err)
		return err
	}

	err = payment.UseWalletBalTx(tx, ctx, debit)
	if err != nil {
		log.Println(err)
		return err
	}

	wallet, err := self.client.Device().WalletTx(tx, ctx)
	if err != nil {
		return err
	}

	availBal, err := wallet.AvailableBalTx(tx, ctx)
	if err != nil {
		return err
	}

	var dbt float64
	if payment.WalletDebit() != nil {
		dbt = *payment.WalletDebit()
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	resp := pmtEvt{
		PaymentAmount:  payment.Amount(),
		WalletDebit:    dbt,
		TotalAmount:    payment.TotalAmount(),
		WalletBal:      wallet.Balance(),
		WalletAvailBal: availBal,
	}

	w.Header().Set("Content-Type", "applicatoin/json")
	json, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	w.Write(json)

	return nil
}

func NewPaymentOpt(api plugin.IPluginApi, prvdr *PaymentProvider, c *mdls.WiredCoinslot) *PaymentOption {
	return &PaymentOption{
		api:      api,
		provider: prvdr,
		coinslot: c,
		client:   nil,
	}
}
