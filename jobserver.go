package config

import (
	"net"
	"strconv"
	"time"
)

type JobServerConfig struct {
	Host              string                      `json:"host"`
	Port              int32                       `json:"port"`
	LogFileName       string                      `json:"logFileName"`
	DeploymentTargets map[string]DeploymentConfig `json:"deploymentTargets"`
	TickerSec         time.Duration               `json:"tickerSec"`
}

func (c *JobServerConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 50051
	}
	return nil
}

func (c *JobServerConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
}
