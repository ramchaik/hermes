package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Service struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func NewService(url *url.URL, alive bool, proxy *httputil.ReverseProxy) *Service {
	return &Service{
		URL:          url,
		Alive:        alive,
		ReverseProxy: proxy,
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
