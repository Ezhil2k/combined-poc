# Unified POC: Target Architecture & Current Service Inventory

## 🚀 Target Business Logic Flow (Post-Refactor)

This project aims to unify multiple POCs into a single, Docker-orchestrated system with the following end-to-end flow:

1. **WebSocket client** sends a message:
   ```json
   {
     "user_id": "123",
     "amount": 49.99
   }
   ```
2. **websocket/** service:
   - Saves the purchase to **PostgreSQL**
   - Publishes a Kafka event to the `user_purchase` topic
3. **grpc_kafka/** service:
   - Consumes the Kafka event
   - Aggregates the total amount spent by user
   - Stores the total in **Redis** (introduced in the unified architecture)
   - Exposes a gRPC endpoint `GetTotalSpent(user_id)` which reads from Redis

**All services (websocket, grpc_kafka, redis, postgres, kafka, zookeeper) will be containerized and managed via a single `docker-compose.yml`.**

---

## 🗂️ Current Service Inventory (Pre-Refactor)

> **Note:** Each service is standalone. There are no nested or sub-services. The following describes the current state before unification.

### 1. go-gorilla (WebSocket Service)
- **Location:** `services/go-gorilla/`
- **Current DB Connection:**
  - Connects to Postgres directly (details in `server/main.go`).
  - Connection parameters may be hardcoded or read from environment variables.
- **Kafka:**
  - Not used currently.
- **Redis:**
  - Not used currently.
- **Notes:**
  - Contains both server and client code.

### 2. trade-engine-poc (gRPC + Kafka)
- **Location:** `services/trade-engine-poc/`
- **Proto Files:**
  - Located in `proto/` (e.g., `wallet.proto`).
- **wallet-service:**
  - Standalone Go service.
  - Exposes a gRPC endpoint.
  - Communicates with other services via gRPC and Kafka only.
  - Does **not** use Redis or Postgres in the current implementation.
- **trade-service:**
  - Standalone Go service.
  - Communicates via Kafka and gRPC only.
  - Does **not** use Redis or Postgres in the current implementation.
- **notification-service:**
  - Standalone Go service (out of scope for unified POC unless needed).
- **Kafka:**
  - Used for event-driven communication (user purchases, etc.).
- **Redis:**
  - **Not used in the current implementation.**
- **Postgres:**
  - **Not used in the current implementation.**
- **Notes:**
  - Redis will be introduced in the unified architecture for aggregation caching.

### 3. go-postgreSQL (Postgres Logic)
- **Location:** `services/go-postgreSQL/`
- **Current Setup:**
  - Standalone Go service for Postgres interaction.
  - May use systemd/manual scripts for Postgres startup.
  - Contains Go code for DB schema, queries, and user logic.
- **Notes:**
  - Will migrate all Postgres setup to Docker Compose.

### 4. go-redis (Redis Logic)
- **Location:** `services/go-redis/`
- **Current Setup:**
  - Standalone Go service for Redis interaction.
  - Not used by other services in the current implementation.
  - May be merged into grpc_kafka for unified flow.

### 5. Postgres (Systemd/Manual)
- **Current Setup:**
  - May be started via systemd or manually on host.
  - Will be migrated to Docker Compose for unified orchestration.

---

## Next Steps
- Refactor all services to use Docker Compose for orchestration.
- Update all connection parameters to use Docker service names.
- Migrate proto files and update imports.
- Implement the new business logic flow as described above.
- Preserve existing logic as reference during refactor. 