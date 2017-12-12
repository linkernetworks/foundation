package config

type HdfsConfig struct {
	Enabled   bool        `json:"enabled"`
	Host      string      `json:"host"`
	Port      int         `json:"port"`
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

func (c *HdfsConfig) SetPort(port int) {
	c.Port = port
}

func (c *HdfsConfig) SetUsername(username string) {
	c.Username = username
}

func (c *HdfsConfig) LoadDefaults() {
	if c.Port == 0 {
		c.Port = 8020
	}

	if len(c.Username) == 0 {
		c.Username = "root"
	}
}

func (c *HdfsConfig) GetInterface() string {
	return c.Interface
}

func (c *HdfsConfig) GetPublic() ServiceConfig {
	return c.Public
}
