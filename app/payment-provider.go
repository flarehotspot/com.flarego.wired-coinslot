package app

import (
	"net/http"

	sdkapi "sdk/api"
)

func NewPaymentProvider(api sdkapi.IPluginApi) *PaymentProvider {
	return &PaymentProvider{
		name: "Wired Coinslots",
		api:  api,
	}
}

type PaymentProvider struct {
	name string
	api  sdkapi.IPluginApi
}

func (self *PaymentProvider) Name() string {
	return self.name
}

func (self *PaymentProvider) OptionsFactory(r *http.Request) []sdkapi.PaymentOption {
	return []sdkapi.PaymentOption{
		{
			Name:        "Coinslot 1",
			RouteName:   "payments:insert_coin",
			RouteParams: map[string]string{"id": "1"},
		},
	}

	// wiredCoinslots, err := config.FindAll(self.api)
	// if err != nil {
	// 	fmt.Println("Error in finding all coinslots: ", err)
	// 	return []sdkapi.PaymentOption{
	// 		{
	// 			Name:      "Coinslot 1",
	// 			RouteName: "payment:insert_coin",
	// 		},
	// 	}
	// }

	// opts := make([]sdkapi.PaymentOption, len(wiredCoinslots))

	// for _, entry := range wiredCoinslots {
	// 	opt := sdkapi.PaymentOption{
	// 		Name:        entry.Name,
	// 		RouteName:   "payment:insert_coin",
	// 		RouteParams: map[string]string{"id": entry.ID},
	// 	}

	// 	opts = append(opts, opt)
	// }

	// return opts
}

func (self *PaymentProvider) GetPaymentOption(r *http.Request) (opt sdkapi.PaymentOption, ok bool) {
	// for _, opt := range self.GetOpts() {
	// if opt.devId == clnt.Id() {
	// 	return opt.opt, true
	// }
	// }
	return sdkapi.PaymentOption{}, false
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
