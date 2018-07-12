package logger

type LoggerConfig struct {
	Dir           string `json:"dir"`
	SuffixPattern string `json:"suffixPattern"`
	LinkName      string `json:"linkName"`
	Level         string `json:"level"`
	MaxAge        string `json:"maxAge"`
}
