package kudis

import (
	"github.com/linkernetworks/config"
)

type Service struct {
	Config *config.KudisConfig
}

func New(cf *config.KudisConfig) *Service {
	return &Service{cf}
}

func (s *Service) NewClient() (*Client, error) {
	return NewInsecure(s.Config.Addr())
}
