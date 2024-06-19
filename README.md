# Hermes

A simple load balancer written in Golang.

Uses Round Robin for load distribution.

## Running the load balancer

```sh
./hermes \
    --port 9000 \
    --services http://localhost:9001,http://localhost:9002,http://localhost:9003 

```

## TODO:

- [ ] Implement weighted round-robin/least connections

>Round Robin - Distribute load equally, assumes all backends have the same processing power

> Weighted Round Robin - Additional weights can be given considering the backend's processing power

> Least Connections - Load is distributed to the servers with least active connections

- [x] Add configuration file support
- [ ] Use a heap for sort out alive backends to reduce search surface
- [ ] Collect statistics
