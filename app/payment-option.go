package app

import (
	sdkpayments "github.com/flarehotspot/core/sdk/api/payments"
)

func NewPaymentOption(opt sdkpayments.PaymentOpt) *PaymentOption {
	return &PaymentOption{
		opt: opt,
	}
}

type PaymentOption struct {
	opt   sdkpayments.PaymentOpt
	devId int64
}
