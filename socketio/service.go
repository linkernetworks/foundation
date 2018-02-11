package socketio

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"errors"
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	socketio "github.com/googollee/go-socket.io"
	"io"
	"time"
)

type Service struct {
	Server            *socketio.Server
	clients           map[string]*client
	connectionTimeout time.Duration
}

func New(cf *config.SocketioConfig) *Service {
	io, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}

	if cf.Ping.Interval != 0 {
		io.SetPingInterval(time.Duration(cf.Ping.Interval) * time.Second)
	}
	if cf.Ping.Timeout != 0 {
		io.SetPingTimeout(time.Duration(cf.Ping.Timeout) * time.Second)
	}
	if cf.MaxConnection != 0 {
		io.SetMaxConnection(cf.MaxConnection)
	}
	return &Service{
		Server:            io,
		clients:           map[string]*client{},
		connectionTimeout: 5 * time.Minute,
	}
}

type client struct {
	socket     socketio.Socket
	channel    chan string
	stopPipe   chan bool
	stopEmit   chan bool
	expiredAt  int64
	pubSubConn *redigo.PubSubConn
	toEvent    string
	closed     bool
}

// Create client with a given token. Front-End can recover client with ths same token
func (s *Service) NewClientSubscription(token string, socket socketio.Socket, psc *redigo.PubSubConn, toEvent string) {
	// try find existed client with token, generate new client if not found

	existedClient, ok := s.clients[token]
	if !ok || existedClient.closed {
		// Create new client
		logger.Infof("socketio: a new client connected with new token: id=%s token=%s.", socket.Id(), token)
		client := &client{
			socket:     socket,
			expiredAt:  time.Now().Unix() + 5*60,
			pubSubConn: psc,
			toEvent:    toEvent,
			closed:     false,
			channel:    make(chan string, 100),
			stopEmit:   make(chan bool, 1),
			stopPipe:   make(chan bool, 1),
		}
		s.clients[token] = client

		go client.pipe() // from redigo to chan
		go client.emit() // to socket event

	} else {
		// if token already exist and client still valid, replace disconnected socket with new socket
		logger.Infof("socketio: a client reconnected with token id=%s token=%s.", socket.Id(), token)
		existedClient.socket = socket
	}
}

func (s *Service) Subscribe(token string, topic string) error {
	client, ok := s.clients[token]
	if !ok {
		msg := fmt.Sprintf("Client token: %s not found and can't subscribe.", token)
		return errors.New(msg)
	}
	return client.pubSubConn.Subscribe(topic)
}

func (s *Service) UnSubscribe(token string, topic string) error {
	client, ok := s.clients[token]
	if !ok {
		msg := fmt.Sprintf("Client token: %s not found and can't unsubscribe.", token)
		return errors.New(msg)
	}
	return client.pubSubConn.Unsubscribe(topic)
}

// pipe from redigo pubsubconn to chan
func (c *client) pipe() error {
Pipe:
	for {
		select {
		case <-c.stopPipe:
			c.stopPipe <- true
			break Pipe
		default:
			switch v := c.pubSubConn.Receive().(type) {
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
func (c *client) emit() {
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
func (s *Service) CleanUp() (lasterr error) {
	now := time.Now().Unix()
	for token, client := range s.clients {
		// send close to client channel
		if client.closed == false && client.expiredAt < now {
			logger.Debugf("socketio: marking client=%s as closed. Active clients: %v", token, len(s.clients))
			client.closed = true

		} else if client.closed {
			logger.Debugf("socketio: closing client=%s. Active clients: %v", token, len(s.clients))
			client.Stop()
			if err := client.pubSubConn.Unsubscribe(); err != nil { // Unsubscribe all
				logger.Error(err)
				lasterr = err
			}
			client.socket.Disconnect()
			delete(s.clients, token)
		}
	}
	return lasterr
}

func (c *client) Stop() {
	c.stopPipe <- true
	<-c.stopPipe
	c.stopEmit <- true
	<-c.stopEmit
}

func (s *Service) Refresh(clientId string) error {
	if len(clientId) == 0 {
		return errors.New("Token provided by front-end is empty. Can't refresh.")
	}
	if client, ok := s.clients[clientId]; ok {
		client.expiredAt = time.Now().Add(s.connectionTimeout).Unix()
		return nil
	}
	return fmt.Errorf("Client: %s not found and can't refresh.", clientId)
}

// Count returns the current number of connected clients in session
func (s *Service) Count() int {
	return s.Server.Count()
}
