# Handler Package

This package contains HTTP handlers for the DoorDarshan Kendra API. It implements the business logic for meeting management, participant operations, and media stream control.

## Overview

The handler package provides RESTful API endpoints for:
- Meeting lifecycle management (create, join, leave, end)
- Participant operations
- Media producer/consumer management
- WebRTC transport management
- Signaling message broadcasting

## Structure

```
pkg/handler/
├── handler.go                    # Base handler interfaces and utilities
├── meeting_v1_handler.go         # Main meeting API handlers
├── meeting_handler_test.go      # Handler tests
├── request/                      # Request DTOs
│   ├── request.go               # Base request structures
│   └── meeting_v1_request.go   # Meeting-specific requests
└── response/                     # Response DTOs
    ├── meeting_response.go      # Meeting responses
    ├── participant_response.go  # Participant responses
    └── participant_connection_response.go  # Connection responses
```

## Key Components

### Handler Interface

The `IMeetingV1Handler` interface defines all meeting-related operations:

```go
type IMeetingV1Handler interface {
    CreateMeeting(c echo.Context) error
    JoinMeeting(c echo.Context) error
    LeaveMeeting(c echo.Context) error
    EndMeeting(c echo.Context) error
    // ... and more
}
```

### Request/Response Models

All request and response structures are defined in the `request/` and `response/` subdirectories:

- **Request Models**: Located in `pkg/handler/request/`
  - `CreateMeetingRequestV1`: Meeting creation parameters
  - `JoinMeetingRequestV1`: Participant join parameters
  - `CreateProducerRequestV1`: Producer creation parameters
  - `CreateConsumerRequestV1`: Consumer creation parameters
  - And more...

- **Response Models**: Located in `pkg/handler/response/`
  - `MeetingResponse`: Meeting information
  - `ParticipantResponse`: Participant details
  - `ParticipantConnectionResponse`: Connection details

## API Endpoints

### Meeting Management

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| POST | `/v1/createMeeting` | `CreateMeeting` | Create a new meeting |
| POST | `/v1/joinMeeting` | `JoinMeeting` | Join an existing meeting |
| POST | `/v1/leaveMeeting` | `LeaveMeeting` | Leave a meeting |
| POST | `/v1/endMeeting` | `EndMeeting` | End a meeting |
| POST | `/v1/preMeetingDetails` | `PreMeetingDetails` | Get pre-meeting details |
| POST | `/v1/activeContainer` | `GetActiveContainerOfMeeting` | Get active container |

### Media Producer Management

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| POST | `/v1/createProducer` | `CreateProducer` | Create a media producer |
| POST | `/v1/pauseProducer` | `PauseProducer` | Pause a producer |
| POST | `/v1/resumeProducer` | `ResumeProducer` | Resume a producer |
| POST | `/v1/closeProducer` | `CloseProducer` | Close a producer |

### Media Consumer Management

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| POST | `/v1/createConsumer` | `CreateConsumer` | Create a media consumer |
| POST | `/v1/pauseConsumer` | `PauseConsumer` | Pause a consumer |
| POST | `/v1/resumeConsumer` | `ResumeConsumer` | Resume a consumer |
| POST | `/v1/closeConsumer` | `CloseConsumer` | Close a consumer |

### Transport Management

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| POST | `/v1/connectProducerTransport` | `ConnectProducerTransport` | Connect producer transport |
| POST | `/v1/connectConsumerTransport` | `ConnectConsumerTransport` | Connect consumer transport |
| POST | `/v1/recreateProducerTransport` | `RecreateProducerTransport` | Recreate producer transport |
| POST | `/v1/recreateConsumerTransport` | `RecreateConsumerTransport` | Recreate consumer transport |
| POST | `/v1/restartIce` | `RestartIce` | Restart ICE for all transports |
| POST | `/v1/restartProducerIce` | `RestartProducerIce` | Restart producer ICE |
| POST | `/v1/restartConsumerIce` | `RestartConsumerIce` | Restart consumer ICE |

### Query Operations

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| POST | `/v1/getProducersOfMeeting` | `GetProducersOfMeeting` | Get all producers in a meeting |
| POST | `/v1/getRTPCapabilities` | `GetRTPCapabilities` | Get RTP capabilities |

### Health Check

| Method | Endpoint | Handler | Description |
|--------|----------|---------|-------------|
| GET | `/health` | `HealthProbe` | Health check endpoint |

## Request/Response Formats

### Example: Create Meeting Request

```json
{
  "meeting_id": "meeting-123",
  "tenant": "default",
  "meeting_capacity": 100,
  "producer_meeting_capacity": 50,
  "consumer_meeting_capacity": 100
}
```

### Example: Join Meeting Request

```json
{
  "meeting_id": "meeting-123",
  "participant_id": "participant-456",
  "instance_id": "instance-789",
  "tenant": "default"
}
```

### Example: Create Producer Request

```json
{
  "meeting_id": "meeting-123",
  "participant_id": "participant-456",
  "instance_id": "instance-789",
  "kind": "audio",
  "rtp_parameters": [...],
  "tenant": "default",
  "start_recording": false
}
```

## Signaling Integration

The handlers send signaling messages to the Signaling Platform using one of two approaches. **Both approaches ultimately use Redis**, but the difference is where the Redis client dependency lives:

### Option 1: HTTP API (No Redis Client in DoorDarshan Kendra)
- Handlers make HTTP POST requests to Signaling Platform endpoints
- **No Redis client required in DoorDarshan Kendra**
- Signaling Platform handles Redis operations
- Direct synchronous communication

### Option 2: Direct Redis Streams (Redis Client Required in DoorDarshan Kendra)
- Handlers publish messages directly to Redis Streams
- **Redis client required in DoorDarshan Kendra**
- Asynchronous messaging
- See `pkg/data/redis_repository.go` for Redis operations

**Signal types include:**
- `DoordarshanStreamPublish`: When a participant publishes a stream
- `DoordarshanStreamUnPublish`: When a participant stops publishing
- `DoordarshanUserJoined`: When a participant joins
- `DoordarshanUserLeft`: When a participant leaves
- `DoordarshanDisconnect`: When a participant disconnects

See `pkg/signaling-platform/README.md` for more details on signaling message format and integration options.

## Error Handling

Handlers use the `HandleCommonError` function for consistent error responses:

```go
func HandleCommonError(err error) error {
    if errors.Is(err, echo.ErrUnsupportedMediaType) {
        return echo.NewHTTPError(http.StatusUnsupportedMediaType, "only application/json content type is supported")
    }
    return echo.NewHTTPError(http.StatusBadRequest, err.Error())
}
```

## Rate Limiting

Rate limiting is configured per endpoint. Default rate limit is 10 QPS (configurable via `RATE_LIMIT_DEFAULT_API_RATE_LIMIT_QPS`).

## Dependencies

- **Echo v4**: HTTP framework
- **Validator v10**: Request validation
- **Zap**: Structured logging
- **SFU Package**: SFU cluster management
- **Data Package**: Database operations (MySQL required, Redis client optional)
- **Redis Repository**: Redis operations (Option 2, only if using direct Redis Streams)

## Testing

Run handler tests:

```bash
go test ./pkg/handler/...
```

See `meeting_handler_test.go` for example test cases.

## Adding New Handlers

1. Define request/response DTOs in `request/` and `response/` directories
2. Implement handler method in `meeting_v1_handler.go`
3. Add route in `pkg/server/routes.go`
4. Update Swagger documentation in `docs/swagger.yaml`
5. Add tests in `meeting_handler_test.go`

## Related Documentation

- [API Documentation](../../docs/README.md) - Swagger/OpenAPI documentation
- [SFU Integration](../sfu/README.md) - SFU cluster management
- [Data Layer](../data/README.md) - Database and Redis operations
- [Signaling Platform](../signaling-platform/README.md) - Signaling message format
