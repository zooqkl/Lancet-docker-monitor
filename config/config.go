package config

import (
	"Lancet/logging"
	"Lancet/monitor"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"strings"
)

const (
	envRootPrefix = "FABRIC_LANCET"
)

var (
	configFactory *ConfigFactory
)
var logger = lancetlogging.GetLogger()

// Load config from path
func LoadFromFile(name string) (*ConfigFactory, error) {
	if name == "" {
		return nil, errors.New(fmt.Sprintf("invalid config filename : %s", name))
	}
	configFactory = &ConfigFactory{cfg: &Config{
		Hosts:        make(map[string]MonitorHosts),
		IntervalTime: 0,
		Time:         0,
	}}

	v := newViper(envRootPrefix)
	v.SetConfigFile(name)

	err := reloadConfigFromFile(v, configFactory)
	if err != nil {
		panic(err.Error())
	}
	// reload config when config file changed
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		v.ReadInConfig()
		logger.Errorf("ConfigChange ! reloadConfigFromFile!")
		//reloadConfigFromFile(v, configFactory)
		monitorSwitch := v.GetBool("monitorSwitch")
		if !monitorSwitch {
			logger.Debugf("monitorSwitch is %t ! stop monitor!", monitorSwitch)
			ms := monitor.NewMonitorSwitch(nil)
			logger.Debugf("monitorSwitch is %t ! stop monitor!，%v", monitorSwitch, ms)
			ms.StopMonitor()
			panic(fmt.Errorf("monitorSwitch is %t,montor Program forcibly stopped！", monitorSwitch))
		}
	})
	return configFactory, nil
}

func GetConfigFactory() (*ConfigFactory, error) {
	if configFactory == nil {
		return nil, fmt.Errorf("[GetConfigFactory] Config factory has not been initialized yet!")
	}
	return configFactory, nil
}

func newViper(cmdRootPrefix string) *viper.Viper {
	myViper := viper.New()
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	return myViper
}

func reloadConfigFromFile(v *viper.Viper, cf *ConfigFactory) error {
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error loading config file '%s' : %s", v.ConfigFileUsed(), err)
	}
	logger.Debugf("ConfigFactory is %v ;\n  viper is %v", cf, v)
	cf.cfg.Time = v.GetDuration("monitorTime")
	cf.cfg.IntervalTime = v.GetDuration("intervalTime")
	cf.cfg.MonitorSwitch = v.GetBool("monitorSwitch")
	v.UnmarshalKey("monitorHosts", &cf.cfg.Hosts)
	v.UnmarshalKey("mailNotice", &cf.cfg.Mail)
	v.UnmarshalKey("tls", &cf.cfg.Tls)

	err := cf.CheckConfig()
	if err != nil {
		panic(fmt.Errorf("Check config error: %s", err))
	}
	return nil
}
