# Fintech Backend (Go Microservices)

A production-style fintech backend that simulates core payment platform behavior: authentication, user profile management, wallet operations, transfers, transaction history, and asynchronous notifications.

This project is designed to demonstrate real backend engineering practices for hiring review:
- Microservices architecture
- Clean architecture boundaries
- REST + gRPC communication
- Event-driven notifications
- Security, observability, and test coverage

## What This System Does

- `Auth Service`: register/login with JWT + bcrypt password hashing
- `User Service`: CRUD user profiles with Redis read cache
- `Wallet Service`: create wallet, deposit, withdraw, transfer, transaction history
- `Notification Service`: Kafka-based event consumer for mock email/SMS notifications
- `API Gateway`: public entrypoint with JWT auth and rate limiting

## Architecture

```text
Client (Web/Mobile/Postman/curl)
        |
        v
+------------------------------+
|          API Gateway         |
|   REST, JWT, Rate Limiting   |
+---------------+--------------+
                |
                | gRPC
   +------------+------------+----------------+
   |                         |                |
   v                         v                v
+---------+           +-------------+   +------------+
|  Auth   |           |    User     |   |   Wallet   |
| Service |           |   Service   |   |  Service   |
+----+----+           +------+------+   +------+-----+
     |                       |                 |
     +-----------------------+-----------------+
                             |
                             v
                    +------------------+
                    |    PostgreSQL    |
                    +------------------+
                             |
                             v
                    +------------------+
                    |      Redis       |
                    +------------------+

Auth/Wallet events --> Kafka topic (notifications.v1) --> Notification Service
```

## Tech Stack

- Language: `Go 1.22`
- HTTP framework: `Gin`
- RPC: `gRPC`
- Database: `PostgreSQL` (`pgx`)
- Cache: `Redis`
- Message broker: `Kafka` (Redpanda in Compose)
- Logging: `Zap` structured logs
- Metrics: `Prometheus` format endpoints (`/metrics`)
- Testing: `testing` + `testify`
- DevOps: `Docker`, `Docker Compose`, `GitHub Actions`

## Repository Layout

```text
.
├── services/
│   ├── api-gateway/
│   ├── auth-service/
│   ├── user-service/
│   ├── wallet-service/
│   └── notification-service/
├── shared/
│   ├── config/
│   ├── contracts/       # gRPC contracts/descriptors
│   ├── db/
│   ├── cache/
│   ├── events/
│   ├── middleware/
│   ├── metrics/
│   ├── security/
│   └── grpcx/
├── build/Dockerfile
├── docker-compose.yml
├── Makefile
├── scripts/
└── .github/workflows/ci.yml
```

## Quick Start (Docker)

### Prerequisites

- Docker Desktop (running)

### Run

```bash
docker compose up --build -d
docker compose ps
```

Gateway will be available at:
- `http://localhost:8080`

### Stop

```bash
docker compose down
```

## Local Run (Without Docker)

You need PostgreSQL, Redis, and Kafka running first.

```bash
go run ./services/auth-service/cmd
go run ./services/user-service/cmd
go run ./services/wallet-service/cmd
go run ./services/notification-service/cmd
go run ./services/api-gateway/cmd
```

## Environment Configuration

Use `.env.example` as reference.

Important variables:
- `JWT_SECRET`
- `POSTGRES_URL`
- `REDIS_ADDR`
- `KAFKA_BROKERS`
- `AUTH_GRPC_ADDR`, `USER_GRPC_ADDR`, `WALLET_GRPC_ADDR`

## Public API (via Gateway)

Base URL: `http://localhost:8080/api/v1`

### Auth
- `POST /auth/register`
- `POST /auth/login`

### Users (JWT required)
- `POST /users/`
- `GET /users/`
- `GET /users/:id`
- `PUT /users/:id`
- `DELETE /users/:id`

### Wallets (JWT required)
- `POST /wallets/`
- `GET /wallets/:user_id`
- `POST /wallets/:user_id/deposit`
- `POST /wallets/:user_id/withdraw`
- `POST /wallets/transfer`
- `GET /wallets/:user_id/transactions`

## Example Usage

### 1. Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123"}'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123"}'
```

### 3. Create user profile
```bash
curl -X POST http://localhost:8080/api/v1/users/ \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","full_name":"John Doe","phone":"989111111111"}'
```

### 4. Create wallet
```bash
curl -X POST http://localhost:8080/api/v1/wallets/ \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<AUTH_USER_UUID>","currency":"USD"}'
```

### 5. Deposit
```bash
curl -X POST http://localhost:8080/api/v1/wallets/<AUTH_USER_UUID>/deposit \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"amount":"100.00","reference":"salary"}'
```

### 6. Transfer
```bash
curl -X POST http://localhost:8080/api/v1/wallets/transfer \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"from_user_id":"<FROM_AUTH_UUID>","to_user_id":"<TO_AUTH_UUID>","amount":"25.00","reference":"p2p-transfer"}'
```

## Security

- Bcrypt password hashing
- JWT authentication middleware in gateway
- Input validation with `validator/v10`
- Rate limiting with `x/time/rate`
- Clear transport boundary between public REST and internal gRPC

## Observability

Each service exposes:
- `GET /health`
- `GET /metrics`

This is Prometheus/Grafana-ready.

## Testing & Quality

### Run all tests
```bash
go test ./...
```

### Coverage suite
```bash
make test
make coverage
```

Coverage gate in CI: `>= 70%`.

## CI/CD

GitHub Actions workflow:
1. `go mod tidy`
2. `go test ./...`
3. Coverage suite + threshold check
4. `go build ./...`
5. `docker compose config`

## Verified Status (Current)

Validated on this machine:
- `go test ./...` passed
- Docker Compose stack starts successfully
- Services healthy and reachable
- End-to-end flow passed:
  - register/login
  - user create/get
  - wallet create
  - deposit/withdraw/transfer
  - transaction history