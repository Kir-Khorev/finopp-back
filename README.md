# Finopp Backend

Financial advisor API built with Go.

## Quick Start

### Prerequisites

- Docker Desktop installed and running

### Run with Docker Compose

```bash
# Start all services (API, Postgres, Redis)
docker-compose up --build

# Stop services
docker-compose down

# Stop and remove volumes (clean state)
docker-compose down -v
```

### Access

- API: http://localhost:8080
- Health check: http://localhost:8080/health

### Environment Variables

Copy `.env.example` to `.env` and configure:

- `GROQ_API_KEY` - get from https://console.groq.com

## Project Structure

```
finopp-back/
├── cmd/api/              # Application entry point
├── internal/             # Private application code
│   ├── auth/            # Authentication
│   ├── users/           # User management
│   ├── finance/         # Financial profiles
│   ├── advice/          # AI advice service
│   ├── analytics/       # Analytics
│   └── common/          # Shared code
├── pkg/                  # Public libraries
│   ├── config/          # Configuration
│   └── logger/          # Logging
├── migrations/           # Database migrations
├── docker-compose.yml    # Docker services
└── Dockerfile           # API container
```

## API Endpoints

### Health

- `GET /health` - Health check

### Auth (coming soon)

- `POST /api/v1/auth/register` - Register user
- `POST /api/v1/auth/login` - Login user

### Profile (coming soon)

- `GET /api/v1/profile` - Get user profile
- `PUT /api/v1/profile` - Update profile

### Advice (coming soon)

- `POST /api/v1/advice` - Get financial advice

## Development

### Run locally (without Docker)

Requires Go 1.23+, PostgreSQL, Redis installed locally.

```bash
# Install dependencies
go mod download

# Run migrations manually
# (connect to Postgres and run SQL from internal/common/db.go)

# Run API
go run cmd/api/main.go
```
