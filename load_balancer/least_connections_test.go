package loadbalancer

import (
	"net/url"
	"strconv"
	"sync"
	"testing"
)

// MockService for testing
type MockService struct {
	*Service
}

func NewMockService(activeConn int) *MockService {
	serviceURL, _ := url.Parse("http://service" + strconv.Itoa(activeConn) + ".com")
	return &MockService{
		Service: &Service{
			URL:        serviceURL,
			Alive:      true,
			ActiveConn: activeConn,
			mux:        sync.RWMutex{},
		},
	}
}

func TestLeastConnectionsStrategy(t *testing.T) {
	// Setup: Create mock services with different numbers of active connections
	services := []*Service{
		NewMockService(10).Service, // 10 active connections
		NewMockService(5).Service,  // 5 active connections, should be selected first
		NewMockService(7).Service,  // 7 active connections
	}

	// Instantiate LeastConnectionsStrategy with the mock services
	lcs := NewLeastConnectionsStrategy(services)

	// Invoke GetNextService and assert it returns the service with the least connections
	selectedService := lcs.GetNextService()
	if selectedService != services[1] {
		t.Errorf("Expected service with 5 active connections to be selected, got %d", selectedService.ActiveConn)
	}

	// Simulate a scenario where the service with the least connections changes
	services[1].ActiveConn = 12 // Now, this service has the most connections
	services[2].ActiveConn = 4  // This should be selected next

	selectedService = lcs.GetNextService()
	if selectedService != services[2] {
		t.Errorf("Expected service with 4 active connections to be selected, got %d", selectedService.ActiveConn)
	}
}
