package loadbalancer

import (
	"log"
	"sync/atomic"
)

type Strategy interface {
	GetNextService() *Service
	UpdateServiceStats(s *Service)
}

type RoundRobinStrategy struct {
	services []*Service
	current  uint64
}

func NewRoundRobinStrategy(services []*Service) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		services: services,
	}
}

func (s *RoundRobinStrategy) nextIndex() int {
	if len(s.services) == 0 {
		log.Fatal("No services attached")
	}
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.services)))
}

func (s *RoundRobinStrategy) GetNextService() *Service {
	next := s.nextIndex()
	l := len(s.services) + next
	for i := next; i < l; i++ {
		idx := i % len(s.services)
		if s.services[idx].Alive {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.services[idx]
		}
	}
	return nil
}

func (s *RoundRobinStrategy) UpdateServiceStats(service *Service) {}
