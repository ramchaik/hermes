package loadbalancer

import (
	"net/http/httputil"
	"net/url"
	"strconv"
	"testing"
)

// TestWeightedRoundRobinDistribution tests the distribution of requests among services based on their weights
func TestWeightedRoundRobinDistribution(t *testing.T) {
	// Assuming each Service has a Count field or you're tracking counts externally
	// Initialize the count map
	serviceCounts := make(map[*Service]int)

	services := make([]*Service, 3)
	for i := range services {
		serviceURL, _ := url.Parse("http://service" + strconv.Itoa(i+1) + ".com")
		services[i] = NewService(
			serviceURL, // URL
			true,       // Active
			i+1,        // Weight
			httputil.NewSingleHostReverseProxy(serviceURL), // Reverse Proxy
		)
	}

	wrr := NewWeightedRoundRobinStrategy(services)

	// Simulate 600 requests to check distribution
	for i := 0; i < 600; i++ {
		selectedService := wrr.GetNextService()
		serviceCounts[selectedService]++
	}

	// Check if the distribution matches the weights approximately
	// Since we have weights 1, 2, and 3, the distribution should be close to 1:2:3
	totalRequests := 0
	for _, service := range services {
		totalRequests += serviceCounts[service]
	}

	for _, service := range services {
		expectedDistribution := float64(service.Weight) / float64(1+2+3) // Calculate expected distribution
		actualDistribution := float64(serviceCounts[service]) / float64(totalRequests)

		if actualDistribution < expectedDistribution-0.1 || actualDistribution > expectedDistribution+0.1 {
			t.Errorf("Distribution for service with weight %d is not as expected. Expected around %f, got %f", service.Weight, expectedDistribution, actualDistribution)
		}
	}
}
