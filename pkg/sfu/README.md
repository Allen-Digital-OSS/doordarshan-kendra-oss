# SFU Package

This package provides the abstraction layer for SFU (Selective Forwarding Unit) cluster management. It handles container allocation, router management, and meeting capacity calculations.

## Overview

The SFU package abstracts the interaction with SFU clusters, providing a unified interface for:
- Container allocation and management
- Router ID allocation for producers and consumers
- Meeting capacity management
- Container monitoring and health checks

## Architecture

```
┌─────────────────────┐
│  Meeting Handler    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  IClusterHandler    │  ← Interface
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ ClusterHandler     │  ← Implementation
│ (Tenant-specific)   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   MySQL Database    │  ← Container/Router metadata
└─────────────────────┘
```

## Key Components

### IClusterHandler Interface

The `IClusterHandler` interface defines all SFU cluster operations:

```go
type IClusterHandler interface {
    GetTenant() string
    SetDb(db *sql.DB)
    GetDb() *sql.DB
    GetFreeContainer(ctx context.Context, request FreeContainerRequest) (*FreeContainersResponse, error)
    AllocateProducerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, string, []string, []string, error)
    AllocateConsumerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, string, []string, []string, error)
    // ... and more
}
```

### BaseClusterHandler

The `BaseClusterHandler` provides common functionality for all cluster handlers:

- Database query operations
- Container monitoring
- Common utility functions

### Tenant-Specific Handlers

Each tenant can have its own cluster handler implementation. The default handler is `TenantDefaultClusterHandler` located in `tenant_default_cluster_handler.go`.

## Core Operations

### 1. Container Allocation

**GetFreeContainer**: Finds an available container with sufficient capacity

```go
type FreeContainerRequest struct {
    Capacity         int  // Overall meeting capacity
    ProducerCapacity int  // Producer capacity
    ConsumerCapacity int  // Consumer capacity
}

type FreeContainersResponse struct {
    ContainerId            string
    Host                   string
    WorkerCapacity         *WorkerCapacity
    ProducerWorkerCapacity *WorkerCapacity
    ConsumerWorkerCapacity []WorkerCapacity
}
```

### 2. Router Allocation

**AllocateProducerRouterIdForMeetingAndParticipantId**: Allocates a router for a participant's producer streams

**AllocateConsumerRouterIdForMeetingAndParticipantId**: Allocates a router for a participant's consumer streams

**AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId**: Allocates both producer and consumer routers

### 3. Container Monitoring

**MonitorContainers**: Monitors container health and heartbeat status

Returns:
- `containerHeartbeatMap`: Map of container ID to last heartbeat timestamp
- `containerMeetingsMap`: Map of container ID to list of meeting IDs
- `containerTenantMap`: Map of container ID to tenant

### 4. Meeting Configuration

**MeetingConfiguration**: Gets the configuration for a meeting in a specific container

Returns:
- Worker configurations with capacity information
- List of worker IDs

## Database Schema

The SFU package interacts with MySQL tables:

- `worker_container`: Maps workers to containers
- `router_worker`: Maps routers to workers
- `router_meeting`: Maps routers to meetings
- `container_heartbeat`: Container health monitoring

## Capacity Management

### Capacity Calculation

The `GetCalculatedMeetingCapacity` function calculates meeting capacity based on:
- Overall meeting capacity
- Producer capacity
- Consumer capacity

### Worker Capacity

Each worker has:
- `Capacity`: Overall capacity
- `ProducerCapacity`: Maximum producers
- `ConsumerCapacity`: Maximum consumers

## Request/Response Generators

### SFU Request Generator

Located in `sfu_request_generator.go`, generates SFU-specific requests for:
- Creating routers
- Creating transports
- Creating producers/consumers
- ICE restart operations

### SFU Response Generator

Located in `sfu_response_generator.go`, processes SFU responses and extracts:
- Router IDs
- Transport IDs
- Producer/consumer IDs
- ICE candidates

## Implementation Example

### Creating a Custom Cluster Handler

```go
type CustomTenantHandler struct {
    *BaseClusterHandler
    tenant string
}

func (c *CustomTenantHandler) GetTenant() string {
    return c.tenant
}

func (c *CustomTenantHandler) GetFreeContainer(
    ctx context.Context,
    request FreeContainerRequest,
) (*FreeContainersResponse, error) {
    // Custom implementation
    // ...
}
```

### Registering in main.go

```go
tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
    "default": &sfu.TenantDefaultClusterHandler{},
    "custom":   &CustomTenantHandler{tenant: "custom"},
}
```

## Container Lifecycle

1. **Allocation**: Container is allocated when a meeting is created
2. **Active**: Container hosts meetings and participants
3. **Monitoring**: Heartbeat is checked periodically
4. **Replacement**: Containers can be replaced if unhealthy

## Error Handling

Common errors:
- **Container not found**: No available container with required capacity
- **Database errors**: Connection or query failures
- **Capacity exceeded**: Requested capacity exceeds available resources

## Configuration

SFU configuration is managed through:
- Database tables (container, worker, router metadata)
- Environment variables (for connection settings)
- Tenant-specific handler implementations

## Testing

Run SFU package tests:

```bash
go test ./pkg/sfu/...
```

## Best Practices

1. **Container Selection**: Always check capacity before allocation
2. **Router Management**: Ensure proper cleanup when participants leave
3. **Monitoring**: Regularly check container health
4. **Error Recovery**: Implement retry logic for transient failures
5. **Capacity Planning**: Monitor and adjust capacity based on usage

## Related Documentation

- [Handler Package](../handler/README.md) - API handlers that use SFU
- [Data Layer](../data/README.md) - Database operations
- [Application Bootstrap](../../cmd/doordarshan-kendra/README.md) - How handlers are registered

## Troubleshooting

### Container Allocation Fails

- Check database connectivity
- Verify container capacity in database
- Check container heartbeat status

### Router Allocation Issues

- Verify meeting exists in database
- Check router-worker mappings
- Ensure sufficient worker capacity

### Capacity Calculation Errors

- Verify input parameters
- Check database for current capacity usage
- Review capacity calculation logic
