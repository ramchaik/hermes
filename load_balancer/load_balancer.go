package loadbalancer

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Key int

const (
	Attempts Key = iota
	Retry
)

// Run a health check every X secs
func healthCheck(sp *ServicePool) {
	t := time.NewTicker(time.Duration(sp.HealthCheckInSec) * time.Second)
	for range t.C {
		log.Printf("Starting health check...")
		sp.HealthCheck()
		log.Printf("Health Check completed")
	}
}

func Run() {
	config := ParseAndLoadConfig()
	sp := NewServicePool(config)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(sp.Handler),
	}

	// Start health check (passive)
	go healthCheck(sp)

	fmt.Printf("Load Balancer started at %d\n", config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
