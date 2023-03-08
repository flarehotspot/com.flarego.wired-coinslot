package coinslots

import (
	"sync"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/config"
)

type CoinslotsMgr struct {
	mu        sync.RWMutex
	api       plugin.IPluginApi
	cfg       *config.Config
	coinslots []*WiredCoinslot
}

func (mgr *CoinslotsMgr) Register(def *config.WiredCoinslotDef) {

}

func (mgr *CoinslotsMgr) All() []*WiredCoinslot {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return mgr.coinslots
}
