# Data Package

This package provides data access layer abstractions for MySQL and Redis operations. It handles database connections, connection pooling, and data persistence.

## Overview

The data package provides:
- **MySQL Client**: Database connection and query operations
- **Redis Client**: Redis cluster connection and stream operations
- **Redis Repository**: High-level Redis operations (streams, pub/sub)

## Structure

```
pkg/data/
├── mysql_client.go      # MySQL connection and client
├── redis_client.go      # Redis cluster client
└── redis_repository.go  # Redis repository with stream operations
```

## MySQL Client

### Overview

The `MySQLClient` manages MySQL database connections with connection pooling.

### Configuration

MySQL connection is configured via environment variables:

```env
MYSQL_CONNECTION_STRING=root:password@(127.0.0.1:3306)/doordarshan
# OR
MYSQL_ENDPOINT=127.0.0.1
MYSQL_PORT=3306
MYSQL_DB_NAME=doordarshan
MYSQL_USERNAME=root
MYSQL_PASSWORD=password
MYSQL_SECRET_FILE=/path/to/credentials/file  # Optional
MYSQL_OPEN_CONNECTIONS=100
```

### Connection String Format

```
username:password@tcp(endpoint:port)/database?charset=utf8mb4&parseTime=True&loc=Local
```

### Usage

```go
// Create MySQL client
mysqlClient := data.NewMySQLClient(appConfig)

// Access database
db := mysqlClient.MysqlDb

// Execute query
rows, err := db.QueryContext(ctx, "SELECT * FROM meetings WHERE meeting_id = ?", meetingID)
```

### Features

- **Connection Pooling**: Configurable max open connections
- **Credential Management**: Supports credential files or direct config
- **Connection String**: Flexible connection string format
- **Error Handling**: Panics on connection failure (startup safety)

## Redis Client

### Overview

The `RedisClient` manages Redis cluster connections with support for:
- Redis Cluster mode
- Single Redis instance (when cluster mode is off)
- TLS/SSL connections
- Authentication (username/password)
- Connection pooling

### Configuration

Redis connection is configured via environment variables:

```env
REDIS_CLUSTER_MODE_ON=true
REDIS_CLUSTER_ADDRESSES=127.0.0.1:7000,127.0.0.1:7001,127.0.0.1:7002
REDIS_CLUSTER_READ_TIMEOUT_IN_SECS=5
REDIS_CLUSTER_WRITE_TIMEOUT_IN_SECS=5
REDIS_CLUSTER_DIAL_TIMEOUT_IN_SECS=5
REDIS_CLUSTER_POOL_TIMEOUT_IN_SECS=60
REDIS_CLUSTER_POOL_SIZE=1000
REDIS_CLUSTER_MIN_IDLE_CONNECTIONS=100
REDIS_CLUSTER_MAX_IDLE_CONNECTIONS=1000
REDIS_CLUSTER_MAX_ACTIVE_CONNECTIONS=10000
REDIS_CLUSTER_USERNAME=  # Optional
REDIS_CLUSTER_PASSWORD=  # Optional
REDIS_CLUSTER_SECRET_FILE=/path/to/credentials  # Optional
REDIS_CLUSTER_TLS=false
REDIS_CLUSTER_SKIP_VERIFY=true
```

### Usage

```go
// Create Redis client
redisClient := data.NewRedisClient(appConfig)

// Access universal client
client := redisClient.ruc

// Use Redis operations
result := client.Ping(ctx)
```

### Features

- **Cluster Support**: Full Redis Cluster support
- **Single Instance**: Falls back to single instance when cluster mode is off
- **TLS Support**: Optional TLS/SSL encryption
- **Connection Pooling**: Configurable pool sizes and timeouts
- **Health Checks**: Automatic ping on connection

## Redis Repository

### Overview

The `RedisRepository` provides high-level operations for Redis Streams, which are used for signaling message broadcasting.

### Stream Operations

#### PushToStream

Publishes a message to a Redis Stream:

```go
type RedisRepository interface {
    PushToStream(
        ctx context.Context,
        stream string,
        values map[string]interface{},
    ) (string, error)
}
```

### Usage Example

```go
// Create repository
redisRepo := data.NewRedisRepository(redisClient, appConfig)

// Push message to stream
messageID, err := redisRepo.PushToStream(
    ctx,
    "room-stream:meeting-123",
    map[string]interface{}{
        "name": "DoordarshanStreamPublish",
        "signal": signalJSON,
        "roomId": "meeting-123",
        "participantId": "participant-456",
        "messageType": "signalMessage",
    },
)
```

### Stream Key Format

Streams use the format: `room-stream:{meeting_id}`

Example: `room-stream:meeting-123`

## Message Format

Messages published to Redis Streams follow this structure:

```go
map[string]interface{}{
    "name":          "SignalType",           // Signal type name
    "signal":        signalJSONString,       // JSON string of signal payload
    "roomId":        meetingID,              // Meeting/room ID
    "participantId": participantID,          // Participant ID
    "messageType":   "signalMessage",        // Always "signalMessage"
    "requestId":     requestID,              // Unique request ID
}
```

## Signal Types

The following signal types are published to Redis Streams:

- `DoordarshanStreamPublish`: Participant publishes a stream
- `DoordarshanStreamUnPublish`: Participant stops publishing
- `DoordarshanUserJoined`: Participant joins meeting
- `DoordarshanUserLeft`: Participant leaves meeting
- `DoordarshanDisconnect`: Participant disconnects

## Credential Management

### Credential Files

Both MySQL and Redis support credential files for secure credential storage:

```go
// Format: username:password
// Example file content:
// myuser:mypassword
```

Set via environment variables:
- `MYSQL_SECRET_FILE=/path/to/mysql/creds`
- `REDIS_CLUSTER_SECRET_FILE=/path/to/redis/creds`

### Direct Configuration

Alternatively, credentials can be provided directly:
- `MYSQL_USERNAME` / `MYSQL_PASSWORD`
- `REDIS_CLUSTER_USERNAME` / `REDIS_CLUSTER_PASSWORD`

## Connection Lifecycle

### Startup

1. Load configuration from environment
2. Read credentials (file or direct)
3. Create connection pool
4. Test connection (ping)
5. Panic on failure (fail-fast)

### Runtime

- Connections are pooled and reused
- Automatic reconnection on failure
- Health checks via ping operations

### Shutdown

- Connections are closed gracefully
- Pending operations complete
- Pool is drained

## Error Handling

### Connection Errors

- **Startup failures**: Panic (application cannot start without DB)
- **Runtime failures**: Return errors to caller
- **Network issues**: Automatic retry (handled by client libraries)

### Common Errors

- `ErrConfigNil`: Configuration is missing
- `ErrClientNil`: Client creation failed
- `ErrPingingRedisCluster`: Redis connection test failed

## Best Practices

1. **Connection Pooling**: Configure appropriate pool sizes based on load
2. **Timeouts**: Set reasonable timeouts for read/write operations
3. **Credential Security**: Use credential files in production
4. **Health Checks**: Monitor connection health
5. **Error Handling**: Always check errors from database operations
6. **Context Usage**: Always pass context for cancellation/timeout support

## Testing

Run data package tests:

```bash
go test ./pkg/data/...
```

### Test Setup

For testing, you'll need:
- MySQL instance (or use testcontainers)
- Redis instance (or use testcontainers)

## Monitoring

### Metrics to Monitor

- Connection pool usage
- Query latency
- Connection errors
- Stream message throughput
- Redis stream lag

### Logging

The package logs:
- Connection attempts
- Connection success/failure
- Credential file usage
- Pool statistics

## Troubleshooting

### MySQL Connection Issues

1. **Check credentials**: Verify username/password
2. **Network**: Ensure MySQL is reachable
3. **Database exists**: Verify database name is correct
4. **Permissions**: Check user has required permissions

### Redis Connection Issues

1. **Cluster mode**: Verify `REDIS_CLUSTER_MODE_ON` setting
2. **Addresses**: Check all cluster node addresses
3. **TLS**: Verify TLS configuration if enabled
4. **Authentication**: Check username/password if required

### Stream Publishing Issues

1. **Stream key format**: Verify `room-stream:{meeting_id}` format
2. **Message format**: Check all required fields are present
3. **Redis connectivity**: Ensure Redis is accessible
4. **Permissions**: Verify Redis user has stream write permissions

## Related Documentation

- [Handler Package](../handler/README.md) - Uses Redis Repository for signaling
- [Signaling Platform](../signaling-platform/README.md) - Message format details
- [Application Bootstrap](../../cmd/doordarshan-kendra/README.md) - How clients are initialized
