/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package config

import "time"

type Config struct {
	Hosts         map[string]MonitorHosts
	IntervalTime  time.Duration
	Time          time.Duration
	MonitorSwitch bool
	Tls           Tls
	Mail          MailNotice
}

type MonitorHosts struct {
	Address    string
	ApiVersion string
}

type MailNotice struct {
	MailUser    string
	MailPasswd  string
	SmtpHost    string
	ReceiveMail []string
}

type Tls struct {
	TlsSwitch      bool
	ClientCertPath []string
}
