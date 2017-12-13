package config

import (
	"encoding/json"
	"net"
	"strconv"
)

type MemcachedConfig struct {
	Host      string           `json:"host"`
	Port      int              `json:"port"`
	Interface string           `json:"interface"`
	Public    *MemcachedConfig `json:"public"`
}

func (c *MemcachedConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *MemcachedConfig) SetHost(host string) {
	c.Host = host
}

func (c *MemcachedConfig) SetPort(port int) {
	c.Port = port
}

func (c *MemcachedConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 11211
	}
	return nil
}

func (c *MemcachedConfig) GetInterface() string {
	return c.Interface
}

func (c *MemcachedConfig) GetPublic() ServiceConfig {
	return c.Public
}

func (c *MemcachedConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c *MemcachedConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	return c.LoadDefaults()
}
