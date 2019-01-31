/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package config

import (
	"time"
	"fmt"
)

type ConfigFactory struct {
	cfg *Config
}

func (cf *ConfigFactory) CheckConfig() error {
	//  check every config to ensure GetXXXConfig method is safe without error return
	monitorSwitch := cf.cfg.MonitorSwitch
	if !monitorSwitch {
		return fmt.Errorf("MonitorSwitch is %t,please set MonitorSwitch is true!", monitorSwitch)
	}

	monitorTime := cf.cfg.Time.Seconds()
	intervalTime := cf.cfg.IntervalTime.Seconds()
	if monitorTime <= 0 || intervalTime <= 0 {
		return fmt.Errorf("monitorTime is %d , intervalTime is %d ,The value must greater than zero!", int(monitorTime), int(intervalTime))
	}

	if cf.cfg.Tls.TlsSwitch && len(cf.cfg.Tls.ClientCertPath) != 3 {
		return fmt.Errorf("TlsConfig Error! TlsSwitch is %t,CertPath is %v", cf.cfg.Tls.TlsSwitch, cf.cfg.Tls.ClientCertPath)
	}
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
