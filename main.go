package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Key int

const (
	Attempts Key = iota
	Retry
)

type Service struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type ServicePool struct {
	services []*Service
	current  uint64
}

func (sp *ServicePool) NextIndex() int {
	if len(sp.services) == 0 {
		log.Fatal("No services attached")
	}
	return int(atomic.AddUint64(&sp.current, uint64(1)) % uint64(len(sp.services)))
}

func (sp *ServicePool) GetNextPeer() *Service {
	next := sp.NextIndex()
	l := len(sp.services) + next
	for i := next; i < l; i++ {
		idx := i % len(sp.services)
		if sp.services[idx].Alive {
			if i != next {
				atomic.StoreUint64(&sp.current, uint64(idx))
			}
			return sp.services[idx]
		}
	}
	return nil
}

func (sp *ServicePool) MarkBackendStatus(serviceUrl *url.URL, alive bool) {
	for _, s := range sp.services {
		if s.URL.String() == serviceUrl.String() {
			s.SetAlive(alive)
			break
		}
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

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// isBackendAlive checks whether a service is Alive by establishing a TCP connection
func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Printf("Site unreachable; Error: %v\n", err)
		return false
	}
	_ = conn.Close()
	return true
}

func (sp *ServicePool) HealthCheck() {
	for _, s := range sp.services {
		status := "up"

		alive := isBackendAlive(s.URL)
		s.SetAlive(alive)

		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]", s.URL, status)
	}
}

func (sp *ServicePool) AddService(service *Service) {
	sp.services = append(sp.services, service)
}

// Run a health check every 20secs
func healthCheck() {
	t := time.NewTicker(20 * time.Second)
	for range t.C {
		log.Printf("Starting health check...")
		serverPool.HealthCheck()
		log.Printf("Health Check completed")
	}
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

var serverPool ServicePool

func main() {
	var serverList string
	var port int

	flag.StringVar(&serverList, "services", "", "Load balanced services, use comma separated list")
	flag.IntVar(&port, "port", 9000, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide services to load balance")
	}

	tokens := strings.Split(serverList, ",")
	for _, t := range tokens {
		serviceUrl, err := url.Parse(t)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serviceUrl)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[%s] %s\n", serviceUrl.Host, err.Error())
			retries := GetRetryFromContext(r)
			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(r.Context(), Retry, retries+1)
				proxy.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// After 3 retires, mark service as down
			serverPool.MarkBackendStatus(serviceUrl, false)

			attempts := GetAttemptsFromContext(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
			lb(w, r.WithContext(ctx))
		}

		serverPool.AddService(&Service{
			URL:          serviceUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})
		log.Printf("Configured service: %s\n", serviceUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	// Start health check
	go healthCheck()

	fmt.Printf("Load Balancer started at %d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
