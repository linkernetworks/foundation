package timer

import (
	"time"

	"github.com/linkernetworks/foundation/logger"
)

type TimerHandler func() error

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
					go func() {
						if err := handler(); err != nil {
							logger.Errorf("%s: error=%v", key, err)
						}
					}()
				}

			case <-signal:
				ticker.Stop()
				return
			}
		}
	}()

	return signal
}
