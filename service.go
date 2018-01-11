package socketio

import (
	"bitbucket.org/linkernetworks/aurora/src/logger"
	"errors"
	"fmt"
	redis "github.com/garyburd/redigo/redis"
	socketio "github.com/googollee/go-socket.io"
	"time"
)

type Service struct {
	Server  *socketio.Server
	clients map[string]*client
}

func NewService(maxConnection int) *Service {
	io, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}
	io.SetMaxConnection(maxConnection)
	return &Service{
		Server:  io,
		clients: map[string]*client{},
	}
}

type client struct {
	socket     socketio.Socket
	channel    chan string
	stopPipe   chan bool
	stopEmit   chan bool
	expiredAt  int64
	pubSubConn *redis.PubSubConn
	toEvent    string
	closed     bool
}

// Create client with a given token. Front-End can recover client with ths same token
func (s *Service) NewClientSubscription(token string, socket socketio.Socket, psc *redis.PubSubConn, toEvent string) {
	// try find existed client with token, generate new client if not found

	existedClient, ok := s.clients[token]
	if !ok || existedClient.closed {
		// Create new client
		logger.Infof("WS a new client socketId: %s connected with new token %s.", socket.Id(), token)
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

		go client.pipe() // from redis to chan
		go client.emit() // to socket event

	} else {
		// if token already exist and client still valid, replace disconnected socket with new socket
		logger.Infof("WS a client socketId: %s reconnected with token %s.", socket.Id(), token)
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

// pipe from redis pubsubconn to chan
func (c *client) pipe() error {
Pipe:
	for {
		select {
		case <-c.stopPipe:
			c.stopPipe <- true
			break Pipe
		default:
			switch v := c.pubSubConn.Receive().(type) {
			case redis.Message:
				c.channel <- string(v.Data)
				logger.Debugf("REDIS: received message channel: %s message: %s", v.Channel, v.Data)
			case redis.Subscription:
				// v.Kind could be "subscribe", "unsubscribe" ...
				logger.Debugf("REDIS: subscription channel:%s kind:%s count:%d", v.Channel, v.Kind, v.Count)
				if v.Count == 0 {
					return nil
				}
			// when the connection is closed, redigo returns an error "connection closed" here
			case error:
				logger.Error("REDIS: ", v)
				//	return v
				break Pipe
			}
		}
	}
	fmt.Println("REDIS: pipe exited")
	return nil
}

// emit chan message to socket event
func (c *client) emit() {
Emit:
	for {
		select {
		case <-c.stopEmit:
			logger.Debug("SOCKET: channel recieve close signal")
			c.stopEmit <- true
			break Emit

		case msg := <-c.channel:
			if err := c.socket.Emit(c.toEvent, msg); err != nil {
				logger.Errorf("SOCKET: emit error. %s", err.Error())
			}
		}
	}
}
func (s *Service) CleanUp() error {
	now := time.Now().Unix()
	logger.Debugf("SOCKET: clean up triggered. Server connection: %v. Active client: %v", s.Count(), len(s.clients))
	for token, client := range s.clients {
		// send close to client channel
		if client.closed == false && client.expiredAt < now {
			logger.Debugf("SOCKET: marking client token: %s as closed. Active clients: %v", token, len(s.clients))
			client.closed = true

		} else if client.closed {
			logger.Debugf("SOCKET: closing client token: %s. Active clients: %v", token, len(s.clients))

			client.Stop()
			if err := client.pubSubConn.Unsubscribe(); err != nil { // Unsubscribe all
				logger.Error(err)
			}
			client.socket.Disconnect()
			close(client.channel)
			delete(s.clients, token)
		}
	}
	return nil
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
	client, ok := s.clients[clientId]
	if !ok {
		msg := fmt.Sprintf("Client: %s not found and can't refresh.", clientId)
		return errors.New(msg)
	}
	client.expiredAt = time.Now().Unix() + 5*60
	return nil
}

// Count returns the current number of connected clients in session
func (s *Service) Count() int {
	return s.Server.Count()
}
