package socketio

import (
	"github.com/linkernetworks/config"
	"github.com/linkernetworks/logger"
	"errors"
	"fmt"
	socketio "github.com/c9s/go-socket.io"
	redigo "github.com/gomodule/redigo/redis"
	"sync"
	"time"
)

type Service struct {
	sync.Mutex
	Server            *socketio.Server
	Clients           map[string]*Client
	ConnectionTimeout time.Duration
}

func New(cf *config.SocketioConfig) *Service {
	server, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}

	if cf.Ping.Interval != 0 {
		server.SetPingInterval(cf.Ping.Interval * time.Second)
	}
	if cf.Ping.Timeout != 0 {
		server.SetPingTimeout(cf.Ping.Timeout * time.Second)
	}
	if cf.MaxConnection != 0 {
		server.SetMaxConnection(cf.MaxConnection)
	}
	return &Service{
		Server:            server,
		Clients:           map[string]*Client{},
		ConnectionTimeout: 5 * time.Minute,
	}
}

func (s *Service) GetClient(token string) (c *Client, ok bool) {
	s.Lock()
	c, ok = s.Clients[token]
	s.Unlock()
	return c, ok
}

// Create client with a given token. Front-End can recover client with ths same token
func (s *Service) NewClient(token string, socket socketio.Socket, psc *redigo.PubSubConn, toEvent string) *Client {
	// Create new client
	logger.Infof("[socketio] a new client connected: id=%s token=%s.", socket.Id(), token)
	client := &Client{
		PubSubConn: psc,
		Socket:     socket,
		BufSize:    100,
		ExpiredAt:  time.Now().Add(10 * time.Minute),
		toEvent:    toEvent,
	}

	// critical section
	s.Lock()
	s.Clients[token] = client
	s.Unlock()

	client.Start()

	return client
}

func (s *Service) CleanUp() (lasterr error) {
	var now = time.Now()
	var expiredTokens []string
	for token, client := range s.Clients {
		// send close to client channel
		if client.ExpiredAt.Before(now) {
			expiredTokens = append(expiredTokens, token)
		}
	}

	s.Lock()
	for _, token := range expiredTokens {
		if client, ok := s.Clients[token]; ok {
			client.Stop()
			if client.Socket != nil {
				client.Socket.Disconnect()
			}
			delete(s.Clients, token)
		}
	}
	s.Unlock()

	return lasterr
}

func (s *Service) Refresh(clientId string) error {
	if len(clientId) == 0 {
		return errors.New("Token provided by front-end is empty. Can't refresh.")
	}

	s.Lock()
	defer s.Unlock()

	if client, ok := s.Clients[clientId]; ok {
		client.KeepAlive(s.ConnectionTimeout)
		return nil
	}
	return fmt.Errorf("Client: %s not found and can't refresh.", clientId)
}

// Count returns the current number of connected clients in session
func (s *Service) Count() int {
	return s.Server.Count()
}
