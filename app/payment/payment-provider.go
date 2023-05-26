package payment

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/flarehotspot/sdk/api/connmgr"
	"github.com/flarehotspot/sdk/api/http/contexts"
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
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

func (self *PaymentProvider) FindOpt(clnt connmgr.IClientDevice) (opt *PaymentOption, ok bool) {
	self.mu.RLock()
	defer self.mu.RUnlock()

	for _, opt := range self.options {
		if opt.client.Device().Id() == clnt.Device().Id() {
			return opt, true
		}
	}

	return nil, false
}

func (self *PaymentProvider) PaymentReceived(w http.ResponseWriter, r *http.Request) {
	f := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(f, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clntSym := r.Context().Value(contexts.ClientCtxKey)
	clnt, ok := clntSym.(connmgr.IClientDevice)
	if !ok {
		log.Println("Could not determine client device.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opt, ok := self.FindOpt(clnt)
	if !ok {
		errmsg := "Cannot determine pending purchase for client: " + clnt.Device().IpAddress()
		http.Error(w, errmsg, http.StatusInternalServerError)
		return
	}

	opt.PaymentReceived(r.Context(), amount)
	log.Printf("Payment received: %f", amount)
	w.WriteHeader(http.StatusOK)
}

func (self *PaymentProvider) UseWalletBal(w http.ResponseWriter, r *http.Request) {
	f := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(f, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	clntSym := r.Context().Value(contexts.ClientCtxKey)
	clnt, ok := clntSym.(connmgr.IClientDevice)
	if !ok {
		log.Println("Could not determine client device.")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	opt, ok := self.FindOpt(clnt)
	if !ok {
		errmsg := "Cannot determine pending purchase for client: " + clnt.Device().IpAddress()
		http.Error(w, errmsg, http.StatusInternalServerError)
		return
	}

	err = opt.UseWalletBal(w, r, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (self *PaymentProvider) Done(w http.ResponseWriter, r *http.Request) {
	clntSym := r.Context().Value(contexts.ClientCtxKey)
	clnt, ok := clntSym.(connmgr.IClientDevice)
	if !ok {
		errmsg := "Cannot determine client device"
		http.Error(w, errmsg, http.StatusInternalServerError)
		return
	}

	opt, ok := self.FindOpt(clnt)
	if !ok {
		errmsg := "Cannot determine pending purchase for client: " + clnt.Device().IpAddress()
		http.Error(w, errmsg, http.StatusInternalServerError)
		return
	}

	opt.Done(w, r)
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
