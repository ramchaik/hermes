# Hermes Load Balancer

Welcome to **Hermes**, a robust HTTP load balancer and reverse proxy written from scratch in Go. Hermes is designed to optimize your network's load distribution, ensuring smooth and reliable service delivery.

## :star: Key Features

- HTTP load balancing and reverse proxy functionality
- Multiple load distribution strategies
- Active and passive health checks for backend services
- Easy configuration via YAML file or command-line arguments
- Docker support for simple deployment

## :zap: Load Distribution Strategies

Hermes supports multiple strategies to cater to different needs and infrastructure setups:

- **Round Robin**: A fair and equal distribution method, perfect for when all backend services have similar processing capabilities.
- **Weighted Round Robin**: Tailor the load distribution based on the processing power of each backend, assigning more weight to stronger servers.
- **Least Connections**: Smartly routes traffic to the servers with the fewest active connections, minimizing response times and maximizing efficiency.

## :heart: Health Checks

Hermes ensures high availability through two types of health checks:

- **Active Checks**: During request processing, if a selected backend is unresponsive, it's immediately marked as down.
- **Passive Checks**: Regular pings are sent to backends at fixed intervals to monitor their status.

## :rocket: Getting Started

Running Hermes is straightforward, with options for configuration file, command-line arguments, or Docker deployment.

### Using a Configuration File

Ensure you have a `config.yaml` set up with your desired settings.

```sh
./hermes
```

### Using Command-Line Arguments

```sh
./hermes --help
```

### :whale: Running with Docker Compose

```sh
docker compose up
```
Hermes comes with Docker configurations for seamless deployment across any setup.
To customize Hermes' deployment, modify the docker-compose.yaml file with your preferred settings.

## :gear: Configuration

Hermes can be configured using a YAML file. Here's an explanation of the configuration options:
```yaml
port: 80 # Port on which the load balancer will be running
healthCheckInSec: 20 # Interval for passive health checks in seconds
strategy: "least-connections" # Load balancing strategy
services: # Backend services configuration
  - name: srv1
    url: http://localhost:9001
    weight: 30
  - name: srv2
    url: http://localhost:9002
    weight: 1
  - name: srv3
    url: http://localhost:9003
    weight: 1
```

### Configuration Options

- port: The port on which Hermes will listen for incoming requests.
- healthCheckInSec: The interval (in seconds) for passive health checks.
- strategy: The load balancing strategy. Possible values are "round-robin", "weighted-round-robin", or "least-connections".
- services: A list of backend services.
    - name: A unique identifier for the service.
    - url: The URL of the backend service.
    - weight: The weight assigned to the service (required for "weighted-round-robin" strategy).

### :sparkles: Why Hermes?

- Efficiency: Optimized for minimal latency and maximum throughput.
- Flexibility: Supports multiple load distribution strategies.
- Reliability: Active and passive health checks ensure high availability.
- Ease of Use: Simple setup with Docker support for hassle-free deployment.

### :construction: TODO

- [ ] Implement a heap for sorting alive backends to reduce search complexity

### :clap: Acknowledgements

This project stands on the shoulders of giants. A heartfelt thank you to the following resources and communities:

[Let's Create a Simple Load Balancer With Go](https://kasvith.me/posts/lets-create-a-simple-lb-go/) - This insightful blog post was the spark that ignited the creation of Hermes. It provided foundational knowledge and inspiration.

### :handshake: Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### :memo: License

This project is licensed under the MIT License.
