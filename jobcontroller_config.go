package config

import (
	"fmt"
	"time"
)

type JobControllerConfig struct {
	Host              string                      `json:"host"`
	Port              int                         `json:"port"`
	Logger            LoggerConfig                `json:"logger"`
	DeploymentTargets map[string]DeploymentConfig `json:"deploymentTargets"`
	TickerSec         time.Duration               `json:"tickerSec"`
}

func (c *JobControllerConfig) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
