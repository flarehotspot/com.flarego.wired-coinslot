package payment

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
	"github.com/gorilla/mux"
)

type PaymentProvider struct {
	mu      sync.RWMutex
	name    string
	api     plugin.IPluginApi
	model   *models.WiredCoinslotModel
	options []*PaymentOption
}

func (self *PaymentProvider) Name() string {
	return self.name
}

func (self *PaymentProvider) PaymentOpts() []payments.IPaymentOpt {
	opts := []payments.IPaymentOpt{}
	for _, opt := range self.options {
		opts = append(opts, opt)
	}
	return opts
}

func (self *PaymentProvider) AddPaymentOpt(opt *PaymentOption) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.options = append(self.options, opt)
}

func (self *PaymentProvider) LoadOpts() {
	coinslots, err := self.model.All()
	if err != nil {
		log.Println(err)
		return
	}
	for _, c := range coinslots {
		opt := NewPaymentOpt(self.api, self, c)
		self.AddPaymentOpt(opt)
	}
}

func (self *PaymentProvider) FindOpt(name string) (opt *PaymentOption, ok bool) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	for _, opt := range self.options {
		if opt.Name() == name {
			return opt, true
		}
	}
	return nil, false
}

func (self *PaymentProvider) PaymentReceived(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]
	f := params["amount"]
	amount, err := strconv.ParseFloat(f, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opt, ok := self.FindOpt(name)
	if !ok {
		http.Error(w, "Invalid payment option: "+name, http.StatusInternalServerError)
		return
	}

	opt.PaymentReceived(context.Background(), amount)
  log.Printf("Payment received: %f", amount)
	w.WriteHeader(http.StatusOK)
}

func NewPaymentProvider(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) *PaymentProvider {
	provider := PaymentProvider{
		name:    "Wired Coinslots",
		api:     api,
		model:   mdl,
		options: []*PaymentOption{},
	}

	provider.LoadOpts()

	return &provider
}
