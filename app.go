package config

type AppConfig struct {
	Logger   LoggerConfig    `json:"logger"`
	Brand    BrandConfig     `json:"brand"`
	Session  *SessionConfig  `json:"session"`
	Socketio *SocketioConfig `json:"socketio"`
	BaseURL  string          `json:"baseURL"`

	EnableAuthentication bool `json:"enableAuthentication"`

	// TODO: move this the dataset config
	MaxThumbnailWidth  uint   `json:"maxThumbnailWidth"`
	MaxThumbnailHeight uint   `json:"maxThumbnailHeight"`
	DbVersion          string `json:"dbVersion"`
	Version            string `json:"version"`
}

func (a *AppConfig) GetJobSetailURL(jobID string) string {
	return a.BaseURL + "/#/jobs/view/" + jobID
}
