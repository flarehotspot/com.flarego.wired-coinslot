package payment

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/flarehotspot/sdk/api/devices"
	"github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/plugin"
	mdls "github.com/flarehotspot/wired-coinslot/app/models"
)

type pmtEvt struct {
	Amount float64 `json:"amount"`
}

type PaymentOption struct {
	mu       sync.RWMutex
	api      plugin.IPluginApi
	provider *PaymentProvider
	coinslot *mdls.WiredCoinslot
	paying   devices.IClientDevice
	purchase models.IPurchase
}

func (self *PaymentOption) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, clnt devices.IClientDevice, prch models.IPurchase) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.paying != nil && self.paying.Device().Id() != clnt.Device().Id() {
		err := prch.Cancel(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		self.api.HttpApi().Respond().SetFlashMsg(w, "error", "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	self.paying = clnt
	self.purchase = prch
	self.api.HttpApi().Respond().PortalView(w, r, "insert-coin.html", nil)
}

func (self *PaymentOption) PaymentReceived(ctx context.Context, amount float64) {
	tx, err := self.api.Db().BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer tx.Rollback()

	payment, err := self.purchase.PaymentTx(tx, ctx)
	if err != nil {
		log.Println(err)
		return
	}

	newAmount := payment.Amount() + amount
	err = payment.UpdateTx(tx, ctx, newAmount, nil, nil)
	if err != nil {
		log.Println(err)
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	data := pmtEvt{Amount: amount}
	self.paying.Emit("payment:received", data)
}

func NewPaymentOpt(api plugin.IPluginApi, prvdr *PaymentProvider, c *mdls.WiredCoinslot) *PaymentOption {
	return &PaymentOption{
		api:      api,
		provider: prvdr,
		coinslot: c,
		paying:   nil,
	}
}
