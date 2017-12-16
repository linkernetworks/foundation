package config

type JupyterConfig struct {
	BaseUrl string        `json:"baseUrl"`
	Session SessionConfig `json:"session"`
	Dev     DevConfig     `json:"dev"`
}

type DevConfig struct {
	BaseUrl string `json:"baseUrl"`
	HostUrl string `json:"hostUrl"`
}
