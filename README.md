# DoorDarshan Kendra OSS

DoorDarshan Kendra is an open-source WebRTC meeting orchestration service that manages meeting lifecycle, participant connections, and media stream routing. This service acts as a control plane for WebRTC meetings, handling SFU (Selective Forwarding Unit) interactions and meeting state management.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Important Note: Signaling Platform](#important-note-signaling-platform)
- [Features](#features)
- [Dependencies](#dependencies)
- [Local Setup](#local-setup)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## 🎯 Overview

DoorDarshan Kendra provides a RESTful API for managing WebRTC meetings, including:
- Meeting creation and lifecycle management
- Participant join/leave operations
- Media producer/consumer management (audio/video streams)
- WebRTC transport management
- SFU cluster coordination
- Meeting capacity and resource management

## 🏗️ Architecture

```
┌─────────────────┐
│   Client Apps   │
└────────┬────────┘
         │
         │ HTTP/REST API
         ▼
┌─────────────────────────────────────┐
│   DoorDarshan Kendra (This Service) │
│  - Meeting Management               │
│  - Participant Orchestration         │
│  - SFU Coordination                  │
└────────┬────────────────────────────┘
         │
         ├──► HTTP API (Signaling Platform) [Option 1]
         │    OR
         ├──► Redis Streams (Signaling Messages) [Option 2]
         │
         ├──► MySQL (Meeting State)
         │
         └──► SFU Clusters (Media Routing)
```

### Component Interactions

1. **DoorDarshan Kendra**: This open-source service handles meeting orchestration
2. **SFU Clusters**: External SFU services that handle actual media routing
3. **MySQL**: Stores meeting metadata and state
4. **Signaling Platform**: **NOT OPEN SOURCE** - Can be integrated via:
   - **Option 1**: HTTP API endpoints (no Redis client in DoorDarshan Kendra - Signaling Platform handles Redis)
   - **Option 2**: Direct Redis Streams (Redis client required in DoorDarshan Kendra)
5. **Redis**: **OPTIONAL in DoorDarshan Kendra** - Redis client only needed if using Option 2. With Option 1, Signaling Platform handles Redis operations.

## ⚠️ Important Note: Signaling Platform

**The Signaling Platform is NOT part of this open-source repository.**

DoorDarshan Kendra communicates with the Signaling Platform using one of two approaches. **Both approaches ultimately use Redis**, but the difference is where the Redis client dependency lives:

### Integration Options

#### Option 1: HTTP API (No Redis Client in DoorDarshan Kendra)

DoorDarshan Kendra calls HTTP endpoints of the Signaling Platform. The Signaling Platform then pushes messages to Redis Streams.

- DoorDarshan Kendra makes HTTP POST requests to Signaling Platform endpoints
- **No Redis client required in DoorDarshan Kendra**
- Signaling Platform handles Redis operations

#### Option 2: Direct Redis Streams (Redis Client Required in DoorDarshan Kendra)

DoorDarshan Kendra directly publishes signaling messages to **Redis Streams**. The Signaling Platform then consumes from these streams.

- DoorDarshan Kendra publishes to Redis Streams (format: `room-stream:{meeting_id}`)
- **Redis client required in DoorDarshan Kendra**
- Signaling Platform consumes from Redis Streams
- DoorDarshan Kendra uses Redis client (see `pkg/data/redis_repository.go`)

### What This Means for Users

If you want to use this codebase, you will need to:

1. **Choose an integration approach:**
   - **Option 1**: Use HTTP API - No Redis client needed in DoorDarshan Kendra (Redis is handled by Signaling Platform)
   - **Option 2**: Use Redis Streams - Redis client required in DoorDarshan Kendra

2. **Implement your own Signaling Platform** that:
   - **For Option 1**: Exposes HTTP endpoints that DoorDarshan Kendra can call, and pushes messages to Redis Streams
   - **For Option 2**: Consumes from Redis Streams (format: `room-stream:{meeting_id}`)
   - Manages WebSocket connections with meeting participants
   - Broadcasts messages to connected clients
   - Handles the message format defined in `pkg/signaling-platform/requests.go`

3. **Or integrate with an existing signaling solution** that supports either approach

### Redis Stream Message Format (Option 2 Only)

**Note**: This section only applies if you're using direct Redis Streams (Option 2). If using HTTP API (Option 1), the Signaling Platform handles Redis operations.

When using Option 2, messages are published to Redis Streams with the key format: `room-stream:{meeting_id}`

The message structure in the stream includes:
- `name`: Signal type (e.g., "DoordarshanStreamPublish")
- `signal`: JSON string containing the signal payload
- `roomId`: Meeting/room identifier
- `participantId`: Participant identifier
- `messageType`: Always "signalMessage"
- `requestId`: Unique request identifier

**Signal Types:**
- `DoordarshanStreamPublish`: When a participant publishes a media stream
- `DoordarshanStreamUnPublish`: When a participant stops publishing
- `DoordarshanUserJoined`: When a participant joins a meeting
- `DoordarshanUserLeft`: When a participant leaves a meeting
- `DoordarshanDisconnect`: When a participant disconnects

**Implementation Reference:**
- See `pkg/handler/meeting_v1_handler.go` for how messages are published
- See `pkg/signaling-platform/requests.go` for the request structure
- See `pkg/data/redis_repository.go` for Redis stream operations (Option 2 only)

## ✨ Features

- ✅ Meeting lifecycle management (create, join, leave, end)
- ✅ Participant management with capacity controls
- ✅ Media producer/consumer management
- ✅ WebRTC transport management (create, connect, recreate)
- ✅ ICE restart capabilities
- ✅ Producer/consumer pause/resume
- ✅ Multi-tenant support
- ✅ SFU cluster abstraction
- ✅ Redis-based signaling message broadcasting
- ✅ OpenTelemetry integration for observability
- ✅ Comprehensive API documentation (Swagger UI)

## 📦 Dependencies

### System Requirements

- **Go**: 1.24.0 or higher
- **MySQL**: 5.7+ or 8.0+
- **Redis**: 6.0+ (Optional - only required if using Redis Streams for signaling)
- **Make** (optional, for build scripts)

**Note**: Redis client in DoorDarshan Kendra is **optional** and only needed if you choose Option 2 (direct Redis Streams). If you use Option 1 (HTTP API), the Signaling Platform handles Redis operations, so no Redis client is required in DoorDarshan Kendra.

### Go Dependencies

Key dependencies (see `go.mod` for complete list):

- **Echo v4**: Web framework
- **Redis Go Client v9**: Redis connectivity
- **MySQL Driver**: Database connectivity
- **Google Wire**: Dependency injection
- **Viper**: Configuration management
- **Zap**: Structured logging
- **OpenTelemetry**: Observability and tracing
- **Swaggo**: API documentation generation

## 🚀 Local Setup

### Prerequisites

1. Install Go 1.24.0 or higher: https://golang.org/dl/
2. Install MySQL: https://dev.mysql.com/downloads/
3. Install Redis: https://redis.io/download (Optional - only if using Option 2: direct Redis Streams)

### Step 1: Clone the Repository

```bash
git clone https://github.com/Allen-Digital-OSS/doordarshan-kendra-oss.git
cd doordarshan-kendra-oss
```

### Step 2: Install Dependencies

```bash
go mod download
```

### Step 3: Setup Database

Create the MySQL database:

```bash
mysql -u root -p
CREATE DATABASE doordarshan;
```

### Step 4: Configure Environment

The application uses `configs/local.env` by default when `ENV=local`. You can either:

**Option 1: Edit the existing file directly**
```bash
# Edit configs/local.env with your local settings
nano configs/local.env  # or use your preferred editor
```

**Option 2: Create a custom config file**
```bash
# Create your own config file
cp configs/local.env configs/my-local.env
# Edit my-local.env, then set ENV=my-local when running
```

Edit the configuration file with your local settings:

```env
# Server Configuration
SERVER_PORT=8000
SERVER_LOG_LEVEL=info

# MySQL Configuration
MYSQL_CONNECTION_STRING=root:password@(127.0.0.1:3306)/doordarshan

# Redis Configuration (Optional - only if using Option 2: direct Redis Streams)
# If using Option 1: HTTP API, Redis client is not required in DoorDarshan Kendra
# (Signaling Platform will handle Redis operations)
REDIS_CLUSTER_MODE_ON=false  # Set to false for single Redis instance
REDIS_CLUSTER_ADDRESSES=127.0.0.1:6379
REDIS_CLUSTER_PASSWORD=  # Leave empty if no password

# Signaling Platform (Your implementation)
# Option 1: HTTP API (no Redis client required in DoorDarshan Kendra)
# Option 2: Direct Redis Streams (requires Redis configuration above)

# SFU Configuration (Adjust based on your SFU setup)
# See pkg/sfu/ for SFU integration details
```

### Step 5: Install Pre-commit Hooks (Recommended)

```bash
./scripts/setup-hooks.sh
```

This installs hooks to prevent committing secrets. See [SECURITY.md](./SECURITY.md) for details.

### Step 6: Run the Application

```bash
# Option 1: Set environment variable (defaults to "local" if not set)
export ENV=local
go run main.go

# Option 2: Run without setting ENV (will default to "local")
go run main.go
```

The server will start on port 8000 (or your configured port). The application automatically loads the config from `configs/local.env` when `ENV=local`.

### Step 7: Verify Installation

1. **Health Check**:
   ```bash
   curl http://localhost:8000/health
   ```

2. **Swagger UI**: Open http://localhost:8000/swagger in your browser

3. **API Root**: http://localhost:8000/ (redirects to Swagger UI)

## 📁 Project Structure

```
doordarshan-kendra-oss/
├── cmd/
│   └── doordarshan-kendra/    # Application entry point and wire setup
│       ├── run.go              # Application bootstrap
│       ├── wire.go             # Dependency injection setup
│       └── wire_gen.go          # Generated wire code
│
├── configs/                     # Configuration files
│   └── local.env                # Local environment template
│
├── docs/                        # API Documentation
│   ├── swagger.yaml             # OpenAPI specification
│   ├── swagger.json             # JSON format
│   ├── swagger-ui.html          # Swagger UI interface
│   └── docs.go                  # Generated docs
│
├── pkg/
│   ├── app/                     # Application layer
│   │   └── application.go       # Application lifecycle
│   │
│   ├── clients/                 # External service clients
│   │
│   ├── common/                  # Shared utilities
│   │   ├── config.go            # Configuration structures
│   │   ├── log.go               # Logging utilities
│   │   └── util.go              # Common utilities
│   │
│   ├── constant/                # Constants
│   │   └── const.go
│   │
│   ├── data/                     # Data access layer
│   │   ├── mysql_client.go      # MySQL client
│   │   ├── redis_client.go      # Redis client
│   │   └── redis_repository.go  # Redis operations
│   │
│   ├── handler/                  # HTTP handlers
│   │   ├── handler.go           # Handler interfaces
│   │   ├── meeting_v1_handler.go  # Meeting API handlers
│   │   ├── request/             # Request DTOs
│   │   └── response/            # Response DTOs
│   │
│   ├── log/                      # Logging
│   │   └── logger.go            # Logger implementation
│   │
│   ├── myerrors/                 # Error handling
│   │   └── errors.go
│   │
│   ├── server/                   # HTTP server
│   │   ├── server.go            # Server setup
│   │   ├── routes.go            # Route definitions
│   │   └── *_middleware.go     # Middleware components
│   │
│   ├── sfu/                      # SFU integration
│   │   ├── sfu_cluster_handler.go  # SFU cluster management
│   │   └── *_generator.go      # Request/response generators
│   │
│   └── signaling-platform/       # Signaling platform models
│       └── requests.go          # Request structures
│
├── scripts/                      # Utility scripts
│   └── setup-hooks.sh           # Pre-commit hook setup
│
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── SECURITY.md                   # Security guidelines
└── README.md                     # This file
```

### Folder Documentation

For detailed documentation on specific components, refer to the README files in each folder:

- [`pkg/handler/README.md`](pkg/handler/README.md) - API handler documentation and request/response formats
- [`pkg/sfu/README.md`](pkg/sfu/README.md) - SFU integration guide and cluster management
- [`pkg/data/README.md`](pkg/data/README.md) - Data layer documentation (MySQL, Redis operations)
- [`pkg/signaling-platform/README.md`](pkg/signaling-platform/README.md) - Signaling platform integration guide
- [`docs/README.md`](docs/README.md) - API documentation generation and Swagger setup
- [`cmd/doordarshan-kendra/README.md`](cmd/doordarshan-kendra/README.md) - Application bootstrap and dependency injection
- [`docs/system-design.md`](docs/system-design.md) - System design documentation and architecture decisions

## 📚 API Documentation

### Interactive Documentation

Once the server is running, access the Swagger UI at:
- **Swagger UI**: http://localhost:8000/swagger
- **Swagger YAML**: http://localhost:8000/swagger.yaml
- **Swagger JSON**: http://localhost:8000/swagger.json

### API Endpoints

#### Meeting Management
- `POST /v1/createMeeting` - Create a new meeting
- `POST /v1/joinMeeting` - Join an existing meeting
- `POST /v1/leaveMeeting` - Leave a meeting
- `POST /v1/endMeeting` - End a meeting
- `POST /v1/preMeetingDetails` - Get pre-meeting details
- `POST /v1/activeContainer` - Get active container for a meeting

#### Participant Management
- `POST /v1/getProducersOfMeeting` - Get all producers in a meeting
- `POST /v1/getRTPCapabilities` - Get RTP capabilities

#### Media Producer Management
- `POST /v1/createProducer` - Create a media producer
- `POST /v1/pauseProducer` - Pause a producer
- `POST /v1/resumeProducer` - Resume a producer
- `POST /v1/closeProducer` - Close a producer

#### Media Consumer Management
- `POST /v1/createConsumer` - Create a media consumer
- `POST /v1/pauseConsumer` - Pause a consumer
- `POST /v1/resumeConsumer` - Resume a consumer
- `POST /v1/closeConsumer` - Close a consumer

#### Transport Management
- `POST /v1/connectProducerTransport` - Connect producer transport
- `POST /v1/connectConsumerTransport` - Connect consumer transport
- `POST /v1/recreateProducerTransport` - Recreate producer transport
- `POST /v1/recreateConsumerTransport` - Recreate consumer transport
- `POST /v1/restartIce` - Restart ICE for all transports
- `POST /v1/restartProducerIce` - Restart producer ICE
- `POST /v1/restartConsumerIce` - Restart consumer ICE

#### Health
- `GET /health` - Health check endpoint

See the Swagger UI for detailed request/response schemas and examples.

## ⚙️ Configuration

Configuration is managed through environment variables. See `configs/local.env` for all available options.

### Key Configuration Sections

- **Server**: Port, timeouts, CORS, logging
- **MySQL**: Connection string
- **Redis**: Cluster settings, connection pool, timeouts (Optional - only for Option 2: direct Redis Streams)
- **Signaling Platform**: Endpoint and timeout (for HTTP API approach)
- **OpenTelemetry**: Tracing and metrics export
- **Rate Limiting**: API rate limits

## 🤝 Contributing
We follow a fork-based contribution model.

Workflow:
1.	Fork the repository
2.	Create a feature branch in your fork
3.	Make your changes
4.	Run pre-commit hooks
5.	Open a Pull Request against main

Direct pushes to the main repository are disabled.

## 🔒 Security

This project uses automated security scanning to prevent secrets from being committed:

- **Gitleaks**: Scans for secrets in commits
- **Pre-commit hooks**: Runs checks before each commit
- **GitHub Actions**: Enforces checks on all pushes/PRs

See [SECURITY.md](./SECURITY.md) for detailed security guidelines.

## 📄 License

Please check the [LICENSE](LICENSE) file in the repository root for the full license text.

## 🙏 Acknowledgments

- Built with [Echo](https://echo.labstack.com/) web framework
- Uses [Google Wire](https://github.com/google/wire) for dependency injection
- Observability powered by [OpenTelemetry](https://opentelemetry.io/)

## 📞 Support

- **Issues**: Open an issue on GitHub
- **Questions**: See existing issues or open a new one

---

**Note**: Remember that you'll need to implement or integrate a Signaling Platform service. You can choose between:
- **Option 1: HTTP API**: No Redis client required in DoorDarshan Kendra (Signaling Platform handles Redis)
- **Option 2: Direct Redis Streams**: Redis client required in DoorDarshan Kendra

The Signaling Platform should manage WebSocket connections and broadcast messages to meeting participants. Both approaches ultimately use Redis, but Option 1 moves the Redis dependency to the Signaling Platform.
