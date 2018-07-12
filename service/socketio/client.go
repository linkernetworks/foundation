package socketio

import (
	"time"

	"github.com/linkernetworks/logger"

	socketio "github.com/c9s/go-socket.io"
	redigo "github.com/gomodule/redigo/redis"
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
			logger.Debugf("[socketio] subscription: kind=%s channel=%s count=%d", v.Kind, v.Channel, v.Count)
			if v.Count == 0 {
				logger.Debugf("[socketio] subscription count reached 0, exiting")
				break PIPE
			}
		case redigo.Message:
			c.C <- string(v.Data)
		// when the connection is closed, redigo returns an error "connection closed" here
		case error:
			logger.Warnf("[socketio] redis reader error=%v", v)
			break PIPE
		}
	}
	close(c.C)
	close(c.done)
}

// emit chan message to socket event
func (c *Client) write() {
	for msg := range c.C {
		if c.Socket != nil {
			if err := c.Socket.Emit(c.toEvent, msg); err != nil {
				logger.Errorf("[socketio] redis writer: event '%s' emit error: %v", c.toEvent, err)
			}
		}
	}
}

func (c *Client) Start() {
	c.done = make(chan bool)
	c.C = make(chan string, c.BufSize)

	// subscribe to system:events by default so the connection keeps at least
	// one subscription to prevent the read loop exits.
	// the loop won't exit if there is no topic subscribed yet
	c.Subscribe(":magic:")

	// open the redis connection and start reading before we start the go
	// routine for writing.
	go c.read() // from redigo to channel

	// the write method is blocking, it will exit when the channel is closed.
	go c.write() // from channel to socket event
}

func (c *Client) Stop() error {
	// wait for the reader exits, so that we can safely close the connection
	defer func() { <-c.done }()
	// Unsubscribe all channels
	return c.PubSubConn.Unsubscribe()
}
