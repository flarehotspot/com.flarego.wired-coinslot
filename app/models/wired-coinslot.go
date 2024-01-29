package models

import (
	"context"
	"fmt"
	"time"
)

func NewWiredCoinslot(mdl *WiredCoinslotModel) *WiredCoinslot {
	return &WiredCoinslot{mdl: mdl}
}

type WiredCoinslot struct {
	mdl               *WiredCoinslotModel
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

// CurrentDeviceId int64
func (self *WiredCoinslot) CurrentDeviceId() int64 {
	if self.curr_device_id != nil {
		return *self.curr_device_id
	}
	return -1
}

func (self *WiredCoinslot) SetCurrentDeviceId(id int64) {
    self.curr_device_id = &id
}

// Update coinslot
func (self *WiredCoinslot) UpdateTx(ctx context.Context) error {
	_, err := self.mdl.Update(ctx, self.id, self.alias, self.curr_device_id, self.coinPin, self.coinInhibitPin, self.coinRelayActive, self.coinRelayDelaySec, self.coinBouncetime, self.billPin, self.billInhibitPin, self.billRelayActive, self.billRelayDelaySec, self.billBouncetime)
	return err
}
