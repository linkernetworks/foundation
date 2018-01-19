package config

type JobUpdaterConfig struct {
	BufferSize int          `json:"bufferSize"`
	Logger     LoggerConfig `json:"logger"`
}
