[![Dependencies](https://github.com/greemin/question-voting-app/actions/workflows/dependabot/dependabot-updates/badge.svg)](https://github.com/greemin/question-voting-app/actions/workflows/dependabot/dependabot-updates)
[![Tests](https://github.com/greemin/question-voting-app/actions/workflows/ci.yml/badge.svg)](https://github.com/greemin/question-voting-app/actions/workflows/ci.yml)
[![Publish](https://github.com/greemin/question-voting-app/actions/workflows/publish.yml/badge.svg)](https://github.com/greemin/question-voting-app/actions/workflows/publish.yml)

# :question: :ballot_box_with_check: Question Voting App
A real-time, lightweight application for crowdsourcing and ranking questions during a presentation, Q&A session, or meeting. Built using a Go backend for fast API handling and a React/Vite frontend for a modern, responsive user experience.

**Live demo:** https://question-app.duckdns.org

## 🛠️ Setup and Installation

You must have **Docker** and **Docker Compose** installed on your system.

### Running locally (SQLite — default)

SQLite is the default storage backend. No database container is needed.

    docker compose up --build

The following services will be available:
* **Frontend:** http://localhost:5173
* **Backend API:** http://localhost:8081

Session data is persisted in a named Docker volume (`sqlite-dev-data`). To wipe it, run `docker compose down -v`.

### Running locally (MongoDB)

To use MongoDB instead, activate the `mongo` profile and set `DB_DRIVER`:

    DB_DRIVER=mongodb docker compose --profile mongo up --build

The MongoDB instance is available at `mongodb://devroot:devpassword@localhost:27017`.

## Storage backends

The backend supports two storage drivers, selected via the `DB_DRIVER` environment variable:

| `DB_DRIVER` | Description |
|---|---|
| `sqlite` (default) | Embedded SQLite database. No extra service required. Data stored at `SQLITE_FILE` (default: `data.db`). |
| `mongodb` | External MongoDB instance. Requires `MONGO_URI` to be set. |

Sessions expire after 24 hours. SQLite uses a background cleanup goroutine; MongoDB uses a TTL index.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `DB_DRIVER` | `sqlite` | Storage backend (`sqlite` or `mongodb`) |
| `SQLITE_FILE` | `data.db` | Path to the SQLite database file |
| `MONGO_URI` | — | MongoDB connection string (required when `DB_DRIVER=mongodb`) |
| `PORT` | `8081` | Backend listen port |
| `CORS_ORIGINS` | `http://localhost:5174` | Allowed CORS origin |
| `ENV` | — | Set to `production` to enable secure cookies |

## Running the tests

    # Unit tests
    go test ./...

    # Integration tests (requires Docker for the MongoDB test)
    go test -tags integration -v ./internal/storage/...

## Load Testing

k6 load tests targeting the production environment are in [`k6/`](k6/). See [`k6/README.md`](k6/README.md) for usage.
