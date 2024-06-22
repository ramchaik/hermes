package loadbalancer

import "sync"

type WeightedRoundRobinStrategy struct {
	services []*Service
	mu       sync.Mutex
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
