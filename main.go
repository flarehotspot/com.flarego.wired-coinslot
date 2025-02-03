//go:build !mono

package main

import (
	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/app"
)

func main() {}

func Init(api sdkapi.IPluginApi) {
	app.Init(api)
}
