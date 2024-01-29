package models

import (
	"fmt"
	"time"
)

type WiredCoinslot struct {
	id                int64
	alias             *string
	curr_device_id    *int64
	coinPin           uint
	coinInhibitPin    uint
	coinRelayActive   bool
	coinRelayDelaySec uint
	coinBouncetime    uint

	billPin           *uint
	billInhibitPin    uint
	billRelayActive   bool
	billRelayDelaySec uint
	billBouncetime    uint

	createdAt time.Time
}

func (self *WiredCoinslot) Id() int64 {
	return self.id
}

func (self *WiredCoinslot) Name() string {
	if self.alias != nil {
		return *self.alias
	}
	return fmt.Sprintf("Coinslot Pin # %d", self.coinPin)
}
