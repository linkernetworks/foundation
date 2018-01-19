package config

import (
	"net"
	"strconv"
)

type KudisConfig struct {
	Host   string       `json:"host"`
	Port   int32        `json:"port"`
	Logger LoggerConfig `json:"logger"`
}

func (c *KudisConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 50051
	}
	return nil
}

func (c *KudisConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
