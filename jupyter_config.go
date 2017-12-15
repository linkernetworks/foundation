package config

type JupyterConfig struct {
	BaseUrl      string        `json:"baseUrl"`
	LocalBaseUrl string        `json:"localBaseUrl"`
	Localhost    string        `json:"localhost"`
	Session      SessionConfig `json:"session:`
}
