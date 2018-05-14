package config

type JobUpdaterConfig struct {
	BufferSize  int    `json:"bufferSize"`
	LogFileName string `json:"logFileName"`
}
