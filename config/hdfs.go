package config

import (
	"github.com/linkernetworks/serviceconfig"
	"net"
	"strconv"
)

type HdfsConfig struct {
	Enabled   bool        `json:"enabled"`
	Host      string      `json:"host"`
	Port      int32       `json:"port"`
	Username  string      `json:"username"`
	Interface string      `json:"interface"`
	Public    *HdfsConfig `json:"public"`
}

func (c *HdfsConfig) Unresolved() bool {
	return c.Host == ""
}

func (c *HdfsConfig) SetHost(host string) {
	c.Host = host
}

func (c *HdfsConfig) SetPort(port int32) {
	c.Port = port
}

func (c *HdfsConfig) SetUsername(username string) {
	c.Username = username
}

func (c *HdfsConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 8020
	}

	if len(c.Username) == 0 {
		c.Username = "root"
	}
	return nil
}

func (c *HdfsConfig) GetInterface() string {
	return c.Interface
}

func (c *HdfsConfig) GetPublic() serviceconfig.ServiceConfig {
	return c.Public

}

func (c *HdfsConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
