package config

type JupyterConfig struct {
	BaseURL string `json:"baseUrl"`

	// default bind to "0.0.0.0"
	Bind string `json:"bind"`

	DefaultImage string              `json:"defaultImage"`
	WorkingDir   string              `json:"workingDir"`
	Cache        *JupyterCacheConfig `json:"cache"`
	Dev          *DevProxyConfig     `json:"dev"`
}

type JupyterCacheConfig struct {
	Prefix string `json:"prefix"`
	Age    int    `json:"age"`
}

type DevProxyConfig struct {
	BaseURL     string `json:"baseUrl"`
	HostAddress string `json:"hostUrl"`
}
