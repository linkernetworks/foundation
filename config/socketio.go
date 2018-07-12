package config

import "time"

type PingConfig struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
}

type SocketioConfig struct {
	MaxConnection int        `json:"maxConnection"`
	Ping          PingConfig `json:"ping"`
}
