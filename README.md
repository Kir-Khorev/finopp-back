# Finopp Backend

Go REST API for AI-powered financial advisory application. Uses Groq LLM (Llama 3.3 70B) for generating financial advice.

## 🚀 Quick Start

### Prerequisites

- **Docker Desktop** installed and running
- **Groq API Key** from https://console.groq.com

### Installation & Run

```bash
# 1. Create .env file
cp .env.example .env

# 2. Edit .env and add your Groq API key
nano .env  # or use your editor

# 3. Start all services (API, Postgres, Redis)
docker-compose up --build -d

# 4. Check status
docker-compose ps

# 5. Test API
curl http://localhost:8080/health
```

API will be available at **http://localhost:8080**

### Stop Services

```bash
# Stop containers
docker-compose down

# Stop and remove data (clean state)
docker-compose down -v
```

---

## 📁 Project Structure

```
finopp-back/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point - starts server
│
├── internal/                 # Private application code (not importable)
│   ├── auth/                # Authentication & authorization
│   │   ├── handler.go       # HTTP handlers (register, login)
│   │   ├── service.go       # Business logic (JWT, passwords)
│   │   ├── repository.go    # Database queries
│   │   └── models.go        # Data structures
│   │
│   ├── advice/              # AI advice feature
│   │   ├── handler.go       # HTTP handler for /advice endpoint
│   │   ├── service.go       # Groq API integration
│   │   └── models.go        # Request/response types
│   │
│   ├── users/               # User management (planned)
│   ├── finance/             # Financial profiles (planned)
│   ├── analytics/           # Analytics (planned)
│   │
│   └── common/              # Shared utilities
│       ├── db.go           # PostgreSQL connection + migrations
│       └── redis.go        # Redis connection
│
├── pkg/                     # Public libraries (importable by other projects)
│   ├── config/
│   │   └── config.go       # Environment variable loading
│   └── logger/             # Logging utilities
│
├── migrations/              # SQL migration files (if using migrate tool)
├── docker-compose.yml       # Defines 3 services: api, postgres, redis
├── Dockerfile              # Multi-stage build for Go API
├── .env.example            # Template for environment variables
├── .env                    # Actual secrets (NEVER COMMIT THIS)
├── go.mod                  # Go module definition
└── go.sum                  # Dependency checksums
```

---

## 🔌 API Endpoints

### Health Check
- **GET** `/health` - Returns server status

### Authentication
- **POST** `/api/v1/auth/register` - Register new user
  - Body: `{ "email": "...", "password": "...", "name": "..." }`
- **POST** `/api/v1/auth/login` - Login user
  - Body: `{ "email": "...", "password": "..." }`
  - Returns JWT token

### AI Advice
- **POST** `/api/v1/advice` - Get financial advice from AI
  - Body: `{ "question": "Что такое инвестиции?" }`
  - Returns: `{ "answer": "..." }`

### Coming Soon
- Profile management (`/api/v1/profile`)
- Financial data tracking
- Analytics

---

## 🛠️ Development

### Option 1: Docker (Recommended)

**Pros:** Consistent environment, no local dependencies  
**Cons:** Slower rebuild, no hot-reload

```bash
# Start services
docker-compose up --build

# View logs
docker-compose logs -f api

# Rebuild after code changes
docker-compose up --build -d

# Execute commands in container
docker exec -it finopp-api sh
```

### Option 2: Local Go (Faster iteration)

**Pros:** Hot-reload, faster development  
**Cons:** Requires Go 1.23+ installed

```bash
# Start only database services
docker-compose up postgres redis -d

# Install Go dependencies
go mod download

# Run API locally
go run cmd/api/main.go

# Or with hot-reload (install air first)
air
```

### Option 3: Hybrid (Best of both)

```bash
# 1. Start DB in Docker
docker-compose up postgres redis -d

# 2. Run API locally
go run cmd/api/main.go
```

---

## 🗄️ Database

### PostgreSQL

**Connection Details:**
- Host: `localhost`
- Port: `5432`
- Database: `finopp_db`
- User: `finopp`
- Password: `finopp_pass`

**Access via CLI:**
```bash
docker exec -it finopp-postgres psql -U finopp -d finopp_db
```

**Common SQL Commands:**
```sql
-- List tables
\dt

-- Describe table
\d users

-- Query users
SELECT * FROM users;

-- Exit
\q
```

### Migrations

Currently migrations run automatically on startup via `common.RunMigrations()`.

**Migration is in:** `internal/common/db.go`

**To add new table:**
1. Edit `RunMigrations()` in `internal/common/db.go`
2. Add new `CREATE TABLE IF NOT EXISTS ...` statement
3. Restart API

---

## 🔴 Redis

**Used for:** Caching, session storage (future), rate limiting (future)

**Connection Details:**
- Host: `localhost`
- Port: `6379`
- No password (local dev)

**Access via CLI:**
```bash
docker exec -it finopp-redis redis-cli

# Test commands
> PING
> SET test "hello"
> GET test
> DEL test
> exit
```

---

## 🧪 Testing

### Manual Testing

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456","name":"Test User"}'

# Get AI advice
curl -X POST http://localhost:8080/api/v1/advice \
  -H "Content-Type: application/json" \
  -d '{"question":"Что такое акции?"}'
```

### Unit Tests (Not implemented yet)

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth
```

---

## 📦 Dependencies

Key Go packages:
- **echo** - Web framework (routing, middleware)
- **pq** - PostgreSQL driver
- **redis** - Redis client
- **jwt** - JWT token generation
- **bcrypt** - Password hashing

To add new dependency:
```bash
go get github.com/package/name
go mod tidy
```

---

## ⚙️ Configuration

### Environment Variables (.env)

**Required:**
```env
GROQ_API_KEY=your_key_here    # Get from https://console.groq.com
```

**Optional (defaults shown):**
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=finopp
DB_PASSWORD=finopp_pass
DB_NAME=finopp_db
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key-change-in-production
```

**Load mechanism:** `pkg/config/config.go` reads from `.env` file and environment.

---

## 🐳 Docker Details

### Services

1. **api** - Go application (port 8080)
2. **postgres** - Database (port 5432)
3. **redis** - Cache (port 6379)

### Useful Commands

```bash
# View all containers
docker-compose ps

# View logs
docker-compose logs -f [service]  # api, postgres, redis

# Restart single service
docker-compose restart api

# Rebuild single service
docker-compose up --build api -d

# Remove everything
docker-compose down -v

# Access container shell
docker exec -it finopp-api sh

# Check resource usage
docker stats
```

---

## ⚠️ Common Issues

### Port already in use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in .env
PORT=8081
```

### Database connection refused

1. Check Postgres is running: `docker-compose ps`
2. Check `.env` has correct DB credentials
3. Wait for healthcheck: `docker-compose logs postgres`
4. Verify connection: `docker exec -it finopp-postgres psql -U finopp -d finopp_db`

### Groq API errors

1. Verify `GROQ_API_KEY` in `.env`
2. Check API quota: https://console.groq.com
3. Test API directly: `curl https://api.groq.com/...`

### Code changes not reflecting

- If using Docker: `docker-compose up --build -d`
- If using local Go: Just restart `go run cmd/api/main.go` (or use `air` for hot-reload)

---

## 🚀 Deployment (Future)

Deploy to:
- **Render.com** (already configured via `render.yaml`)
- **Railway.app**
- **Fly.io**
- **AWS ECS**

**Remember:**
- Set all environment variables in platform
- Use managed PostgreSQL/Redis in production
- Enable HTTPS
- Set strong JWT_SECRET
- Configure CORS properly

---

## 📝 Notes for Developers

- **Standard Go project layout:** `cmd/`, `internal/`, `pkg/`
- **Echo framework:** Similar to Express.js (if you know Node)
- **No ORM:** Raw SQL queries (simple, fast, clear)
- **Migrations:** Auto-run on startup (see `internal/common/db.go`)
- **CORS:** Enabled for all origins (dev only - restrict in prod!)
- **Errors:** Return JSON with `{"error": "message"}` format
- **Logging:** Uses Echo's built-in logger + `log` package

---

## 🎯 Roadmap

- [ ] Add comprehensive tests
- [ ] JWT authentication middleware
- [ ] User profile endpoints
- [ ] Financial data tracking
- [ ] Analytics endpoints
- [ ] Rate limiting
- [ ] API documentation (Swagger)
- [ ] CI/CD pipeline
