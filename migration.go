package config

type MigrationConfig struct {
	Host   string       `json:"host"`
	Port   int32        `json:"port"`
	Logger LoggerConfig `json:"logger"`
}
