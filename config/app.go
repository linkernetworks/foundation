package config

// TODO: move to the aurora repository
type AppConfig struct {
	Brand    BrandConfig     `json:"brand"`
	Session  *SessionConfig  `json:"session"`
	Socketio *SocketioConfig `json:"socketio"`

	BaseURL string `json:"baseURL" env:"AURORA_BASE_URL"`

	EnableAuthentication bool `json:"enableAuthentication" env:"AURORA_ENABLE_AUTHENTICATION"`

	// TODO: move this the dataset config
	MaxThumbnailWidth  uint `json:"maxThumbnailWidth"`
	MaxThumbnailHeight uint `json:"maxThumbnailHeight"`

	DbVersion   string `json:"dbVersion" env:"AURORA_DB_VERSION"`
	Version     string `json:"version"`
	LogFileName string `json:"logFileName"`
}

func (a *AppConfig) GetJobSetailURL(jobID string) string {
	return a.BaseURL + "/#/jobs/view/" + jobID
}
