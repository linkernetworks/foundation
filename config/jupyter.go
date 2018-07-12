package config

type JupyterConfig struct {
	BaseURL string `json:"baseUrl"`

	// the default jupyternotebook docker image name
	DefaultImage string `json:"defaultImage"`

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
