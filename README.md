# Microservices Messenger

Go backend: auth, chat, notification, API gateway.

Stack: chi, JWT, gRPC, Kafka, Redis, PostgreSQL, Prometheus, Grafana, Docker Compose.

```bash
docker compose up --build
```

Gateway: http://localhost:9000 
Swagger: http://localhost:9080/swagger · http://localhost:9081/swagger 
Grafana: http://localhost:3000 (admin/admin) 

```bash
make test
make lint       
```
