package payment

import (
	"fmt"
	"net/http"

	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	// uuid "github.com/flarehotspot/sdk/libs/go.uuid"
)

type PaymentMethod struct {
	api plugin.IPluginApi
}

func (self *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (self *PaymentMethod) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	// pur, err := self.api.PaymentsApi().ParsePurchaseRequest(r)
	// if err != nil {
		// fmt.Fprint(w, err.Error(), http.StatusInternalServerError)
		// return
	// }

	// id, err := uuid.NewV4()
	// if err != nil {
		// fmt.Fprint(w, err.Error(), http.StatusInternalServerError)
		// return
	// }

	// info := payments.PaymentInfo{
		// Event: payments.EventStart,
		// // Id:    id,
		// // Items: pur.Items,
		// Amount: *&payments.UnitAmount{
			// CurrencyCode: currencies.CurrencyPhilippinePeso,
			// Value:        11.1,
		// },
	// }
	// self.api.PaymentsApi().EmitEvent(pur.CallbackUrl, &info)
	fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentMethod(api plugin.IPluginApi) payments.IPaymentMethod {
	return &PaymentMethod{api}
}
