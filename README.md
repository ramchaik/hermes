# Hermes Load Balancer

Welcome to **Hermes**, a sleek and efficient load balancer in Golang. Hermes is designed to optimize your network's load distribution, ensuring smooth and reliable service delivery.

## :zap: Load Distribution Strategies

Hermes is versatile, supporting multiple strategies to cater to different needs and infrastructure setups:

- **Round Robin**: A fair and equal distribution method, perfect for when all backend services boast similar processing capabilities.

- **Weighted Round Robin**: Tailor the load distribution based on the processing power of each backend, assigning more weight to stronger servers.

- **Least Connections**: Smartly routes traffic to the servers with the fewest active connections, minimizing response times and maximizing efficiency.

## :rocket: Getting Started

Running Hermes is a breeze, whether you prefer using a `config.yaml` or command-line arguments. Plus, with Docker support, deployment is as simple as pulling an image and running a container.

### Using a Configuration File

Ensure you have a `config.yaml` set up with your desired settings.

```sh
./hermes
```

### Using Command-Line Arguments

```sh
./hermes --help
```

### :whale: Running with Docker

```sh
docker pull hermes/loadbalancer:latest
docker run -d -p 80:80 hermes/loadbalancer:latest
```

Hermes is also available as a Docker image, making it even easier to deploy in any environment.

This command pulls the latest Hermes Docker image and runs it as a detached process, mapping the container's port 80 to the host's port 80.

### :sparkles: Why Hermes?

- Efficiency: Optimized for minimal latency and maximum throughput.
- Flexibility: Supports multiple load distribution strategies.
- Ease of Use: Simple setup with Docker support for hassle-free deployment.

Dive into Hermes and experience the next level of load balancing!

## TODO

- [ ] Use a heap for sort out alive backends to reduce search surface
- [ ] Collect statistics
