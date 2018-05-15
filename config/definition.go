/*
Copyright Yunphant Corp. All Rights Reserved.
*/

package config

import "time"

type Config struct {
	Hosts        map[string]MonitorHosts
	IntervalTime time.Duration
	Time         time.Duration
}

type MonitorHosts struct {
	Address    string
	ApiVersion string
}
