package main

import (
	"Lancet/config"
	"Lancet/logging"
	"Lancet/monitor"
	"time"
)

var logger = lancetlogging.GetLogger()

func main() {
	cfy, err := config.LoadFromFile("./config/config.yaml")
	if err != nil {
		logger.Errorf("LoadFromFile Error: %s", err)
	}
	cf := cfy.GetAllConfig()
	monitor.NewMail(cf.Mail.MailUser, cf.Mail.MailPasswd, cf.Mail.SmtpHost, cf.Mail.ReceiveMail)
	mcs := make([]*monitor.MonitorCli, 0)
	for hostname, host := range cf.Hosts {
		mc, err := monitor.NewMonitorCliFromConf(hostname, host.Address, host.ApiVersion, cf.IntervalTime)
		if err != nil {
			logger.Errorf("NewMonitorCliFromConf Error: %s", err)
			panic(err)
		}
		mcs = append(mcs, mc)
	}
	monitorSwitch := monitor.NewMonitorSwitch(mcs)
	monitor.FinishMonitor = make(chan bool)

	monitorSwitch.StartMonitor()
	logger.Debugf("MonitorTime  is  %s", cf.Time)
	time.Sleep(cf.Time)
	monitorSwitch.StopMonitor()
}
