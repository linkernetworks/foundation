package config

import (
	"net"
	"strconv"
)

type RedisConfig struct {
	Host      string       `json:"host"`
	Port      int          `json:"port"`
	Interface string       `json:"interface"`
	Public    *RedisConfig `json:"public"`
}

func (c *RedisConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *RedisConfig) SetHost(host string) {
	c.Host = host
}

func (c *RedisConfig) SetPort(port int) {
	c.Port = port
}

func (c *RedisConfig) LoadDefaults() {
	if c.Port == 0 {
		c.Port = 6379
	}
}

func (c *RedisConfig) GetInterface() string {
	return c.Interface
}

func (c *RedisConfig) GetPublic() ServiceConfig {
	return c.Public
}

func (c *RedisConfig) URL() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
