# Go Distributed Loggers Demo
Simple Go + Consul Distributed System  

## Components
- **consul**: Consul instance that provides support for leader election and service discovery
- **distributed-logger**: The Distributed Logger nodes expose a REST API that logs received messages to Stdout. Only the cluster leader accepts messages at any given time. A new node takes over in case of leader node failure.
- **producer**: The producer queries Consul periodically to determine the distributed-logger leader and sends it a numbered message.

## Demo instructions
Start services
```
docker-compose up -d --scale distributed-logger=3
```
Tail logs 
```
docker-compose logs -f distributed-logger
```