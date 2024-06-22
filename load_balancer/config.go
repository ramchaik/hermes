package loadbalancer

import (
	"flag"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	URL    string `yaml:"url"`
	Name   string `yaml:"name"`
	Weight int    `yaml:"weight"`
}

type StrategyConfig string

const (
	RoundRobin         StrategyConfig = "round-robin"
	WeightedRoundRobin StrategyConfig = "weighted-round-robin"
	LeastConnections   StrategyConfig = "least-connections"
)

type Config struct {
	Port             int             `yaml:"port"`
	Services         []ServiceConfig `yaml:"services"`
	Strategy         StrategyConfig  `yaml:"strategy"`
	HealthCheckInSec int             `yaml:"healthCheckInSec"`
}

func loadFileConfig(filename string) *Config {
	if filename == "" {
		filename = "config.yaml"
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return &Config{}
	}

	var config Config

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	return &config
}

func ParseAndLoadConfig() *Config {
	var configFile string
	var serverList string
	var strategy string
	var healthCheckInSec int
	var port int

	flag.StringVar(&configFile, "file", "config.yaml", "YAML Config file for load balance")
	flag.StringVar(&serverList, "services", "", "Load balanced services, use comma separated list")
	flag.IntVar(&port, "port", 80, "Port to serve")
	flag.StringVar(&strategy, "strategy", string(RoundRobin), "Strategy for load distribution")
	flag.IntVar(&healthCheckInSec, "healthCheckInSec", 20, "Do a health check for each service every X seconds")
	flag.Parse()

	config := loadFileConfig(configFile)

	// Load values from CLI if not found on config file
	if len(config.Services) == 0 && serverList != "" {
		tokens := strings.Split(serverList, ",")
		for _, t := range tokens {
			cs := ServiceConfig{
				URL:    t,
				Name:   t,
				Weight: 1, // Default
			}
			config.Services = append(config.Services, cs)
		}
	}

	// Set default weight for each service
	for i := range config.Services {
		if config.Services[i].Weight == 0 {
			config.Services[i].Weight = 1
		}
	}

	if config.Port == 0 && port > 0 {
		config.Port = port
	}

	// Default strategy round robin
	if config.Strategy == "" {
		config.Strategy = RoundRobin
	}

	if len(config.Services) == 0 {
		log.Fatal("Please provide services to load balance")
	}

	return config
}
