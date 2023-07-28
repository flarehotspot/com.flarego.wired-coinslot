package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	mdls "github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/sdk/api/connmgr"
	"github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/utils/flash"
)

type pmtEvt struct {
	PaymentAmount  float64
	PaymentTotal   float64
	WalletDebit    float64
	WalletBal      float64
	WalletAvailBal float64
}

type PaymentOption struct {
	mu         sync.RWMutex
	api        plugin.IPluginApi
	provider   *PaymentProvider
	coinslot   *mdls.WiredCoinslot
	deviceId   *int64
	purchaseId *int64
}

func NewPaymentOpt(api plugin.IPluginApi, prvdr *PaymentProvider, c *mdls.WiredCoinslot) *PaymentOption {
	return &PaymentOption{
		api:      api,
		provider: prvdr,
		coinslot: c,
	}
}

func (self *PaymentOption) Name() string {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.coinslot.Name()
}

func (self *PaymentOption) ErrResp(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (self *PaymentOption) DeviceId() int64 {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return *self.deviceId
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, client connmgr.IClientDevice, purchase models.IPurchase) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.deviceId != nil && *self.deviceId != client.Id() {
		err := purchase.Cancel(r.Context())
		if err != nil {
			self.ErrResp(w, err)
			return
		}
		self.api.HttpApi().Respond().SetFlashMsg(w, flash.Error, "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	clntId := client.Id()
	purId := purchase.Id()
	self.deviceId = &clntId
	self.purchaseId = &purId

	stat, err := purchase.Stat(r.Context())
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	data := map[string]interface{}{
		"PaymentTotal":   fmt.Sprintf("%0.2f", stat.PaymentTotal),
		"WalletDebit":    fmt.Sprintf("%0.2f", stat.WalletDebit),
		"WalletBal":      fmt.Sprintf("%0.2f", stat.WalletBal),
		"WalletAvailBal": fmt.Sprintf("%0.2f", stat.WalletAvailBal),
		"UseWallet":      stat.WalletDebit > 0,
	}

	self.api.HttpApi().Respond().PortalView(w, r, "insert-coin.html", nil, data)
}

func (self *PaymentOption) PaymentReceived(ctx context.Context, clnt connmgr.IClientDevice, amount float64) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	purchase, err := self.api.Models().Purchase().Find(ctx, *self.purchaseId)
	if err != nil {
		log.Println(err)
		return
	}

	desc := ""
	_, err = purchase.AddPayment(ctx, amount, desc)
	if err != nil {
		log.Printf("Error while updating payment: %+v\n", err)
		return
	}
	data := map[string]any{"amount": amount}
	self.api.ClientMgr().SocketEmit(clnt, "payment:received", data)
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

	p, err := self.api.Models().Purchase().FindTx(tx, ctx, *self.purchaseId)
	if err != nil {
		return err
	}

	err = p.UpdateTx(tx, ctx, debit, p.WalletTxId(), p.CancelledAt(), p.ConfirmedAt(), nil)
	if err != nil {
		return err
	}

	stat, err := p.StatTx(tx, ctx)
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
	self.mu.RLock()
	defer self.mu.RUnlock()

	p, err := self.api.Models().Purchase().Find(r.Context(), *self.purchaseId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	self.api.PaymentsApi().ExecCallback(w, r, p)
}

func (self *PaymentOption) Cancel(w http.ResponseWriter, r *http.Request) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	clnt, err := self.api.ClientReg().CurrentClient(r)
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	pur, err := self.api.Models().Purchase().PendingPurchase(r.Context(), clnt.Id())
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	err = pur.Cancel(r.Context())
	if err != nil {
		self.ErrResp(w, err)
		return
	}

	self.api.HttpApi().Respond().SetFlashMsg(w, flash.Info, "Purchase was cancelled.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
