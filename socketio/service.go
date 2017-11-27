package socketio

import (
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

func NewService() *Service {
	io, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}
	return &Service{
		Server:  io,
		clients: map[string]*client{},
	}
}

type client struct {
	socket     socketio.Socket
	channel    chan string
	expiredAt  int64
	pubSubConn *redis.PubSubConn
	toEvent    string
	closed     bool
}

// Create client with a given token. Front-End can recover client with ths same token
func (s *Service) NewClientSubscription(token string, socket socketio.Socket, psc *redis.PubSubConn, toEvent string) {
	// try find existed client with token, generate new client if not found
	if existedClient, ok := s.clients[token]; !ok {
		// Create new client
		//TODO import logger
		// fmt.Printf("WS a new client socketId: %s connected with new token %s.\n", socket.Id(), token)
		client := &client{
			socket:     socket,
			expiredAt:  time.Now().Unix() + 5*60,
			pubSubConn: psc,
			toEvent:    toEvent,
			closed:     false,
		}
		client.channel = make(chan string, 10)
		s.clients[token] = client

		go client.pipe() // from redis to chan
		go client.emit() // to socket event

	} else {
		// if token already exist, replace disconnected socket with new socket
		fmt.Printf("WS a client socketId: %s reconnected with token %s.\n", socket.Id(), token)
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
	for {
		switch v := c.pubSubConn.Receive().(type) {
		case redis.Message:
			c.channel <- string(v.Data)
			//TODO use logger instead
			// fmt.Printf("REDIS: received message channel: %s message: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			// v.Kind could be "subscribe", "unsubscribe" ...
			fmt.Printf("REDIS: subscription channel:%s kind:%s count:%d\n", v.Channel, v.Kind, v.Count)
		// when the connection is closed, redigo returns an error "connection closed" here
		case error:
			fmt.Printf("REDIS: pubsub error, exiting:", v)
			return v
		}
	}
	fmt.Println("REDIS: pipe exited")
	return nil
}

// emit chan message to socket event
func (c *client) emit() {
	for msg := range c.channel {
		if err := c.socket.Emit(c.toEvent, msg); err != nil {
			// FIXME emit EOF
			//fmt.Printf("REDIS: emit error. %s \n", err.Error())
		}
	}
}

func (s *Service) Close(clientId string) error {
	client := s.clients[clientId]
	if err := client.pubSubConn.Close(); err != nil {
		return err
	}

	if client.closed {
		return nil
	}

	close(client.channel)
	delete(s.clients, clientId)
	client.pubSubConn.Close()
	client.socket.Disconnect()
	client.closed = true

	return nil
}

func (s *Service) CleanUp() error {
	now := time.Now().Unix()
	for token, client := range s.clients {
		if client.expiredAt < now {
			if err := s.Close(token); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Refresh(clientId string) error {
	if len(clientId) == 0 {
		msg := fmt.Sprint("Token is empty. Can't refresh.")
		return errors.New(msg)
	}
	client, ok := s.clients[clientId]
	if !ok {
		msg := fmt.Sprint("Client: %s not found and can't refresh.", clientId)
		return errors.New(msg)
	}
	client.expiredAt = time.Now().Unix() + 5*60
	return nil
}
