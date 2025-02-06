package src

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	sdkapi "sdk/api"

	sdkutils "github.com/flarehotspot/sdk-utils"
	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	WiredCoinslotsPrefix string = "wired_coinslots"
)

func InitWiredCoinslots(api sdkapi.IPluginApi) {
	_, err := api.Config().Plugin().List(WiredCoinslotsPrefix)
	if errors.Is(err, os.ErrNotExist) {
		mainCoinslot := NewWiredCoinslot(api, "Main Vendo")
		if err := mainCoinslot.Save(); err != nil {
			api.Logger().Error(err.Error())
		}
	}
}

func NewWiredCoinslot(api sdkapi.IPluginApi, name string) *WiredCoinslot {
	return &WiredCoinslot{
		api:  api,
		Id:   sdkutils.RandomStr(16),
		Name: name,
	}
}

func GetAllWiredCoinslots(api sdkapi.IPluginApi) ([]*WiredCoinslot, error) {
	coinslotEntries, err := api.Config().Plugin().List(WiredCoinslotsPrefix)
	if err != nil {
		return nil, err
	}

	coinslots := make([]*WiredCoinslot, len(coinslotEntries))
	for i, entry := range coinslotEntries {
		b, err := api.Config().Plugin().Read(entry.Path)
		if err != nil {
			fmt.Println("Error reading wired coinslot config:", err)
			continue
		}

		var c WiredCoinslot
		if err := json.Unmarshal(b, &c); err != nil {
			fmt.Println("Error parsing wired coinslot: ", err)
			continue
		}

		c.api = api
		coinslots[i] = &c
	}

	return coinslots, nil
}

func FindWiredCoinslot(api sdkapi.IPluginApi, id string) (*WiredCoinslot, error) {
	var c WiredCoinslot
	b, err := api.Config().Plugin().Read(filepath.Join(WiredCoinslotsPrefix, id))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(b), &c); err != nil {
		return nil, err
	}

	c.api = api
	return &c, nil
}

type WiredCoinslot struct {
	api      sdkapi.IPluginApi
	Id       string
	Name     string
	DeviceID *pgtype.UUID
}

func (c *WiredCoinslot) ConfigPath() string {
	return filepath.Join(WiredCoinslotsPrefix, c.Id)
}

func (c *WiredCoinslot) Save() error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return c.api.Config().Plugin().Write(c.ConfigPath(), b)
}
