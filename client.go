package socketio

import (
	"time"

	"bitbucket.org/linkernetworks/aurora/src/logger"

	socketio "github.com/c9s/go-socket.io"
	redigo "github.com/garyburd/redigo/redis"
)

type Client struct {
	*redigo.PubSubConn
	BufSize   int
	Socket    socketio.Socket
	C         chan string
	done      chan bool
	ExpiredAt time.Time
	toEvent   string
}

func (c *Client) SetSocket(socket socketio.Socket) {
	c.Socket = socket
}

func (c *Client) KeepAlive(timeout time.Duration) {
	c.ExpiredAt = time.Now().Add(timeout)
}

// pipe from redigo pubsubconn to chan
func (c *Client) read() {
PIPE:
	for {
		msg := c.Receive()
		switch v := msg.(type) {
		case redigo.Subscription:
			logger.Infof("subscription: kind=%s channel=%s count=%d", v.Kind, v.Channel, v.Count)
			if v.Count == 0 {
				break PIPE
			}
		case redigo.Message:
			c.C <- string(v.Data)
		// when the connection is closed, redigo returns an error "connection closed" here
		case error:
			logger.Errorf("redis: error=%v", v)
			break PIPE
		}
	}
	close(c.C)
	c.done <- true
}

// emit chan message to socket event
func (c *Client) write() {
	for msg := range c.C {
		if c.Socket != nil {
			if err := c.Socket.Emit(c.toEvent, msg); err != nil {
				logger.Errorf("socketio: event '%s' emit error: %v", c.toEvent, err)
			}
		}
	}
}

func (c *Client) Start() {
	c.C = make(chan string, c.BufSize)
	c.done = make(chan bool)

	// subscribe to system:events by default so the connectxion keeps at least
	// one subscription to prevent the read loop exits.
	c.Subscribe("system:events")

	go c.write() // to socket event
	go c.read()  // from redigo to chan
}

func (c *Client) Stop() {
	// unsubscribe all channels
	c.PubSubConn.Unsubscribe()
	<-c.done
}
