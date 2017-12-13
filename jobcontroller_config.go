package config

import (
	"encoding/json"
	"net"
	"strconv"
	"time"
)

type JobControllerConfig struct {
	Host              string                      `json:"host"`
	Port              int                         `json:"port"`
	Logger            LoggerConfig                `json:"logger"`
	DeploymentTargets map[string]DeploymentConfig `json:"deploymentTargets"`
	TickerSec         time.Duration               `json:"tickerSec"`
}

func (c *JobControllerConfig) LoadDefaults() error {
	if c.Port == 0 {
		c.Port = 50051
	}
	return nil
}

func (c *JobControllerConfig) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c *JobControllerConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	return c.LoadDefaults()
}
