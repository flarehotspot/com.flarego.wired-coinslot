package models

import (
	"context"
	"database/sql"
	plugin "github.com/flarehotspot/sdk/api/plugin"
)

func NewWiredCoinslotModel(api plugin.PluginApi) (*WiredCoinslotModel, error) {
	allStmt, err := api.SqlDb().Prepare(`
    SELECT id, alias, curr_device_id, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
      bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime, created_at
    FROM wired_coinslots
    `)

	if err != nil {
		return nil, err
	}

	return &WiredCoinslotModel{api, allStmt}, nil
}

type WiredCoinslotModel struct {
	api     plugin.PluginApi
	allStmt *sql.Stmt
}

func (self *WiredCoinslotModel) Create(
	ctx context.Context,
	alias *string,
	currDeviceId *int64,
	coinPin uint,
	coinInhibitPin uint,
	coinRelayActive bool,
	coinRelayDelaySec uint,
	coinBouncetime uint,
	billPin *uint,
	billInhibitPin uint,
	billRelayActive bool,
	billRelayDelaySec uint,
	billBouncetime uint,
) (*WiredCoinslot, error) {
	db := self.api.SqlDb()
	result, err := db.ExecContext(ctx, `
  INSERT INTO wired_coinslots (
    alias, curr_device_id, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
    bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime
  ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `,
		alias, currDeviceId, coinPin, coinInhibitPin, coinRelayActive, coinRelayDelaySec, coinBouncetime,
		billPin, billInhibitPin, billRelayActive, billRelayDelaySec, billBouncetime,
	)
	if err != nil {
		return nil, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return self.Find(lastId)
}

func (self *WiredCoinslotModel) All() ([]*WiredCoinslot, error) {
	stmt := self.allStmt
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	coinslots := []*WiredCoinslot{}
	for rows.Next() {
		c := NewWiredCoinslot(self)
		err := rows.Scan(
			&c.id, &c.alias, &c.curr_device_id, &c.coinPin, &c.coinInhibitPin, &c.coinRelayActive, &c.coinRelayDelaySec, &c.coinBouncetime,
			&c.billPin, &c.billInhibitPin, &c.billRelayActive, &c.billRelayDelaySec, &c.billBouncetime, &c.createdAt,
		)
		if err != nil {
			return nil, err
		}
		coinslots = append(coinslots, c)
	}

	return coinslots, nil
}

func (self *WiredCoinslotModel) Find(id int64) (*WiredCoinslot, error) {
	c := NewWiredCoinslot(self)
	err := self.api.SqlDb().QueryRow(`
      SELECT id, alias, curr_device_id, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
    bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime, created_at
      FROM wired_coinslots
      WHERE id = ?
      LIMIT 1
      `, id).Scan(
		&c.id, &c.alias, &c.curr_device_id, &c.coinPin, &c.coinInhibitPin, &c.coinRelayActive, &c.coinRelayDelaySec, &c.coinBouncetime,
		&c.billPin, &c.billInhibitPin, &c.billRelayActive, &c.billRelayDelaySec, &c.billBouncetime, &c.createdAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Update coinslot
func (self *WiredCoinslotModel) Update(
	ctx context.Context,
	id int64,
	alias *string,
	currDeviceId *int64,
	coinPin uint,
	coinInhibitPin uint,
	coinRelayActive bool,
	coinRelayDelaySec uint,
	coinBouncetime uint,
	billPin *uint,
	billInhibitPin uint,
	billRelayActive bool,
	billRelayDelaySec uint,
	billBouncetime uint,
) (*WiredCoinslot, error) {
	db := self.api.SqlDb()
	_, err := db.ExecContext(ctx, `
  UPDATE wired_coinslots SET
    alias = ?, curr_device_id = ?, coin_pin = ?, coin_inhibit_pin = ?, coin_relay_active = ?, coin_relay_delay_sec = ?, coin_bouncetime = ?,
    bill_pin = ?, bill_inhibit_pin = ?, bill_relay_active = ?, bill_relay_delay_sec = ?, bill_bouncetime = ?
  WHERE id = ?
  `,
		alias, currDeviceId, coinPin, coinInhibitPin, coinRelayActive, coinRelayDelaySec, coinBouncetime,
		billPin, billInhibitPin, billRelayActive, billRelayDelaySec, billBouncetime,
		id,
	)
	if err != nil {
		return nil, err
	}

	return self.Find(id)
}
