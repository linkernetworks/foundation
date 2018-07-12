package proxy

import "sync"

type ReverseProxyMap struct {
	sync.RWMutex
	proxies map[string]*AnyProxy
}

func NewProxyMap() *ReverseProxyMap {
	return &ReverseProxyMap{
		proxies: map[string]*AnyProxy{},
	}
}

func (p *ReverseProxyMap) Add(id string, proxy *AnyProxy) {
	p.Lock()
	p.proxies[id] = proxy
	p.Unlock()
}

func (m *ReverseProxyMap) Remove(id string) {
	m.Lock()
	delete(m.proxies, id)
	m.Unlock()
}

func (m *ReverseProxyMap) Get(id string) (p *AnyProxy, ok bool) {
	p, ok = m.proxies[id]
	return p, ok
}
