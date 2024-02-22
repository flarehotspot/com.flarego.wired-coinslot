package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	connmgr "github.com/flarehotspot/sdk/api/connmgr"
	payments "github.com/flarehotspot/sdk/api/payments"
	plugin "github.com/flarehotspot/sdk/api/plugin"
)

func NewPaymentProvider(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	provider := &PaymentProvider{
		name:  "Wired Coinslots",
		api:   api,
		model: mdl,
	}

	api.Payments().NewPaymentProvider(provider)
}

type PaymentProvider struct {
	name  string
	api   plugin.PluginApi
	model *models.WiredCoinslotModel
}

func (self *PaymentProvider) Name() string {
	return self.name
}

func (self *PaymentProvider) GetOpts() []payments.PaymentOpt {
	opts := []payments.PaymentOpt{}
	coinslots, err := self.model.All()
	if err != nil {
		log.Println(err)
		return nil
	}

	for _, c := range coinslots {
		id := strconv.Itoa(int(c.Id()))
		opt := payments.PaymentOpt{
			OptName:      c.Name(),
			VueRouteName: "insert-coin",
			RouteParams:  map[string]string{"id": id},
		}
		opts = append(opts, opt)
	}

	return opts
}

func (self *PaymentProvider) PaymentOpts(clnt connmgr.ClientDevice) []payments.PaymentOpt {
	return self.GetOpts()
}

func (self *PaymentProvider) FindOpt(clnt connmgr.ClientDevice) (opt payments.PaymentOpt, ok bool) {
	// for _, opt := range self.GetOpts() {
	// if opt.devId == clnt.Id() {
	// 	return opt.opt, true
	// }
	// }
	return payments.PaymentOpt{}, false
}

func (self *PaymentProvider) PaymentReceived(w http.ResponseWriter, r *http.Request) {
	// f := r.URL.Query().Get("amount")
	// amount, err := strconv.ParseFloat(f, 32)
	// if err != nil {
	// 	log.Println(err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// clntSym := r.Context().Value(contexts.ClientCtxKey)
	// clnt, ok := clntSym.(connmgr.IClientDevice)
	// if !ok {
	// 	log.Println("Could not determine client device.")
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// opt, ok := self.FindOpt(clnt)
	// if !ok {
	// 	errmsg := "Cannot determine pending purchase for client: " + clnt.IpAddr()
	// 	http.Error(w, errmsg, http.StatusInternalServerError)
	// 	return
	// }

	// opt.PaymentReceived(r.Context(), clnt, amount)
	// log.Printf("Payment received: %f", amount)
	// w.WriteHeader(http.StatusOK)
}

func (self *PaymentProvider) UseWalletBal(w http.ResponseWriter, r *http.Request) {
	// f := r.URL.Query().Get("amount")
	// amount, err := strconv.ParseFloat(f, 32)
	// if err != nil {
	// 	log.Println(err)
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// clntSym := r.Context().Value(contexts.ClientCtxKey)
	// clnt, ok := clntSym.(connmgr.IClientDevice)
	// if !ok {
	// 	log.Println("Could not determine client device.")
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// opt, ok := self.FindOpt(clnt)
	// if !ok {
	// 	errmsg := "Cannot determine pending purchase for client: " + clnt.IpAddr()
	// 	http.Error(w, errmsg, http.StatusInternalServerError)
	// 	return
	// }

	// err = opt.UseWalletBal(w, r, amount)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }
}

func (self *PaymentProvider) Done(w http.ResponseWriter, r *http.Request) {
	// clnt, err := self.api.ClientReg().CurrentClient(r)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// opt, ok := self.FindOpt(clnt)
	// if !ok {
	// 	errmsg := "Cannot determine pending purchase for client: " + clnt.IpAddr()
	// 	http.Error(w, errmsg, http.StatusInternalServerError)
	// 	return
	// }

	// opt.Done(w, r)
}

func (self *PaymentProvider) Cancel(w http.ResponseWriter, r *http.Request) {
	// clnt, err := self.api.ClientReg().CurrentClient(r)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// opt, ok := self.FindOpt(clnt)
	// if !ok {
	// 	errmsg := "Cannot determine pending purchase for client: " + clnt.IpAddr()
	// 	http.Error(w, errmsg, http.StatusInternalServerError)
	// 	return
	// }

	// opt.Cancel(w, r)
}
