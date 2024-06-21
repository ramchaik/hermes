package loadbalancer

import (
	"log"
	"sync"
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

type WeightedRoundRobinStrategy struct {
	services []*Service
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

func NewWeightedRoundRobinStrategy(services []*Service) *WeightedRoundRobinStrategy {
	return &WeightedRoundRobinStrategy{
		services: services,
	}
}

func (wrr *WeightedRoundRobinStrategy) GetNextService() *Service {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	totalWeight := 0
	maxWeightService := wrr.services[0]

	for _, service := range wrr.services {
		if !service.IsAlive() {
			continue
		}
		service.CurrentWeight += service.Weight
		totalWeight += service.Weight
		if service.CurrentWeight > maxWeightService.CurrentWeight {
			maxWeightService = service
		}
	}

	if maxWeightService == nil {
		return nil
	}

	maxWeightService.CurrentWeight -= totalWeight

	return maxWeightService
}

func (wrr *WeightedRoundRobinStrategy) UpdateServiceStats(s *Service) {
	// This method is not needed for weighted round-robin
}
