# Assignment - Golang Transaction Service

A standalone Docker application that processes transactions and manages user balances in PostgreSQL.

## Features

- **POST /user/{userId}/transaction**: Process transactions (win/lose) with idempotency
- **GET /user/{userId}/balance**: Get current user balance
- Idempotent transaction processing (duplicate transactions are ignored)
- Non-negative balance enforcement
- Atomic database transactions for consistency
- Support for multiple source types (game, server, payment)
- Concurrency-safe operations with database locking

## Prerequisites

- Docker
- Docker Compose

## Quick Start

### 1. Build and Run the Application

```bash
docker compose up -d --build
```

This will:
- Start a PostgreSQL database container
- Build and start the Go application container
- Run database migrations automatically
- Seed initial users (IDs: 1, 2, 3 with balances: 100.00, 50.00, 0.00)

### 2. Verify the Application is Running

Check the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response: `OK`

### 3. Test the API

#### Get User Balance

```bash
curl http://localhost:8080/user/1/balance
```

Expected response:
```json
{
  "userId": 1,
  "balance": "100.00"
}
```

#### Process a Win Transaction

```bash
curl -X POST http://localhost:8080/user/1/transaction \
  -H "Content-Type: application/json" \
  -H "Source-Type: game" \
  -d '{
    "state": "win",
    "amount": "10.50",
    "transactionId": "txn-001"
  }'
```

Expected response:
```json
{
  "userId": 1,
  "transactionId": "txn-001",
  "balance": "110.50",
  "message": "Transaction applied successfully"
}
```

#### Process a Lose Transaction

```bash
curl -X POST http://localhost:8080/user/1/transaction \
  -H "Content-Type: application/json" \
  -H "Source-Type: server" \
  -d '{
    "state": "lose",
    "amount": "25.00",
    "transactionId": "txn-002"
  }'
```

Expected response:
```json
{
  "userId": 1,
  "transactionId": "txn-002",
  "balance": "85.50",
  "message": "Transaction applied successfully"
}
```

#### Test Duplicate Transaction (Idempotency)

```bash
curl -X POST http://localhost:8080/user/1/transaction \
  -H "Content-Type: application/json" \
  -H "Source-Type: payment" \
  -d '{
    "state": "win",
    "amount": "10.50",
    "transactionId": "txn-001"
  }'
```

Expected response (same transactionId as before):
```json
{
  "userId": 1,
  "transactionId": "txn-001",
  "balance": "85.50",
  "message": "Duplicate transaction ignored"
}
```

#### Test Insufficient Funds

```bash
curl -X POST http://localhost:8080/user/3/transaction \
  -H "Content-Type: application/json" \
  -H "Source-Type: game" \
  -d '{
    "state": "lose",
    "amount": "100.00",
    "transactionId": "txn-003"
  }'
```

Expected response:
```json
{
  "userId": 3,
  "transactionId": "txn-003",
  "balance": "0.00",
  "message": "Insufficient funds"
}
```

## Running Tests

### Option 1: Run Tests Inside Docker Container
 
First, make sure the application is running:

```bash
docker compose up -d
```

Then run tests inside the app container:

```bash
docker compose exec app go test ./...
```

### Option 2: Run Tests Locally (Requires Local PostgreSQL)

If you have PostgreSQL running locally, you can run tests directly:

```bash
go test ./...
```

**Note**: Tests require a PostgreSQL database. The test suite will:
- Skip tests if database is not available
- Create a test database schema
- Clean up after tests

### Running Specific Test Suites

```bash
# Run only unit tests
docker compose exec app go test ./internal/utils/...

# Run only integration tests
docker compose exec app go test ./internal/core/...
docker compose exec app go test ./internal/http/...
```

## Project Structure

```
assignment/
├── cmd/
│   └── app/
│       └── main.go              # Application entry point
├── internal/
│   ├── core/
│   │   ├── logic.go             # Business logic for transactions
│   │   └── logic_test.go        # Unit tests for transaction logic
│   ├── db/
│   │   └── database.go          # Database connection and migrations
│   ├── http/
│   │   ├── handlers.go          # HTTP route handlers
│   │   └── handlers_test.go     # Integration tests for handlers
│   ├── models/
│   │   ├── user.go              # User model
│   │   └── transaction.go       # Transaction models
│   └── utils/
│       ├── validation.go        # Validation helpers
│       └── validation_test.go   # Unit tests for validation
├── Dockerfile                   # Docker image definition
├── docker-compose.yml           # Docker Compose configuration
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README.md                    # This file
```

## API Endpoints

### POST /user/{userId}/transaction

Processes a transaction for a user.

**Headers:**
- `Source-Type`: `game`, `server`, or `payment` (required)
- `Content-Type`: `application/json` (required)

**Request Body:**
```json
{
  "state": "win | lose",
  "amount": "10.15",
  "transactionId": "unique_identifier"
}
```

**Response Codes:**
- `200 OK`: Transaction processed successfully, duplicate ignored, or insufficient funds
- `400 Bad Request`: Invalid request (missing headers, invalid format, etc.)
- `500 Internal Server Error`: Server error

### GET /user/{userId}/balance

Returns the current balance for a user.

**Response:**
```json
{
  "userId": 1,
  "balance": "9.25"
}
```

**Response Codes:**
- `200 OK`: Success
- `400 Bad Request`: Invalid user ID
- `404 Not Found`: User not found
- `500 Internal Server Error`: Server error

## Database Schema

### Users Table
- `id` (BIGSERIAL PRIMARY KEY): User ID
- `balance` (NUMERIC(10,2)): User balance (default: 0)
- `created_at` (TIMESTAMP): Creation timestamp
- `updated_at` (TIMESTAMP): Last update timestamp

### Transactions Table
- `id` (BIGSERIAL PRIMARY KEY): Transaction ID
- `user_id` (BIGINT): Reference to users table
- `transaction_id` (TEXT UNIQUE): Idempotency key
- `state` (TEXT): `win` or `lose`
- `amount` (NUMERIC(10,2)): Transaction amount
- `source_type` (TEXT): `game`, `server`, or `payment`
- `applied` (BOOLEAN): Whether transaction was applied
- `created_at` (TIMESTAMP): Creation timestamp

## Initial Data

The application automatically seeds three users on startup:
- User ID 1: Balance 100.00
- User ID 2: Balance 50.00
- User ID 3: Balance 0.00

## Stopping the Application

```bash
docker compose down
```

To also remove volumes (database data):

```bash
docker compose down -v
```

## Environment Variables

The application supports the following environment variables:

- `DATABASE_URL`: PostgreSQL connection string (default: `host=postgres user=postgres password=postgres dbname=assignment sslmode=disable`)
- `PORT`: Server port (default: `8080`)

These are configured in `docker-compose.yml` and can be overridden if needed.

## Troubleshooting

### Application won't start

1. Check if ports 8080 and 5432 are available:
   ```bash
   # Windows
   netstat -ano | findstr :8080
   netstat -ano | findstr :5432
   
   # Linux/Mac
   lsof -i :8080
   lsof -i :5432
   ```

2. Check container logs:
   ```bash
   docker compose logs app
   docker compose logs postgres
   ```

3. Verify database is ready:
   ```bash
   docker compose ps
   ```

### Tests are failing

1. Ensure the application containers are running:
   ```bash
   docker compose up -d
   ```

2. Check if test database can be accessed:
   ```bash
   docker compose exec app go test -v ./...
   ```

### Database connection errors

1. Wait a few seconds after starting containers for PostgreSQL to initialize
2. Check PostgreSQL logs:
   ```bash
   docker compose logs postgres
   ```
3. Restart the application container:
   ```bash
   docker compose restart app
   ```

## Development

### Building Locally

```bash
go mod download
go build -o bin/app ./cmd/app
```

### Running Locally (with local PostgreSQL)

```bash
export DATABASE_URL="host=localhost user=postgres password=postgres dbname=assignment sslmode=disable"
export PORT=8080
go run cmd/app/main.go
```

## License

This is an assignment project.
