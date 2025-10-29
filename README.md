# DBMS Backend

A Go backend application using Gin framework and Bun ORM with PostgreSQL.

## Prerequisites

- Go 1.25.3 or higher
- PostgreSQL 12 or higher
- Git

## Setup

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Database Setup

Create a PostgreSQL database:

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE dbms;

# Exit psql
\q
```

Run the schema:

```bash
psql -U postgres -d dbms -f schema.sql
```

### 3. Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your database credentials:

```
DB_USER=postgres
DB_PASSWORD=your_password
DB_HOST=localhost
DB_PORT=5432
DB_NAME=dbms
SHOULD_MIGRATE=true
PORT=8080
```

> `SHOULD_MIGRATE` controls whether migrations run automatically when the server boots. Set it to `false` after the schema is up to avoid re-running migrations on every start.

### 4. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### 5. Managing Migrations

- Migrations live under `migrations/` and are managed through [Bun Migrate](https://bun.uptrace.dev/guide/migrations.html).
- On startup, if `SHOULD_MIGRATE=true`, the app will automatically apply any pending migrations and record them in the `bun_migrations` table.
- To inspect the current migration status manually, run:

```bash
go run main.go # with SHOULD_MIGRATE=true in your environment
```

You can safely set `SHOULD_MIGRATE=false` once your schema is up to date.

## Project Structure

```
dmbs-backend/
├── database/          # Database connection and configuration
│   └── db.go
├── models/            # Bun ORM models
│   ├── user.go
│   ├── room.go
│   ├── room_member.go
│   └── request.go
├── migrations/        # Bun migration definitions
├── routes/            # API routes (to be implemented)
├── schema.sql         # PostgreSQL schema
├── main.go            # Application entry point
├── go.mod             # Go module file
└── .env.example       # Example environment variables
```

## Database Schema

### Tables

- **users**: User information with UUID primary key
- **rooms**: Room information with auto-incrementing ID
- **room_members**: Junction table for many-to-many relationship between users and rooms
- **requests**: Service requests (cleaning/maintenance) linked to rooms and users

### Constraints

- One active request per room per type
- Cascade deletion for room members when room or user is deleted
- Set NULL for requests when user is deleted

## API Endpoints

Health check:
- `GET /health` - Check if server is running

(Additional endpoints to be implemented)

## Development

To add new routes, create handler files in the `routes/` directory and register them in `main.go`.

## Technologies Used

- **Gin**: HTTP web framework
- **Bun**: SQL-first ORM for PostgreSQL
- **PostgreSQL**: Database
- **UUID**: For user IDs
