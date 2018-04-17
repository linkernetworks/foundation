package netutils

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCheckNetworkSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	idleConnsClosed := make(chan struct{})
	srv := &http.Server{Addr: "localhost:8080"}

	go func() {
		err := srv.ListenAndServe()
		assert.Equal(t, err, http.ErrServerClosed)
		close(idleConnsClosed)
	}()
	err := CheckNetworkConnectivity("127.0.0.1", 8080, "tcp", 10)
	assert.NoError(t, err)

	defer cancel()
	srv.Shutdown(ctx)
	//Wait the server closed
	<-idleConnsClosed
}

func TestCheckNetworkFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	idleConnsClosed := make(chan struct{})
	srv := &http.Server{Addr: "localhost:8081"}

	go func() {
		err := srv.ListenAndServe()
		assert.Equal(t, err, http.ErrServerClosed)
		close(idleConnsClosed)
	}()
	err := CheckNetworkConnectivity("127.0.0.1", 8082, "tcp", 3)
	assert.Error(t, err)

	defer cancel()
	srv.Shutdown(ctx)
	//Wait the server closed
	<-idleConnsClosed
}
