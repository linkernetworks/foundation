package gearman

import (
	"github.com/linkernetworks/foundation/service/serviceconfig"
	"net"
	"strconv"
)

type GearmanConfig struct {
	Host      string         `json:"host"`
	Port      int32          `json:"port"`
	Interface string         `json:"interface"`
	Public    *GearmanConfig `json:"public"`
}

func (c *GearmanConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *GearmanConfig) SetHost(host string) {
	c.Host = host
}

func (c *GearmanConfig) SetPort(port int32) {
	c.Port = port
}

func (c *GearmanConfig) LoadDefaults() error {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 4730
	}
	return nil
}

func (c *GearmanConfig) GetInterface() string {
	return c.Interface
}

func (c *GearmanConfig) GetPublic() serviceconfig.ServiceConfig {
	return c.Public
}

func (c *GearmanConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
