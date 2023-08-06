package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	mdls "github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/sdk/v1/api"
	"github.com/flarehotspot/sdk/v1/api/connmgr"
	"github.com/flarehotspot/sdk/v1/api/models"
	"github.com/flarehotspot/sdk/v1/utils/flash"
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
	api        api.IPluginApi
	provider   *PaymentProvider
	coinslot   *mdls.WiredCoinslot
	deviceId   *int64
	purchaseId *int64
}

func NewPaymentOpt(API api.IPluginApi, prvdr *PaymentProvider, c *mdls.WiredCoinslot) *PaymentOption {
	return &PaymentOption{
		api:      API,
		provider: prvdr,
		coinslot: c,
	}
}

func (opt *PaymentOption) Name() string {
	opt.mu.RLock()
	defer opt.mu.RUnlock()
	return opt.coinslot.Name()
}

func (opt *PaymentOption) ErrResp(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (opt *PaymentOption) DeviceId() int64 {
	opt.mu.RLock()
	defer opt.mu.RUnlock()
	return *opt.deviceId
}

func (opt *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, client connmgr.IClientDevice, purchase models.IPurchase) {
	opt.mu.Lock()
	defer opt.mu.Unlock()

	if opt.deviceId != nil && *opt.deviceId != client.Id() {
		err := purchase.Cancel(r.Context())
		if err != nil {
			opt.ErrResp(w, err)
			return
		}
		flash.SetFlashMsg(w, flash.Error, "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	clntId := client.Id()
	purId := purchase.Id()
	opt.deviceId = &clntId
	opt.purchaseId = &purId

	stat, err := purchase.Stat(r.Context())
	if err != nil {
		opt.ErrResp(w, err)
		return
	}

	data := map[string]interface{}{
		"PaymentTotal":   fmt.Sprintf("%0.2f", stat.PaymentTotal),
		"WalletDebit":    fmt.Sprintf("%0.2f", stat.WalletDebit),
		"WalletBal":      fmt.Sprintf("%0.2f", stat.WalletBal),
		"WalletAvailBal": fmt.Sprintf("%0.2f", stat.WalletAvailBal),
		"UseWallet":      stat.WalletDebit > 0,
	}

	opt.api.HttpApi().Respond().PortalView(w, r, "insert-coin.html", data)
}

func (opt *PaymentOption) PaymentReceived(ctx context.Context, clnt connmgr.IClientDevice, amount float64) {
	opt.mu.RLock()
	defer opt.mu.RUnlock()

	purchase, err := opt.api.ModelsApi().Purchase().Find(ctx, *opt.purchaseId)
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
	opt.api.ClientMgr().SocketEmit(clnt, "payment:received", data)
}

func (opt *PaymentOption) UseWalletBal(w http.ResponseWriter, r *http.Request, debit float64) error {
	opt.mu.Lock()
	defer opt.mu.Unlock()

	ctx := r.Context()
	tx, err := opt.api.DbApi().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	p, err := opt.api.ModelsApi().Purchase().FindTx(tx, ctx, *opt.purchaseId)
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

func (opt *PaymentOption) Done(w http.ResponseWriter, r *http.Request) {
	opt.mu.RLock()
	defer opt.mu.RUnlock()

	p, err := opt.api.ModelsApi().Purchase().Find(r.Context(), *opt.purchaseId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	opt.api.PaymentsApi().ExecCallback(w, r, p)
}

func (opt *PaymentOption) Cancel(w http.ResponseWriter, r *http.Request) {
	opt.mu.RLock()
	defer opt.mu.RUnlock()

	clnt, err := opt.api.ClientReg().CurrentClient(r)
	if err != nil {
		opt.ErrResp(w, err)
		return
	}

	pur, err := opt.api.ModelsApi().Purchase().PendingPurchase(r.Context(), clnt.Id())
	if err != nil {
		opt.ErrResp(w, err)
		return
	}

	err = pur.Cancel(r.Context())
	if err != nil {
		opt.ErrResp(w, err)
		return
	}

	flash.SetFlashMsg(w, flash.Info, "Purchase was cancelled.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
