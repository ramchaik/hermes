package loadbalancer

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type Key int

const (
	Attempts Key = iota
	Retry
)

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

// Run a health check every 20secs
func healthCheck() {
	t := time.NewTicker(20 * time.Second)
	for range t.C {
		log.Printf("Starting health check...")
		serverPool.HealthCheck()
		log.Printf("Health Check completed")
	}
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

var serverPool ServicePool

func getProxyErrorHandler(proxy *httputil.ReverseProxy, serviceUrl *url.URL) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[%s] %s\n", serviceUrl.Host, err.Error())
		retries := GetRetryFromContext(r)
		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), Retry, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// After 3 retires, mark service as down
		serverPool.MarkBackendStatus(serviceUrl, false)

		attempts := GetAttemptsFromContext(r)
		log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		lb(w, r.WithContext(ctx))
	}
}

func setupService(serviceUrl *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(serviceUrl)
	proxy.ErrorHandler = getProxyErrorHandler(proxy, serviceUrl)
	service := NewService(serviceUrl, true, proxy)
	serverPool.AddService(service)
	log.Printf("Configured service: %s\n", serviceUrl)
}

func parseAndLoadConfig(serverList string, port *int, setupFn func(surl *url.URL)) {
	flag.StringVar(&serverList, "services", "", "Load balanced services, use comma separated list")
	flag.IntVar(&*port, "port", 9000, "Port to serve")
	flag.Parse()

	if len(serverList) == 0 {
		log.Fatal("Please provide services to load balance")
	}

	tokens := strings.Split(serverList, ",")
	for _, t := range tokens {
		serviceUrl, err := url.Parse(t)
		if err != nil {
			log.Fatal(err)
		}
		setupFn(serviceUrl)
	}
}

func Run() {
	var serverList string
	var port int

	parseAndLoadConfig(serverList, &port, setupService)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	// Start health check
	go healthCheck()

	fmt.Printf("Load Balancer started at %d\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
