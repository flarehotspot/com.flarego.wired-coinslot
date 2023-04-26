package models

import (
	"fmt"
	"time"
)

type WiredCoinslot struct {
	id                int64
	alias             *string
	coinPin           uint
	coinInhibitPin    uint
	coinRelayActive   bool
	coinRelayDelaySec uint
	coinBouncetime    uint

	billPin           *uint
	billInhibitPin    *uint
	billRelayActive   bool
	billRelayDelaySec uint
	billBouncetime    uint

	createdAt time.Time
}

func (self *WiredCoinslot) Name() string {
	if self.alias != nil {
		return *self.alias
	}
	return fmt.Sprintf("Coinslot Pin # %d", self.coinPin)
}
