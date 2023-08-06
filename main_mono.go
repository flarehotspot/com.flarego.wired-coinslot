//go:build mono

package wiredcoinslot

import (
	"log"

	"github.com/flarehotspot/sdk"
	"github.com/flarehotspot/sdk/v1.0.0/api"
)

func Init(_sdk sdk.SDK) {
	sym, err := _sdk.GetVersion(api.VERSION)
	if err != nil {
		log.Println("Unable to get plugin api: ", err)
	}

	apiv1 := sym.(api.IPluginApi)
	log.Printf("Success loading plugin: %s", apiv1.Name())
}
