package src

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	sdkapi "sdk/api"
	"sync"

	sdkutils "github.com/flarehotspot/sdk-utils"
	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	WiredCoinslotsPrefix string = "wired_coinslots"
)

var (
	UsedCoinslots sync.Map
)

func InitWiredCoinslots(api sdkapi.IPluginApi) {
	_, err := api.Config().Plugin().List(WiredCoinslotsPrefix)
	fmt.Println("InitWiredCoinslots Error: ", err)
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
		ID:   sdkutils.RandomStr(16),
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

func FindUsedCoinslot(api sdkapi.IPluginApi, deviceID pgtype.UUID) (*WiredCoinslot, error) {
	deviceIDStr := sdkutils.PgUuidToString(deviceID)
	var coinslotID string
	UsedCoinslots.Range(func(key, value any) bool {
		if value.(string) == deviceIDStr {
			coinslotID = key.(string)
			return false
		}
		return true
	})

	if coinslotID == "" {
		return nil, nil
	}

	return LoadWiredCoinslot(api, coinslotID)
}

func LoadWiredCoinslot(api sdkapi.IPluginApi, coinslotID string) (*WiredCoinslot, error) {
	var c WiredCoinslot
	b, err := api.Config().Plugin().Read(filepath.Join(WiredCoinslotsPrefix, coinslotID))
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
	api  sdkapi.IPluginApi
	ID   string
	Name string
}

func (c *WiredCoinslot) ConfigPath() string {
	return filepath.Join(WiredCoinslotsPrefix, c.ID)
}

func (c *WiredCoinslot) GetID() string {
	return c.ID
}

func (c *WiredCoinslot) GetName() string {
	return c.Name
}

func (c *WiredCoinslot) CanBeUsedBy(deviceID pgtype.UUID) bool {
	deviceIDStr := sdkutils.PgUuidToString(deviceID)
	if v, ok := UsedCoinslots.Load(c.ID); ok {
		if v.(string) == deviceIDStr {
			return true
		}
		return false
	}
	return true
}

func (c *WiredCoinslot) UseBy(deviceID pgtype.UUID) {
	deviceIDStr := sdkutils.PgUuidToString(deviceID)
	UsedCoinslots.Store(c.ID, deviceIDStr)
}

func (c *WiredCoinslot) DoneUsing() {
	UsedCoinslots.Delete(c.ID)
}

func (c *WiredCoinslot) Save() error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return c.api.Config().Plugin().Write(c.ConfigPath(), b)
}
