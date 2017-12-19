package config

type AppConfig struct {
	Logger               LoggerConfig   `json:"logger"`
	Session              *SessionConfig `json:"session"`
	EnableAuthentication bool           `json:"enableAuthentication"`
	MaxThumbnailWidth    uint           `json:"maxThumbnailWidth"`
	MaxThumbnailHeight   uint           `json:"maxThumbnailHeight"`
	DbVersion            string         `json:"dbVersion"`
	Version              string         `json:"version"`
}

type LoggerConfig struct {
	Dir         string `json:"dir"`
	FilePattern string `json:"filePattern"`
	LinkName    string `json:"linkName"`
	Level       string `json:"level"`
}

type SessionConfig struct {
	Size     int    `json:"size"`
	Protocal string `json:"protocal"`
	RedisUrl string `json:"redisUrl"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	KeyPair  string `json:"keyPair"`
}
