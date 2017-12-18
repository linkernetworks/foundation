package config

type JupyterConfig struct {
	BaseUrl string       `json:"baseUrl"`
	Cache   *CacheConfig `json:"cache"`
	Dev     *DevConfig   `json:"dev"`
}

type CacheConfig struct {
	Prefix string `json:"prefix"`
	Age    int    `json:"age"`
}

type DevConfig struct {
	BaseUrl string `json:"baseUrl"`
	HostUrl string `json:"hostUrl"`
}
