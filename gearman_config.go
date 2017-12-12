package config

import (
	"net"
	"strconv"
)

type GearmanConfig struct {
	Host      string         `json:"host"`
	Port      int            `json:"port"`
	Interface string         `json:"interface"`
	Public    *GearmanConfig `json:"public"`
}

func (c *GearmanConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *GearmanConfig) SetHost(host string) {
	c.Host = host
}

func (c *GearmanConfig) SetPort(port int) {
	c.Port = port
}

func (c *GearmanConfig) LoadDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 4730
	}
}

func (c *GearmanConfig) GetInterface() string {
	return c.Interface
}

func (c *GearmanConfig) GetPublic() ServiceConfig {
	return c.Public
}

func (c *GearmanConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
