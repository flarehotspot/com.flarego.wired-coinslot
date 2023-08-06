package models

import (
	"context"
	"database/sql"

	"github.com/flarehotspot/sdk/v1.0.0/api"
)

type WiredCoinslotModel struct {
	api     api.IPluginApi
	allStmt *sql.Stmt
}

func (self *WiredCoinslotModel) Create(
	ctx context.Context,
	alias *string,
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
	db := self.api.DbApi()
	result, err := db.ExecContext(ctx, `
  INSERT INTO wired_coinslots (
    alias, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
    bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime
  ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `,
		alias, coinPin, coinInhibitPin, coinRelayActive, coinRelayDelaySec, coinBouncetime,
		billPin, billInhibitPin, billRelayActive, billRelayDelaySec, billBouncetime,
	)
	if err != nil {
		return nil, err
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var c WiredCoinslot
	err = db.QueryRowContext(ctx, `
  SELECT id, alias, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
    bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime, created_at
  FROM wired_coinslots
  WHERE id = ?
  LIMIT 1
  `, lastId).Scan(
		&c.id, &c.alias, &c.coinPin, &c.coinInhibitPin, &c.coinRelayActive, &c.coinRelayDelaySec, &c.coinBouncetime,
		&c.billPin, &c.billInhibitPin, &c.billRelayActive, &c.billRelayDelaySec, &c.billBouncetime, &c.createdAt,
	)

	return &c, nil
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
		var c WiredCoinslot
		err := rows.Scan(
			&c.id, &c.alias, &c.coinPin, &c.coinInhibitPin, &c.coinRelayActive, &c.coinRelayDelaySec, &c.coinBouncetime,
			&c.billPin, &c.billInhibitPin, &c.billRelayActive, &c.billRelayDelaySec, &c.billBouncetime, &c.createdAt,
		)
		if err != nil {
			return nil, err
		}
		coinslots = append(coinslots, &c)
	}

	return coinslots, nil
}

func NewWiredCoinslotModel(api api.IPluginApi) (*WiredCoinslotModel, error) {
	allStmt, err := api.DbApi().Prepare(`
    SELECT id, alias, coin_pin, coin_inhibit_pin, coin_relay_active, coin_relay_delay_sec, coin_bouncetime,
      bill_pin, bill_inhibit_pin, bill_relay_active, bill_relay_delay_sec, bill_bouncetime, created_at
    FROM wired_coinslots
    `)

	if err != nil {
		return nil, err
	}

	return &WiredCoinslotModel{api, allStmt}, nil
}
