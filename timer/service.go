package timer

import (
	"time"

	"bitbucket.org/linkernetworks/aurora/src/logger"
)

type TimerHandler func()

type TimerService struct {
	Handlers map[string]TimerHandler
}

func New() *TimerService {
	return &TimerService{
		Handlers: map[string]TimerHandler{},
	}
}

func (s *TimerService) Bind(key string, handler TimerHandler) {
	s.Handlers[key] = handler
}

func (s *TimerService) Run() {
	ticker := time.NewTicker(3 * time.Minute)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				for key, handler := range s.Handlers {
					logger.Debugf("Running timer handler: %s", key)
					go handler()
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
