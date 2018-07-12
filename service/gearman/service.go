package gearman

import (
	"github.com/linkernetworks/logger"
	"github.com/mikespook/gearman-go/client"
)

type Service struct {
	Bind string
}

type Client struct {
	*client.Client
}

type JobCallbacks struct {
	OnException func(data []byte, err error) error
	OnFail      func(data []byte, err error) error
	OnComplete  func(data []byte, err error) error
	OnStatus    func(data []byte, err error) error
}

func NewFromConfig(cf *GearmanConfig) *Service {
	return &Service{Bind: cf.Addr()}
}

func New(bind string) *Service {
	return &Service{Bind: bind}
}

func (g *Service) NewClient() (*Client, error) {
	c, err := client.New(client.Network, g.Bind)
	if err != nil {
		return nil, err
	}

	/*
		c.ErrorHandler = func(e error) {
			logger.Error("gearman client error:", e)
		}
	*/
	return &Client{c}, nil
}

func (gc *Client) Call(workerFn string, payload []byte, callbacks JobCallbacks) (chan error, error) {
	var s = make(chan error)
	jobHandler := func(resp *client.Response) {
		switch resp.DataType {
		case client.WorkException:
			data, err := resp.Result()
			logger.Error("Worker return a warning status for a jobs. code: ", resp.DataType)
			if callbacks.OnException != nil {
				callbacks.OnException(data, err)
			}
			s <- err
			close(s)
			return
		case client.WorkFail:
			data, err := resp.Result()
			logger.Error("Worker return a fail status for a jobs. code: ", resp.DataType)
			if callbacks.OnFail != nil {
				callbacks.OnFail(data, err)
			}
			s <- err
			close(s)
			return
		case client.WorkStatus:
			data, err := resp.Result()
			if err != nil {
				logger.Error("Worker returned error:", err)
			}
			if callbacks.OnStatus != nil {
				callbacks.OnStatus(data, err)
			}
			return
		case client.WorkComplete:
			data, err := resp.Result()
			if err != nil {
				logger.Error("Worker returned error:", err)
			}
			if callbacks.OnComplete != nil {
				s <- callbacks.OnComplete(data, err)
			} else {
				s <- nil
			}
			close(s)
			return
		case client.WorkWarning:
			logger.Error("Worker return a warning status for a jobs. code: ", resp.DataType)
			panic("Worker return a warning status for a jobs.")
		default:
			// might be WorkData check gearman-go/client/common.go for data tyoe code
			logger.Debug("DEFAULT case: if hit this case checking the gearman DataType", resp.DataType)
		}
		s <- nil
	}
	handle, err := gc.Do(workerFn, payload, client.JobNormal, jobHandler)
	if err != nil {
		return s, err
	}

	status, err := gc.Status(handle)
	logger.Debug(handle, *status)

	if err != nil {
		return s, err
	}

	return s, err
}
