package config

import (
	"github.com/flarehotspot/sdk/api/config"
	"github.com/flarehotspot/sdk/libs/yaml-3"
)

type WiredCoinslotDef struct {
	Enable              bool   `yaml:"enable"`
	Alias               string `yaml:"alias"`
	CoinAcceptorPin     int    `yaml:"coin_acceptor_pin"`
	BillAcceptorPin     int    `yaml:"bill_acceptor_pin"`
	CoinRelayPin        int    `yaml:"coin_relay_pin"`
	BillRelayPin        int    `yaml:"bill_relay_pin"`
	CoinRelayActiveHigh bool   `yaml:"coin_relay_active_high"`
	BillRelayActiveHigh bool   `yaml:"bill_relay_active_high"`
	RelayDelaySec       int    `yaml:"relay_delay_sec"`
}

type WiredCoinslotsCfg struct {
	Coinslots map[string]*WiredCoinslotDef `yaml:"coinslots"`
}

type Config struct {
	configAPI config.IConfigApi
}

func (c *Config) Read() (*WiredCoinslotsCfg, error) {
	cfgBytes, err := c.configAPI.Read()
	if err != nil {
		return nil, err
	}
	var cfg WiredCoinslotsCfg
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save(cfg *WiredCoinslotsCfg) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return c.configAPI.Write(b)
}

func NewConfig(cfgApi config.IConfigApi) *Config {
	return &Config{cfgApi}
}
