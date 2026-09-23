# Property Listings API

A production-grade RESTful API for managing property listings, built with Go, Chi, PostgreSQL, and PostGIS. This API supports full CRUD operations, complex filtering, and high-performance geospatial searches.

## 🚀 Tech Stack

- **Language:** Go (1.22+)
- **Router:** [go-chi/chi](https://github.com/go-chi/chi) (Lightweight, idiomatic HTTP router)
- **Database:** PostgreSQL with [PostGIS](https://postgis.net/) (Geospatial extension)
- **DB Driver:** [jackc/pgx/v5](https://github.com/jackc/pgx) (High-performance Postgres driver with connection pooling)
- **Validation:** [go-playground/validator](https://github.com/go-playground/validator) (Struct-level validation)
- **Logging:** `log/slog` (Standard library structured JSON logging)

---

## 🛠️ Setup Instructions

### Prerequisites

- Go 1.22 or higher
- Docker & Docker Compose

### 1. Clone the Repository

```bash
git clone https://github.com/pasDamola/property-listings-api.git
cd property-listings-api
```

### 2. Environment Variables

Copy the example environment file and adjust if necessary:

```bash
cp .env.example .env
```

Ensure your `.env` contains:

```env
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/property_db?sslmode=disable
```

### 3. Start the Database (Docker)

```bash
docker-compose up -d
```

> **Note for Apple Silicon (M1/M2/M3) users:** The `docker-compose.yml` includes `platform: linux/amd64` to ensure the PostGIS image runs correctly via Rosetta emulation.

### 4. Run Database Migrations

Since `psql` may not be installed locally, execute the migration directly inside the running Docker container:

```bash
docker exec -i $(docker-compose ps -q db) psql -U user -d property_db < migrations/001_create_listings_table.sql
```

### 5. Install Go Dependencies

```bash
go mod tidy
```

### 6. Run the Application

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`. You should see structured JSON logs indicating the server is running and connected to the database.

### 7. Run Tests

```bash
go test ./... -v
```

---

## 📡 API Endpoints

| Method | Endpoint                      | Description                                       |
| ------ | ------------------------------ | -------------------------------------------------- |
| POST   | `/api/v1/listings`            | Create a new property listing                     |
| GET    | `/api/v1/listings/search`     | Search listings with filters and geospatial radius |
| GET    | `/api/v1/listings/{id}`       | Retrieve a specific listing by ID                 |
| PUT    | `/api/v1/listings/{id}`       | Update an existing listing                        |
| DELETE | `/api/v1/listings/{id}`       | Delete a listing                                  |

### Example Search Query

```http
GET /api/v1/listings/search?lat=6.5244&lng=3.3792&radius=10&type=rent&minPrice=1000&page=1&limit=10
```

---

## 🧠 Design Choices

As a senior engineer, I prioritize maintainability, performance, and security. Here are the key architectural decisions made:

### 1. PostGIS for Geospatial Queries (Performance)

The task required searching for listings within X km of a given point. A naive implementation would fetch all records into Go memory and calculate distances using formula. This is an O(N) operation and will crash the server under load.

Instead, I pushed the computation to the database using PostGIS:

- Used the `GEOGRAPHY(POINT, 4326)` data type to account for the Earth's curvature, providing accurate meter-based distances.
- Utilized `ST_DWithin` for radius searches, which leverages a GiST spatial index. This allows PostgreSQL to instantly eliminate 99% of irrelevant records before calculating distances.

### 2. Layered Architecture (Maintainability)

The codebase is strictly separated into `handler` (HTTP), `service` (Business Logic), and `repository` (Data Access) layers.

- This separation of concerns makes the code highly testable. For example, the business logic can be unit-tested without a database, and the HTTP handlers can be tested with mock services.
- It decouples the API contract from the database schema, making future migrations or framework swaps significantly easier.

### 3. Raw SQL with pgx over an ORM (Control & Speed)

I chose `jackc/pgx` and wrote raw, parameterized SQL instead of using an ORM like GORM.

- ORMs often generate inefficient SQL for complex geospatial queries.
- Raw SQL gives me complete control over the query plan and allows me to leverage PostgreSQL-specific features like `ST_MakePoint`.
- I used parameterized queries (`$1`, `$2`, `$3`) exclusively, which completely prevents SQL injection attacks.

### 4. Production-Ready Safeguards

- **Graceful Shutdown:** The server listens for `SIGTERM`/`SIGINT` signals and gracefully drains connections before exiting, ensuring zero downtime during deployments.
- **Structured Logging:** Utilizes Go's standard `log/slog` package to emit JSON logs. This is critical for ingestion by log aggregators (like Datadog or ELK) for alerting and monitoring.
- **Centralized Error Handling:** Custom middleware recovers from panics, logs the stack trace internally, and returns a clean, standardized JSON 500 response to the client without leaking internal architecture.
- **Strict Validation:** `go-playground/validator` ensures malformed payloads are rejected at the edge (HTTP layer) before consuming database resources.

---

## 🔮 What I'd Improve with More Time

Given more time, the following features would be prioritized to make this a truly enterprise-grade service:

1. **Authentication & Authorization** — Implement JWT-based authentication middleware. Currently, any user can mutate listings. In production, only verified agents with the correct role-based access control (RBAC) should be able to create, update, or delete their own listings.
2. **Automated Migrations** — Integrate a tool like `golang-migrate` so the database schema updates automatically on application startup or during the CI/CD pipeline, removing the need for manual SQL execution.
3. **Caching Layer** — Introduce Redis to cache frequent search queries. Geospatial searches are computationally expensive; caching popular location searches would drastically reduce database load and improve response times.
4. **Rate Limiting** — Add `httprate` middleware to prevent API abuse and brute-force attacks.
5. **API Documentation** — Integrate Swagger/OpenAPI using `swaggo/swag` to provide interactive, self-documenting API endpoints for frontend teams.
6. **CI/CD Pipeline** — Set up a GitHub Actions workflow to automatically run `go test`, `go vet`, and `golangci-lint` on every push, ensuring code quality and preventing regressions.