/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package config

import "time"

type ConfigFactory struct {
	cfg *Config
}

func (cf *ConfigFactory) CheckConfig() error {
	// TODO: check every config to ensure GetXXXConfig method is safe without error return
	return nil
}

func (cf *ConfigFactory) GetMonitorHostsConfig() map[string]MonitorHosts {
	return cf.cfg.Hosts
}

func (cf *ConfigFactory) GetMonitorTimeConfig() time.Duration {
	return cf.cfg.Time
}

func (cf *ConfigFactory) GetAllConfig() Config {
	return *cf.cfg
}
