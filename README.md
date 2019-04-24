# Go Distributed Loggers Demo
Simple Go + Consul Distributed System

## Demo instructions
Start services
```
docker-compose up -d --scale distributed-logger=3
```
Tail logs 
```
docker-compose logs -f distributed-logger
```