//go:build !mono

package main

import (
	"com.flarego.wired-coinslot/app"
	plugin "sdk/api/plugin"
)

func main() {}

func Init(api plugin.PluginApi) {
	if err := api.Migrate(); err != nil {
		api.Logger().Error(err.Error())
		return
	}

	app.Init(api)
}
