//go:build !dev
package main

import "github.com/flarehotspot/sdk/interfaces/plugins"

func Init(g plugins.PluginGlobals) *plugins.PluginParams {
	return plugins.NewPluginParams(g)
}

// func Start() {
// log.Println("Hello from payment.Init()...")
// }
