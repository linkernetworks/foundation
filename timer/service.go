package timer

import (
	"time"

	"bitbucket.org/linkernetworks/aurora/src/logger"
)

type TimerHandler func()

type TimerService struct {
	Interval time.Duration
	Handlers map[string]TimerHandler
}

func New(interval time.Duration) *TimerService {
	return &TimerService{
		Interval: interval,
		Handlers: map[string]TimerHandler{},
	}
}

func (s *TimerService) Bind(key string, handler TimerHandler) {
	s.Handlers[key] = handler
}

func (s *TimerService) Run() chan struct{} {
	ticker := time.NewTicker(s.Interval)

	signal := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				for key, handler := range s.Handlers {
					logger.Debugf("Running timer handler: %s", key)
					go handler()
				}

			case <-signal:
				ticker.Stop()
				return
			}
		}
	}()

	return signal
}
