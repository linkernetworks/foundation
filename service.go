package socketio

import (
	"fmt"
	redis "github.com/garyburd/redigo/redis"
	socketio "github.com/googollee/go-socket.io"
	"time"
)

type Service struct {
	Server  *socketio.Server
	clients map[string]client
}

func NewService() *Service {
	io, err := socketio.NewServer(nil)
	if err != nil {
		panic(err)
	}
	return &Service{
		Server:  io,
		clients: map[string]client{},
	}
}

type client struct {
	socket     socketio.Socket
	channel    chan string
	expiredAt  int64
	pubSubConn *redis.PubSubConn
	toEvent    string
}

func (s *Service) NewClientSubscription(socket socketio.Socket, psc *redis.PubSubConn, toEvent string) {
	client := client{
		socket:     socket,
		expiredAt:  time.Now().Unix() + 3600,
		pubSubConn: psc,
		toEvent:    toEvent,
	}
	s.clients[socket.Id()] = client

	go client.pipe() // from redis to chan
	go client.emit() // to socket event
}

func (s *Service) Subscribe(clientId string, topic string) error {
	return s.clients[clientId].pubSubConn.Subscribe(topic)
}

func (s *Service) UnSubscribe(clientId string, topic string) error {
	return s.clients[clientId].pubSubConn.Unsubscribe(topic)
}

// pipe from redis pubsubconn to chan
func (c *client) pipe() error {
	for {
		switch v := c.pubSubConn.Receive().(type) {
		case redis.Message:
			c.channel <- string(v.Data)
			//TODO use logger instead
			fmt.Sprintf("REDIS: received message %s: %s\n", v.Channel, v.Data)
		case redis.Subscription:
			// v.Kind could be "subscribe", "unsubscribe" ...
			fmt.Sprintf("REDIS: subscription channel:%s kind:%s count:%d\n", v.Channel, v.Kind, v.Count)
		// when the connection is closed, redigo returns an error "connection closed" here
		case error:
			fmt.Sprintf("REDIS: pubsub error, exiting:", v)
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
			fmt.Println(err.Error())
		}
	}
}

func (s *Service) Close(clientId string) error {
	client := s.clients[clientId]
	if err := client.pubSubConn.Close(); err != nil {
		return err
	}
	close(client.channel)

	delete(s.clients, clientId)
	return nil
}

func (s *Service) CleanUp() error {
	now := time.Now().Unix()
	for id, client := range s.clients {
		if client.expiredAt < now {
			if err := s.Close(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Refresh(clientId string) {
	client := s.clients[clientId]
	client.expiredAt = time.Now().Unix() + 3600
}
