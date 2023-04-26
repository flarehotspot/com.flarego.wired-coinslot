package coinslots

import (
	"sync"

	"github.com/flarehotspot/sdk/api/plugin"
)

type CoinslotsMgr struct {
	mu        sync.RWMutex
	api       plugin.IPluginApi
	coinslots []*WiredCoinslot
}

func (mgr *CoinslotsMgr) Register() {

}

func (mgr *CoinslotsMgr) All() []*WiredCoinslot {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return mgr.coinslots
}
