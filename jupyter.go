package config

type JupyterConfig struct {
	BaseURL string `json:"baseUrl"`

	// the address that the jupyter notebook will bind to "0.0.0.0"
	Address string `json:"bind"`

	// the default jupyternotebook docker image name
	DefaultImage string `json:"defaultImage"`

	// the working dir of the jupyter notebook process
	WorkingDir string `json:"workingDir"`

	// the cache configuration
	Cache *JupyterCacheConfig `json:"cache"`

	// proxy configuration that will be used for development mode.
	Dev *DevProxyConfig `json:"dev"`
}

func (c *JupyterConfig) LoadDefaults() {
	if c.Cache != nil {
		if c.Cache.Expire == 0 {
			// 60 seconds
			c.Cache.Expire = 10 * 60
		}
	}
}

type JupyterCacheConfig struct {
	Age    int `json:"age"`
	Expire int `json:"expire"`
}

type DevProxyConfig struct {
	BaseURL     string `json:"baseUrl"`
	HostAddress string `json:"hostAddress"`
}
