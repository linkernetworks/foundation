package config

// NotificationConfig describes the NotificationSetting in JSON configuration file.
type NotificationConfig struct {
	OnJobStart   bool `json:"on_job_start"`
	OnJobSuccess bool `json:"on_job_success"`
	OnJobFail    bool `json:"on_job_fail"`
	OnJobStop    bool `json:"on_job_stop"`
	OnJobDelete  bool `json:"on_job_delete"`
}
