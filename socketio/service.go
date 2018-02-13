package socketio

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"errors"
	"fmt"
	socketio "github.com/c9s/go-socket.io"
	redigo "github.com/garyburd/redigo/redis"
	"io"
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
func (s *Service) NewClientSubscription(token string, socket socketio.Socket, psc *redigo.PubSubConn, toEvent string) *Client {
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

func (s *Service) Subscribe(token string, topic string) error {
	client, ok := s.Clients[token]
	if !ok {
		msg := fmt.Sprintf("Client token: %s not found and can't subscribe.", token)
		return errors.New(msg)
	}
	return client.Subscribe(topic)
}

func (s *Service) Unsubscribe(token string, topic string) error {
	client, ok := s.Clients[token]
	if !ok {
		msg := fmt.Sprintf("Client token: %s not found and can't unsubscribe.", token)
		return errors.New(msg)
	}
	return client.Unsubscribe(topic)
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
				logger.Error(err)
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

type Client struct {
	*redigo.PubSubConn
	socket    socketio.Socket
	channel   chan string
	stopPipe  chan bool
	stopEmit  chan bool
	expiredAt time.Time
	toEvent   string
}

func (c *Client) SetSocket(socket socketio.Socket) {
	c.socket = socket
}

func (c *Client) KeepAlive(timeout time.Duration) {
	c.expiredAt = time.Now().Add(timeout)
}

// pipe from redigo pubsubconn to chan
func (c *Client) pipe() error {
Pipe:
	for {
		select {
		case <-c.stopPipe:
			c.stopPipe <- true
			break Pipe
		default:
			switch v := c.Receive().(type) {
			case redigo.Message:
				c.channel <- string(v.Data)
				logger.Debugf("redis: received message channel: %s message: %s", v.Channel, v.Data)
			case redigo.Subscription:
				// v.Kind could be "subscribe", "unsubscribe" ...
				logger.Debugf("redis: subscription channel:%s kind:%s count:%d", v.Channel, v.Kind, v.Count)
				if v.Count == 0 {
					return nil
				}
			// when the connection is closed, redigo returns an error "connection closed" here
			case error:
				if v != io.EOF {
					logger.Errorf("redis: error=%v", v)
				}
				break Pipe
			}
		}
	}
	logger.Debugf("redis: pipe exited")
	return nil
}

// emit chan message to socket event
func (c *Client) emit() {
Emit:
	for {
		select {
		case <-c.stopEmit:
			logger.Debug("socketio: channel recieve close signal")
			c.stopEmit <- true
			break Emit

		case msg := <-c.channel:
			if err := c.socket.Emit(c.toEvent, msg); err != nil {
				logger.Errorf("socketio: emit error. %s", err.Error())
			}
		}
	}
}
func (c *Client) Stop() {
	c.stopPipe <- true
	<-c.stopPipe
	c.stopEmit <- true
	<-c.stopEmit
}
