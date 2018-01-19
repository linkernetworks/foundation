package config

type PingConfig struct {
	Interval int `json:"interval"`
	Timeout  int `json:"timeout"`
}

type SocketioConfig struct {
	MaxConnection int        `json:"maxConnection"`
	Ping          PingConfig `json:"ping"`
}
