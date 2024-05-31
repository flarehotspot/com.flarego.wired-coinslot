//go:build !mono

package main

import (
	"com.flarego.wired-coinslot/app"
	plugin "sdk/api/plugin"
)

func main() {}

func Init(api plugin.PluginApi) {
	app.Init(api)
}
