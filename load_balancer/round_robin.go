package loadbalancer

import (
	"log"
	"sync"
	"sync/atomic"
)

type RoundRobinStrategy struct {
	services []*Service
	current  uint64
	mu       sync.Mutex
}

func NewRoundRobinStrategy(services []*Service) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		services: services,
	}
}

func (rr *RoundRobinStrategy) nextIndex() int {
	if len(rr.services) == 0 {
		log.Fatal("No services attached")
	}
	return int(atomic.AddUint64(&rr.current, uint64(1)) % uint64(len(rr.services)))
}

func (rr *RoundRobinStrategy) GetNextService() *Service {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	next := rr.nextIndex()
	l := len(rr.services) + next
	for i := next; i < l; i++ {
		idx := i % len(rr.services)
		if rr.services[idx].Alive {
			if i != next {
				atomic.StoreUint64(&rr.current, uint64(idx))
			}
			return rr.services[idx]
		}
	}
	return nil
}

func (rr *RoundRobinStrategy) UpdateServiceStats(service *Service) {
	// This method is not needed for round-robin
}
