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
	client   devices.IClientDevice
	purchase models.IPurchase
}

func (self *PaymentOption) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, client devices.IClientDevice, purchase models.IPurchase) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.client != nil && self.client.Device().Id() != client.Device().Id() {
		err := purchase.Cancel(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		self.api.HttpApi().Respond().SetFlashMsg(w, "error", "Somebody is still paying.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	self.client = client
	self.purchase = purchase
	self.api.HttpApi().Respond().PortalView(w, r, "insert-coin.html", nil)
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

func (self *PaymentOption) UseWalletBal(ctx context.Context, bal float64) error {
	payment, err := self.purchase.Payment(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	err = payment.Update(ctx, payment.Amount(), &bal, payment.WalletTxId())
	if err != nil {
		log.Println(err)
		return err
	}

	data := map[string]any{"amount": bal}
	self.client.Emit("payment:received", data)
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
