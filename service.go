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
}

func (s *Service) Reconnect(socket socketio.Socket, token string) error {
	if existedClient, ok := s.clients[token]; ok {
		// Replace disconnected socket with new socket
		existedClient.socket = socket
	} else {
		msg := fmt.Sprintf("Try to reconnect previous client token: %s but not found.", token)
		return errors.New(msg)
	}
	return nil
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
			fmt.Printf("REDIS: received message %s: %s\n", v.Channel, v.Data)
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
			fmt.Printf("REDIS: emit error. %s", err.Error())
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
			fmt.Println("Clean up socket clients")
			if err := s.Close(token); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Refresh(clientId string) error {
	client, ok := s.clients[clientId]
	if !ok {
		msg := fmt.Sprint("Client: %s not found and can't refresh.", clientId)
		return errors.New(msg)
	}
	client.expiredAt = time.Now().Unix() + 5*60
	return nil
}
