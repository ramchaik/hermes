package loadbalancer

import (
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ServiceConfig struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

type Config struct {
	Port     int             `yaml:"port"`
	Services []ServiceConfig `yaml:"services"`
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

func ParseAndLoadConfig(setupFn func(surl *url.URL)) *Config {
	var configFile string
	var serverList string
	var port int

	flag.StringVar(&configFile, "file", "", "YAML Config file for load balance")
	flag.StringVar(&serverList, "services", "", "Load balanced services, use comma separated list")
	flag.IntVar(&port, "port", 9000, "Port to serve")
	flag.Parse()

	config := loadFileConfig(configFile)

	// Load values from CLI if not found on config file
	if len(config.Services) == 0 && serverList != "" {
		tokens := strings.Split(serverList, ",")
		for _, t := range tokens {
			cs := ServiceConfig{
				URL:  t,
				Name: t,
			}
			config.Services = append(config.Services, cs)
		}
	}

	if config.Port == 0 && port > 0 {
		config.Port = port
	}

	if len(config.Services) == 0 {
		log.Fatal("Please provide services to load balance")
	}

	for _, s := range config.Services {
		serviceUrl, err := url.Parse(s.URL)
		if err != nil {
			log.Fatal(err)
		}
		setupFn(serviceUrl)
	}

	return config
}
