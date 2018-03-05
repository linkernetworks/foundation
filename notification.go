package config

// NotificationConfig describes the NotificationSetting in JSON configuration file.
// Refer to src/oauth/entity/notification_setting.go
type NotificationConfig struct {
	EnableSMS    bool `json:"enable_sms"`
	EnableEmail  bool `json:"enable_email"`
	OnJobStart   bool `json:"on_job_start"`
	OnJobSuccess bool `json:"on_job_success"`
	OnJobFail    bool `json:"on_job_fail"`
	OnJobStop    bool `json:"on_job_stop"`
	OnJobDelete  bool `json:"on_job_delete"`
}
