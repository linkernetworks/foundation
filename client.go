package socketio

import (
	"io"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/logger"

	socketio "github.com/c9s/go-socket.io"

	redigo "github.com/garyburd/redigo/redis"
)

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
PIPE:
	for {
		select {
		case <-c.stopPipe:
			c.stopPipe <- true
			break PIPE
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
				break PIPE
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
