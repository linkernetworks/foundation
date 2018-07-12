package netutils

import (
	"context"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckNetworkSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	idleConnsClosed := make(chan struct{})

	rand.Seed(time.Now().UTC().UnixNano())
	port := int(rand.Int31n(30000)) + 30000 // 30000 - 60000

	host := net.JoinHostPort("localhost", strconv.Itoa(port))
	srv := &http.Server{Addr: host}

	go func() {
		err := srv.ListenAndServe()
		assert.Equal(t, err, http.ErrServerClosed)
		close(idleConnsClosed)
	}()
	err := CheckNetworkConnectivity("127.0.0.1", port, "tcp", 10)
	assert.NoError(t, err)

	defer cancel()
	srv.Shutdown(ctx)
	//Wait the server closed
	<-idleConnsClosed
}

func TestCheckNetworkFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	idleConnsClosed := make(chan struct{})

	rand.Seed(time.Now().UTC().UnixNano())
	port := int(rand.Int31n(30000)) + 30000 // 30000 - 60000

	host := net.JoinHostPort("localhost", strconv.Itoa(port))
	srv := &http.Server{Addr: host}

	go func() {
		err := srv.ListenAndServe()
		assert.Equal(t, http.ErrServerClosed, err)
		close(idleConnsClosed)
	}()
	err := CheckNetworkConnectivity("127.0.0.1", port-1, "tcp", 3)
	assert.Error(t, err)

	defer cancel()
	srv.Shutdown(ctx)
	//Wait the server closed
	<-idleConnsClosed
}
