package loadbalancer

import "sync"

type LeastConnectionsStrategy struct {
	services []*Service
	mu       sync.Mutex
}

func NewLeastConnectionsStrategy(services []*Service) *LeastConnectionsStrategy {
	return &LeastConnectionsStrategy{
		services: services,
	}
}

func (lc *LeastConnectionsStrategy) GetNextService() *Service {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var selectedService *Service
	minConnections := int(^uint(0) >> 1) // Initialize to max int value

	for _, service := range lc.services {
		if service.IsAlive() && service.ActiveConn < minConnections {
			selectedService = service
			minConnections = service.ActiveConn
		}
	}

	if selectedService != nil {
		selectedService.mux.Lock()
		selectedService.ActiveConn++
		selectedService.mux.Unlock()
	}

	return selectedService
}

func (lc *LeastConnectionsStrategy) UpdateServiceStats(service *Service) {
	service.mux.Lock()
	service.ActiveConn--
	service.mux.Unlock()
}
