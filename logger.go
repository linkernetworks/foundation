package config

type LoggerConfig struct {
	Dir         string `json:"dir"`
	FilePattern string `json:"filePattern"`
	LinkName    string `json:"linkName"`
	Level       string `json:"level"`
	Reserve     int    `json:"reserve"` // MaxAge in days
}
