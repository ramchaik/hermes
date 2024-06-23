package loadbalancer

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type ServicePool struct {
	Services         []*Service
	Strategy         Strategy
	Handler          func(w http.ResponseWriter, r *http.Request)
	HealthCheckInSec int
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

func getProxyErrorHandler(sp *ServicePool, proxy *httputil.ReverseProxy, serviceUrl *url.URL) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[%s] %s\n", serviceUrl.Host, err.Error())
		retries := GetRetryFromContext(r)
		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), Retry, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// After 3 retires, mark service as down (Active)
		sp.MarkBackendStatus(serviceUrl, false)

		attempts := GetAttemptsFromContext(r)
		log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		sp.Handler(w, r.WithContext(ctx))

		// Decrement active connection count
		service := sp.GetServiceByURL(serviceUrl)
		if service != nil {
			service.mux.Lock()
			service.ActiveConn--
			service.mux.Unlock()
		}
	}
}

func (sp *ServicePool) GetServiceByURL(serviceUrl *url.URL) *Service {
	for _, service := range sp.Services {
		if service.URL.String() == serviceUrl.String() {
			return service
		}
	}
	return nil
}

func setupService(sp *ServicePool, sw int, serviceUrl *url.URL) *Service {
	proxy := httputil.NewSingleHostReverseProxy(serviceUrl)
	proxy.ErrorHandler = getProxyErrorHandler(sp, proxy, serviceUrl)
	service := NewService(serviceUrl, true, sw, proxy)
	sp.AddService(service)
	log.Printf("Configured service: %s\n", serviceUrl)
	return service
}

func (sp *ServicePool) initHandler() {
	sp.Handler = func(w http.ResponseWriter, r *http.Request) {
		attempts := GetAttemptsFromContext(r)
		if attempts > 3 {
			log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Service not available", http.StatusServiceUnavailable)
			return
		}

		peer := sp.GetNextPeer()
		if peer != nil {
			peer.ReverseProxy.ServeHTTP(w, r)
			return
		}
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}
}

func NewServicePool(config *Config) *ServicePool {
	sp := &ServicePool{
		Services: make([]*Service, 0, len(config.Services)),
		Strategy: nil,
	}

	for _, serviceConfig := range config.Services {
		serviceUrl, err := url.Parse(serviceConfig.URL)
		if err != nil {
			log.Fatal(err)
		}
		setupService(sp, serviceConfig.Weight, serviceUrl)
	}

	// Now set up the strategy
	switch config.Strategy {
	case RoundRobin:
		sp.Strategy = NewRoundRobinStrategy(sp.Services)
	case WeightedRoundRobin:
		sp.Strategy = NewWeightedRoundRobinStrategy(sp.Services)
	case LeastConnections:
		sp.Strategy = NewLeastConnectionsStrategy(sp.Services)

	default:
		log.Fatalf("Invalid strategy: %s", config.Strategy)
	}

	log.Printf("Load distribution Strategy: [%s]", config.Strategy)

	// Set Health check in seconds for service pool
	sp.HealthCheckInSec = config.HealthCheckInSec

	// Create the LB Handler
	sp.initHandler()

	return sp
}

func (sp *ServicePool) GetNextPeer() *Service {
	return sp.Strategy.GetNextService()
}

func (sp *ServicePool) UpdateServiceStats(s *Service) {
	sp.Strategy.UpdateServiceStats(s)
}

func (sp *ServicePool) MarkBackendStatus(serviceUrl *url.URL, alive bool) {
	for _, s := range sp.Services {
		if s.URL.String() == serviceUrl.String() {
			s.SetAlive(alive)
			break
		}
	}
}

func (sp *ServicePool) HealthCheck() {
	for _, s := range sp.Services {
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
	sp.Services = append(sp.Services, service)
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
