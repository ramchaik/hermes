package loadbalancer

import (
	"log"
	"net"
	"net/http/httputil"
	"net/url"
	"time"
)

type ServicePool struct {
	services []*Service
	strategy Strategy
}

func setupService(sp *ServicePool, serviceUrl *url.URL) error {
	proxy := httputil.NewSingleHostReverseProxy(serviceUrl)
	proxy.ErrorHandler = getProxyErrorHandler(proxy, serviceUrl)
	service := NewService(serviceUrl, true, proxy)
	sp.AddService(service)
	log.Printf("Configured service: %s\n", serviceUrl)
	return nil
}

func SetupServicePool(config *Config) *ServicePool {
	servicePool := &ServicePool{
		services: make([]*Service, 0, len(config.Services)),
		strategy: nil,
	}

	for _, serviceConfig := range config.Services {
		serviceUrl, err := url.Parse(serviceConfig.URL)
		if err != nil {
			log.Fatal(err)
		}
		setupService(servicePool, serviceUrl)
	}

	// Now set up the strategy
	switch config.Strategy {
	case RoundRobin:
		servicePool.strategy = NewRoundRobinStrategy(serverPool.services)
	default:
		log.Fatalf("Invalid strategy: %s", config.Strategy)
	}

	log.Printf("Load distribution Strategy: [%s]", config.Strategy)

	return servicePool
}

func (sp *ServicePool) GetNextPeer() *Service {
	return sp.strategy.GetNextService()
}

func (sp *ServicePool) UpdateServiceStats(s *Service) {
	sp.strategy.UpdateServiceStats(s)
}

func (sp *ServicePool) MarkBackendStatus(serviceUrl *url.URL, alive bool) {
	for _, s := range sp.services {
		if s.URL.String() == serviceUrl.String() {
			s.SetAlive(alive)
			break
		}
	}
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
