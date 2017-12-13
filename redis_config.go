package config

import (
	"encoding/json"
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

// Implement DefaultLoader
func (c *RedisConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 6379
	}
	if c.Host == "" {
		c.Host = "localhost"
	}
	return nil
}

func (c *RedisConfig) GetInterface() string {
	return c.Interface
}

func (c *RedisConfig) GetPublic() ServiceConfig {
	return c.Public
}

func (c *RedisConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c *RedisConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	return c.LoadDefaults()
}
