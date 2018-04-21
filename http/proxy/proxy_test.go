package proxy

import (
	"bytes"
	"context"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {

	var wg = sync.WaitGroup{}

	// create the backend server (1)
	backendServer1 := newWSEchoServer("1:")
	_, proxy1 := startServerWithRandomPort(t, backendServer1, "127.0.0.1", &wg)
	defer stopServerGracefully(t, backendServer1)

	backendServer2 := newWSEchoServer("2:")
	_, proxy2 := startServerWithRandomPort(t, backendServer2, "127.0.0.1", &wg)
	defer stopServerGracefully(t, backendServer2)

	proxyMap := NewProxyMap()
	proxyMap.Add("1", proxy1)
	proxyMap.Add("2", proxy2)

	router := mux.NewRouter()
	router.HandleFunc("/proxy/{id}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		if proxy, ok := proxyMap.Get(id); ok {
			proxy.ServeHTTP(w, r)
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	frontendServer := &http.Server{
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	port, _ := startServerWithRandomPort(t, frontendServer, "127.0.0.1", &wg)
	defer stopServerGracefully(t, frontendServer)

	var frontendURL = "ws://127.0.0.1:" + strconv.Itoa(port)
	t.Logf("frontend url: %s", frontendURL)

	// wait for the servers get ready
	wg.Wait()

	// let's us define two subprotocols, only one is supported by the server
	clientSubProtocols := []string{"test-protocol", "test-notsupported"}
	h := http.Header{}
	for _, subprot := range clientSubProtocols {
		h.Add("Sec-WebSocket-Protocol", subprot)
	}

	for _, id := range []string{"1", "2"} {
		// frontend server, dial now our proxy, which will reverse proxy our
		// message to the backend websocket server.
		wsc, _, err := websocket.DefaultDialer.Dial(frontendURL+"/proxy/"+id, h)
		assert.NoError(t, err)

		// write the message to the websocket connection
		msg := "hello"
		err = wsc.WriteMessage(websocket.TextMessage, []byte(msg))
		assert.NoError(t, err)

		// read the message from the websocket connection
		messageType, p, err := wsc.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, websocket.TextMessage, messageType)

		// check the response message
		assert.Equal(t, id+":"+msg, string(p))
	}
}

func stopServerGracefully(t *testing.T, svr *http.Server) {
	t.Logf("stopping server...")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	svr.Shutdown(ctx)
	t.Logf("server shutted down.")
}

// start the http server with a random unused port. please note the "address"
// here must be a bind address without port number. e.g., "127.0.0.1"
func startServerWithRandomPort(t *testing.T, svr *http.Server, address string, wg *sync.WaitGroup) (int, *AnyProxy) {
	// start the server and wait for the listener get ready
	var ln = startServer(t, svr, ":0", wg)
	var port = ln.Addr().(*net.TCPAddr).Port
	assert.True(t, port > 0)

	// create a proxy for the backend server. pass the address for both http
	// and ws, and please note that it requires the hostname.
	proxy, err := New(address + ":" + strconv.Itoa(port))
	assert.NoError(t, err)
	return port, proxy
}

func startServer(t *testing.T, svr *http.Server, address string, wg *sync.WaitGroup) net.Listener {
	var ln, err = net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	wg.Add(1)
	go func() {
		wg.Done()
		svr.Serve(ln)
	}()
	return ln
}

func newWSEchoServer(prefix string) *http.Server {
	upgrader := websocketproxy.DefaultUpgrader

	// start the backend server
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wsc, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade http request:", err)
			return
		}

		messageType, p, err := wsc.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			return
		}

		payload := bytes.NewBufferString(prefix)
		_, err = payload.Write(p)
		if err != nil {
			log.Println("Failed to write to the message buffer")
		}

		if err = wsc.WriteMessage(messageType, payload.Bytes()); err != nil {
			log.Println("Failed to write message:", err)
			return
		}
	})

	return &http.Server{
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
