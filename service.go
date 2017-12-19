package gearman

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"github.com/mikespook/gearman-go/client"
)

type Service struct {
	Bind string
}

func NewFromConfig(cf *config.GearmanConfig) *Service {
	addr := cf.Addr()
	return &Service{Bind: addr}
}

func (g *Service) NewClient() *client.Client {
	c, err := client.New(client.Network, g.Bind)
	if err != nil {
		logger.Fatal(err)
	}
	/*
		c.ErrorHandler = func(e error) {
			logger.Info("gearman client error:", e)
		}
	*/
	return c
}
