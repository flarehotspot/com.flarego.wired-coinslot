package app

import sdkapi "sdk/api"

func NewPaymentOption(opt sdkapi.PaymentOption) *PaymentOption {
	return &PaymentOption{
		opt: opt,
	}
}

type PaymentOption struct {
	opt   sdkapi.PaymentOption
	devId int64
}
