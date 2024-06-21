package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"strconv"
	"testing"
)

// TestRoundRobinLoadDistribution tests that the round-robin strategy
// distributes load equally among 3 services.
func TestRoundRobinLoadDistribution(t *testing.T) {
	// Initialize 3 mock services
	services := make([]*Service, 3)
	for i := range services {
		serviceURL, _ := url.Parse("http://service" + strconv.Itoa(i+1) + ".com")
		services[i] = NewService(serviceURL, true, httputil.NewSingleHostReverseProxy(serviceURL))
	}

	// Initialize the round-robin strategy with the services
	rr := NewRoundRobinStrategy(services)

	// Simulate service selection
	serviceCount := make(map[string]int)
	totalRequests := 300
	for i := 0; i < totalRequests; i++ {
		selectedService := rr.GetNextService()
		serviceCount[selectedService.URL.String()]++
	}

	// Verify equal distribution
	expectedCountPerService := totalRequests / len(services)
	for _, service := range services {
		count := serviceCount[service.URL.String()]
		if count != expectedCountPerService {
			t.Errorf("Service %s was selected %d times; expected %d", service.URL.String(), count, expectedCountPerService)
		}
	}
}
