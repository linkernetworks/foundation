package config

type JupyterConfig struct {
	BaseUrl string `json:"baseUrl"`

	// default bind to "0.0.0.0"
	Bind string `json:"bind"`

	DefaultImage string          `json:"defaultImage"`
	WorkingDir   string          `json:"workingDir"`
	Cache        *CacheConfig    `json:"cache"`
	Dev          *DevProxyConfig `json:"dev"`
}

type CacheConfig struct {
	Prefix string `json:"prefix"`
	Age    int    `json:"age"`
}

type DevProxyConfig struct {
	BaseUrl     string `json:"baseUrl"`
	HostAddress string `json:"hostUrl"`
}
