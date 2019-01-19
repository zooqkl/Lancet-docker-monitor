/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package config

import "time"

type Config struct {
	Hosts        map[string]MonitorHosts
	IntervalTime time.Duration
	Time         time.Duration
	Mail         MailNotice
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
