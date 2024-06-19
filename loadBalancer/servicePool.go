package loadbalancer

import (
	"log"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

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
