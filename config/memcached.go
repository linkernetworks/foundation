package config

import (
	"github.com/linkernetworks/foundation/service/serviceconfig"
	"net"
	"strconv"
)

type MemcachedConfig struct {
	Host      string           `json:"host"`
	Port      int32            `json:"port"`
	Interface string           `json:"interface"`
	Public    *MemcachedConfig `json:"public"`
}

func (c *MemcachedConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *MemcachedConfig) SetHost(host string) {
	c.Host = host
}

func (c *MemcachedConfig) SetPort(port int32) {
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

func (c *MemcachedConfig) GetPublic() serviceconfig.ServiceConfig {
	return c.Public
}

func (c *MemcachedConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
