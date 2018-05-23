package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	// vendor imports
	"github.com/koding/websocketproxy"
)

// Reverse Proxy that supports both pure HTTP request and
type AnyProxy struct {
	Backend        string
	WebsocketProxy *websocketproxy.WebsocketProxy
	HttpProxy      *httputil.ReverseProxy
}

func NewWithWebsocketProxy(backend string, wsp *websocketproxy.WebsocketProxy) (*AnyProxy, error) {
	httpURL, err := url.Parse("http://" + backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse http url: %v", err)
	}

	return &AnyProxy{
		Backend:        backend,
		HttpProxy:      httputil.NewSingleHostReverseProxy(httpURL),
		WebsocketProxy: wsp,
	}, nil

}

// New allocates a new AnyProxy object
// "backend" is the backend host for the reverse proxy
func New(backend string) (*AnyProxy, error) {
	httpURL, err := url.Parse("http://" + backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse http url: %v", err)
	}

	wsURL, err := url.Parse("ws://" + backend)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse websocket url: %v", err)
	}

	return &AnyProxy{
		Backend:        backend,
		HttpProxy:      httputil.NewSingleHostReverseProxy(httpURL),
		WebsocketProxy: websocketproxy.NewProxy(wsURL),
	}, nil
}

// ServeHTTP implements the http.Handler that proxies both HTTP and WebSocket
// connections.
func (p *AnyProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Connection") == "Upgrade" || r.Header.Get("Sec-WebSocket-Protocol") != "" {
		p.WebsocketProxy.ServeHTTP(w, r)
	} else {
		p.HttpProxy.ServeHTTP(w, r)
	}
}
