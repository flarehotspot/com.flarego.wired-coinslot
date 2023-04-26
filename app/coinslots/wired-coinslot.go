package coinslots

import (
	"log"
	"os"
	"os/exec"

	"github.com/flarehotspot/sdk/api/plugin"
)

type WiredCoinslot struct {
	api  plugin.IPluginApi
	proc *os.Process
}

func (wc *WiredCoinslot) Init() error {
	cmd := exec.Command("echo", "Init coinslot: ")
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return err
	}
	wc.proc = cmd.Process
	return nil
}

func (wc *WiredCoinslot) Stop() error {
	err := wc.proc.Kill()
	if err != nil {
		return err
	}
	return nil
}
