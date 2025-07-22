# Unified POC: WebSocket (Gorilla), gRPC, Kafka, Postgres, Redis

## Overview

This project demonstrates a unified event-driven architecture using Go microservices, Kafka, Postgres, Redis, and gRPC. It combines multiple POCs into a single, modern, containerized system.

---

## Logic Flow

1. WebSocket client sends a message:
   ```json
   {
     "user_id": "123",
     "amount": 49.99
   }
   ```
2. websocket/ service (Gorilla WebSocket):
   - Receives the message
   - Saves the purchase to PostgreSQL
   - Publishes a Kafka event to the `user_purchase` topic
3. grpc_kafka/wallet-service/:
   - Consumes the Kafka event
   - Aggregates the total amount spent per user
   - Stores the total in Redis
   - Exposes a gRPC endpoint `GetBalance(user_id)` (returns total spent)

---

## Running the System

### 1. Start Infrastructure (Docker Compose)

From the `docker/` directory:
```bash
cd docker
# Start Postgres, Redis, Kafka, Zookeeper
docker-compose up
```
- Postgres: localhost:5432 (user: postgres, password: postgres, db: pocdb)
- Redis: localhost:6379
- Kafka: localhost:9092
- Zookeeper: localhost:2181

### 2. Run Go Services (in separate terminals)

#### WebSocket Service (Gorilla WebSocket)
```bash
cd services/websocket/server
# Install dependencies if needed
# go mod tidy
# Run the server
go run main.go
```
- Listens on: ws://localhost:8080/ws

#### Wallet Service (gRPC + Kafka Consumer)
```bash
cd services/grpc_kafka/wallet-service
# Install dependencies if needed
# go mod tidy
# Run the service
go run main.go
```
- gRPC endpoint: localhost:50051 (method: GetBalance)

---

## Testing the Flow

1. Send a purchase event:
   - Use a WebSocket client (Go, Postman, browser) to send a message to ws://localhost:8080/ws:
     ```json
     { "user_id": "123", "amount": 49.99 }
     ```
2. Check logs:
   - WebSocket service logs: DB insert, Kafka publish
   - Wallet service logs: Kafka event consumed, Redis updated
3. Query total spent via gRPC:
   - Use Postman (gRPC tab) or any gRPC client:
     - Import services/grpc_kafka/proto/wallet.proto
     - Service: WalletService, Method: GetBalance
     - Request:
       ```json
       { "user_id": "123" }
       ```
     - Response:
       ```json
       { "balances": { "total": 49.99 } }
       ```

---

## Notes
- All infra runs in Docker; Go services run locally and connect to infra via localhost.
- The only required proto file is services/grpc_kafka/proto/wallet.proto.
- No other directories or services are needed. 