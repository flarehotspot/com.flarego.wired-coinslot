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
		id:   sdkutils.RandomStr(16),
		name: name,
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

func FindWiredCoinslotByDevice(api sdkapi.IPluginApi, deviceID pgtype.UUID) (*WiredCoinslot, error) {
	coinslots, err := GetAllWiredCoinslots(api)
	if err != nil {
		return nil, err
	}

	idstr := sdkutils.PgUuidToString(deviceID)
	fmt.Println("FindWiredCoinslotByDevice idstr: ", idstr)

	for _, c := range coinslots {
		if c.deviceID != nil && *c.deviceID == idstr {
			return c, nil
		}
	}

	return nil, fmt.Errorf("No coinslot found for device ID: %s", sdkutils.PgUuidToString(deviceID))
}

type WiredCoinslot struct {
	mu       sync.RWMutex
	api      sdkapi.IPluginApi
	id       string
	name     string
	deviceID *string
}

func (c *WiredCoinslot) ConfigPath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return filepath.Join(WiredCoinslotsPrefix, c.id)
}

func (c *WiredCoinslot) Id() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

func (c *WiredCoinslot) Name() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.name
}

func (c *WiredCoinslot) DeviceId() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.deviceID == nil {
		return ""
	}
	return *c.deviceID
}

func (c *WiredCoinslot) SetDeviceID(id *string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deviceID = id
}

func (c *WiredCoinslot) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return c.api.Config().Plugin().Write(c.ConfigPath(), b)
}
