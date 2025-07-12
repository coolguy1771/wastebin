# Wastebin

**Wastebin** is a self-hosted web service that allows you to share pastes anonymously. Wastebin is designed with modern observability practices and uses the following tech stack:

| Component    | Framework/Technology     |
|--------------|--------------------------|
| Backend      | Go with Chi Router       |
| Database     | PostgreSQL/SQLite        |
| Frontend     | React with TypeScript    |
| UI Library   | Material-UI (MUI)        |
| Build Tool   | Vite                     |
| Observability| OpenTelemetry            |

## Configuration

All configuration is done via environment variables with the `WASTEBIN_` prefix.

### Core Configuration

| Environment Variable         | Description                                                    | Default     | Required |
|:----------------------------:|----------------------------------------------------------------|-------------|:--------:|
| `WASTEBIN_WEBAPP_PORT`       | The port wastebin will listen on                              | `3000`      | ❌       |
| `WASTEBIN_LOCAL_DB`          | Use local SQLite database instead of PostgreSQL               | `false`     | ❌       |
| `WASTEBIN_LOG_LEVEL`         | Logging level (DEBUG, INFO, WARN, ERROR)                      | `INFO`      | ❌       |

### Database Configuration

| Environment Variable         | Description                                                    | Default     | Required |
|:----------------------------:|----------------------------------------------------------------|-------------|:--------:|
| `WASTEBIN_DB_USER`           | The user to use when connecting to a database                 | `wastebin`  | ✅*      |
| `WASTEBIN_DB_HOST`           | The hostname or ip address of the database to connect to      | `localhost` | ✅*      |
| `WASTEBIN_DB_PORT`           | The port to connect to the database on                        | `5432`      | ❌       |
| `WASTEBIN_DB_PASSWORD`       | The password to connect to the database with                  |             | ✅*      |
| `WASTEBIN_DB_NAME`           | The name of the database to use                               | `wastebin`  | ❌       |
| `WASTEBIN_DB_MAX_IDLE_CONNS` | The maximum number of idle connections to use                 | `10`        | ❌       |
| `WASTEBIN_DB_MAX_OPEN_CONNS` | The maximum number of connections the database can have       | `50`        | ❌       |

*Required only when using PostgreSQL (when `WASTEBIN_LOCAL_DB=false`)

### Observability Configuration

| Environment Variable            | Description                                                    | Default         | Required |
|:-------------------------------:|----------------------------------------------------------------|-----------------|:--------:|
| `WASTEBIN_TRACING_ENABLED`      | Enable OpenTelemetry tracing                                  | `true`          | ❌       |
| `WASTEBIN_METRICS_ENABLED`      | Enable OpenTelemetry metrics                                  | `true`          | ❌       |
| `WASTEBIN_SERVICE_NAME`         | Service name for observability                                | `wastebin`      | ❌       |
| `WASTEBIN_SERVICE_VERSION`      | Service version for observability                             | `1.0.0`         | ❌       |
| `WASTEBIN_ENVIRONMENT`          | Environment name (development, staging, production)           | `development`   | ❌       |
| `WASTEBIN_OTLP_TRACE_ENDPOINT`  | OTLP trace endpoint (host:port format)                        | `localhost:4318`| ❌       |
| `WASTEBIN_OTLP_METRICS_ENDPOINT`| OTLP metrics endpoint (host:port format)                      | `localhost:4318`| ❌       |
| `WASTEBIN_METRICS_INTERVAL`     | Metrics collection interval in seconds                        | `30`            | ❌       |

## Running Wastebin

### With Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  wastebin:
    image: ghcr.io/coolguy1771/wastebin:latest
    restart: unless-stopped
    environment:
      # Basic Configuration
      - WASTEBIN_WEBAPP_PORT=3000
      - WASTEBIN_LOG_LEVEL=INFO
      
      # Database Configuration (PostgreSQL)
      - WASTEBIN_LOCAL_DB=false
      - WASTEBIN_DB_HOST=postgres
      - WASTEBIN_DB_USER=wastebin
      - WASTEBIN_DB_PASSWORD=mysecretpassword
      - WASTEBIN_DB_NAME=wastebin
      
      # Observability (Optional - comment out if not using)
      - WASTEBIN_TRACING_ENABLED=true
      - WASTEBIN_METRICS_ENABLED=true
      - WASTEBIN_SERVICE_NAME=wastebin
      - WASTEBIN_ENVIRONMENT=production
      - WASTEBIN_OTLP_TRACE_ENDPOINT=jaeger:14268
      - WASTEBIN_OTLP_METRICS_ENDPOINT=prometheus:9090
    ports:
      - "3000:3000"
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/health"]
      timeout: 5s
      interval: 30s
      retries: 3

  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_PASSWORD=mysecretpassword
      - POSTGRES_USER=wastebin
      - POSTGRES_DB=wastebin
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wastebin"]
      timeout: 5s
      interval: 10s
      retries: 5

volumes:
  postgres_data:
```

### With SQLite (Standalone)

For a simple deployment with SQLite:

```yaml
version: '3.8'

services:
  wastebin:
    image: ghcr.io/coolguy1771/wastebin:latest
    restart: unless-stopped
    environment:
      - WASTEBIN_WEBAPP_PORT=3000
      - WASTEBIN_LOCAL_DB=true
      - WASTEBIN_LOG_LEVEL=INFO
    ports:
      - "3000:3000"
    volumes:
      - wastebin_data:/app/data
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/health"]
      timeout: 5s
      interval: 30s
      retries: 3

volumes:
  wastebin_data:
```

### Local Development

For local development:

```bash
# Clone the repository
git clone https://github.com/coolguy1771/wastebin.git
cd wastebin

# Run with local development script
chmod +x run-local.sh
./run-local.sh
```

This will build the frontend and start the server with SQLite for development.

## Features

- 🎨 **Modern UI**: Clean, responsive interface built with React and Material-UI
- 🔒 **Anonymous**: No registration required - share pastes instantly
- ⚡ **Fast**: Built with Go for high performance
- 📊 **Observability**: Full OpenTelemetry integration for monitoring and tracing  
- 🗄️ **Flexible Storage**: Support for both PostgreSQL and SQLite
- 🔧 **Configurable**: Environment-based configuration for easy deployment
- 📱 **Responsive**: Works great on desktop and mobile devices
- 🎯 **Syntax Highlighting**: Automatic language detection and highlighting
- ⏰ **Expiration**: Set custom expiration times for pastes
- 🔥 **Burn After Reading**: One-time view option for sensitive content

## API Endpoints

Wastebin provides a RESTful API for programmatic access:

| Method | Endpoint           | Description                    |
|--------|--------------------|--------------------------------|
| POST   | `/api/v1/paste`    | Create a new paste            |
| GET    | `/api/v1/paste/:id`| Retrieve a paste by ID        |
| DELETE | `/api/v1/paste/:id`| Delete a paste by ID          |
| GET    | `/api/v1/paste/:id/raw` | Get raw paste content    |
| GET    | `/health`          | Health check endpoint         |

### Example Usage

```bash
# Create a paste
curl -X POST http://localhost:3000/api/v1/paste \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "text=Hello World&extension=txt&expires=60"

# Retrieve a paste
curl http://localhost:3000/api/v1/paste/your-paste-id

# Get raw content
curl http://localhost:3000/api/v1/paste/your-paste-id/raw
```

## Observability

Wastebin includes comprehensive observability features:

- **Metrics**: Application metrics via OpenTelemetry (request counts, durations, etc.)
- **Tracing**: Distributed tracing for request flow analysis
- **Health Checks**: Built-in health endpoints for monitoring
- **Structured Logging**: JSON-formatted logs with configurable levels

Compatible with popular observability platforms like Jaeger, Prometheus, and Grafana.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

1. **Prerequisites**: Go 1.21+, Node.js 18+
2. **Clone**: `git clone https://github.com/coolguy1771/wastebin.git`
3. **Backend**: `go mod download && go run cmd/wastebin/main.go`
4. **Frontend**: `cd web && npm install && npm run dev`

### Code Quality

The project uses:
- `golangci-lint` for Go code linting
- `gofmt` for Go code formatting
- ESLint and Prettier for TypeScript/React code
- Comprehensive test coverage

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

If you find a bug or have a suggestion, please open an issue or pull request on [GitHub](https://github.com/coolguy1771/wastebin).
