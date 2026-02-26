# Go Minimal Template — Chi + MySQL

> Minimal Go backend template, zero bloat, ready to extend.

## Stack
- **Router**: [Chi v5](https://github.com/go-chi/chi)
- **Database**: MySQL via `database/sql`
- **Config**: `.env` via `godotenv`

## Structure
```
├── cmd/api/main.go        ← entry point + wire up
├── config/                ← load env
├── internal/
│   ├── handler/           ← HTTP layer
│   ├── service/           ← business logic
│   ├── repository/        ← DB queries
│   └── model/             ← structs
├── pkg/database/          ← DB connection
└── migrations/            ← SQL files
```

## Quick Start

```bash
# 1. copy env
cp .env.example .env

# 2. run migration
mysql -u root -p < migrations/001_init.sql

# 3. install deps
go mod tidy

# 4. run
make run
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/users` | List users |
| POST | `/api/v1/users` | Create user |
| GET | `/api/v1/users/:id` | Get user |
| DELETE | `/api/v1/users/:id` | Delete user |

## Next Steps (เพิ่มเองได้)
- [ ] JWT Auth middleware
- [ ] Input validation (`go-playground/validator`)
- [ ] Structured logging (`zap` / `slog`)
- [ ] Docker + docker-compose
- [ ] GitHub Actions CI
