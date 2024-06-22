package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Service struct {
	URL           *url.URL
	Alive         bool
	Weight        int
	CurrentWeight int
	ActiveConn    int
	mux           sync.RWMutex
	ReverseProxy  *httputil.ReverseProxy
}

func NewService(url *url.URL, alive bool, weight int, proxy *httputil.ReverseProxy) *Service {
	return &Service{
		URL:           url,
		Alive:         alive,
		ReverseProxy:  proxy,
		Weight:        weight,
		CurrentWeight: 0,
		ActiveConn:    0,
	}
}

func (s *Service) SetAlive(alive bool) {
	s.mux.Lock()
	s.Alive = alive
	s.mux.Unlock()
}

func (s *Service) IsAlive() (alive bool) {
	s.mux.RLock()
	alive = s.Alive
	s.mux.RUnlock()
	return
}
