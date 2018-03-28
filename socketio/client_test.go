package socketio

import (
	"net/http"
	"testing"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/service/redis"

	"github.com/stretchr/testify/assert"
)

type FakeSocket struct {
	C     chan *FakeSocketMessage
	rooms []string
}

type FakeSocketMessage struct {
	Event    string
	Messages []string
}

func (s *FakeSocket) Emit(event string, args ...interface{}) error {
	msgs := []string{}
	for _, arg := range args {
		msgs = append(msgs, arg.(string))
	}
	s.C <- &FakeSocketMessage{Event: event, Messages: msgs}
	return nil
}

func (s *FakeSocket) BroadcastTo(room string, event string, args ...interface{}) error { return nil }
func (s *FakeSocket) On(event string, f interface{}) error                             { return nil }
func (s *FakeSocket) Id() string                                                       { return "fake-socket" }
func (s *FakeSocket) Rooms() []string                                                  { return s.rooms }
func (s *FakeSocket) Request() *http.Request                                           { return nil }

func (s *FakeSocket) Join(room string) error {
	s.rooms = append(s.rooms, room)
	return nil
}

func (s *FakeSocket) Leave(removal string) error {
	rooms := []string{}
	for _, room := range s.rooms {
		if room != removal {
			rooms = append(rooms, room)
		}
	}
	s.rooms = rooms
	return nil
}

func (s *FakeSocket) Disconnect() {}

func TestStream(t *testing.T) {
	cf := config.MustRead("../../../config/testing.json")
	r := redis.New(cf.Redis)
	conn := r.GetConnection()
	defer conn.Close()

	psc := conn.PubSub()

	socket := FakeSocket{C: make(chan *FakeSocketMessage, 100)}

	client := &Client{
		PubSubConn: psc.PubSubConn,
		Socket:     &socket,
		BufSize:    100,
		ExpiredAt:  time.Now().Add(10 * time.Minute),
	}
	client.Subscribe("_test_socket_1_")
	client.Subscribe("_test_socket_2_")
	client.Start()

	go func() {
		var err error

		c2 := r.GetConnection()
		_, err = c2.Do("PUBLISH", "_test_socket_1_", "message1")
		assert.NoError(t, err)

		err = c2.Flush()
		assert.NoError(t, err)

		_, err = c2.Do("PUBLISH", "_test_socket_2_", "message2")
		assert.NoError(t, err)

		err = c2.Flush()
		assert.NoError(t, err)
	}()

	m1 := <-socket.C
	assert.Equal(t, "message1", m1.Messages[0])

	m2 := <-socket.C
	assert.Equal(t, "message2", m2.Messages[0])

	client.Stop()
}
