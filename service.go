package socketio

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"errors"
	"fmt"
	socketio "github.com/c9s/go-socket.io"
	redigo "github.com/garyburd/redigo/redis"
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
	io, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}

	if cf.Ping.Interval != 0 {
		io.SetPingInterval(cf.Ping.Interval * time.Second)
	}
	if cf.Ping.Timeout != 0 {
		io.SetPingTimeout(cf.Ping.Timeout * time.Second)
	}
	if cf.MaxConnection != 0 {
		io.SetMaxConnection(cf.MaxConnection)
	}
	return &Service{
		Server:            io,
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
	logger.Infof("socketio: a new client connected with new token: id=%s token=%s.", socket.Id(), token)
	client := &Client{
		PubSubConn: psc,
		socket:     socket,
		expiredAt:  time.Now().Add(10 * time.Minute),
		toEvent:    toEvent,
		channel:    make(chan string, 100),
		stopEmit:   make(chan bool, 1),
		stopPipe:   make(chan bool, 1),
	}

	// critical section
	s.Lock()
	s.Clients[token] = client
	s.Unlock()

	go client.pipe() // from redigo to chan
	go client.emit() // to socket event

	return client
}

func (s *Service) CleanUp() (lasterr error) {
	now := time.Now()

	s.Lock()
	defer s.Unlock()
	for token, client := range s.Clients {
		// send close to client channel
		if client.expiredAt.Before(now) {
			client.Stop()
			if err := client.Unsubscribe(); err != nil { // Unsubscribe all
				logger.Error("client failed to unsubscribe: %v", err)
				lasterr = err
			}
			client.socket.Disconnect()
			delete(s.Clients, token)
		}
	}
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
