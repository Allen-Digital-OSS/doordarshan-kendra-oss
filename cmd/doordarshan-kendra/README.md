# Application Bootstrap

This directory contains the application entry point and dependency injection setup using Google Wire.

## Overview

The `cmd/doordarshan-kendra` package handles:
- Application initialization
- Dependency injection (using Google Wire)
- Configuration loading
- Service startup

## Structure

```
cmd/doordarshan-kendra/
├── run.go        # Application bootstrap and startup
├── wire.go       # Wire dependency injection setup
└── wire_gen.go   # Generated Wire code (auto-generated)
```

## Entry Point

The application entry point is in `main.go` at the project root:

```go
func main() {
    envToConfigPathMap := map[string]string{
        "local": "configs/local.env",
    }
    tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
        "default": &sfu.TenantDefaultClusterHandler{},
    }
    cmd.Run(envToConfigPathMap, tenantClusterHandlerMap)
}
```

## Bootstrap Process

### 1. Environment Detection

The `Run` function determines the environment:

```go
environment := common.GetCurrentEnvironment()
if environment == "" || !ok {
    environment = "local"  // Default
}
```

Environment is set via `ENV` environment variable:
```bash
export ENV=local
```

### 2. Configuration Loading

Configuration is loaded from environment files:

```go
appConfig, err := common.LoadAppConfigFromEnv(configPath, configName)
```

Configuration files are located in `configs/`:
- `configs/local.env` - Local development
- Add more environments as needed

### 3. Database Initialization

MySQL client is created and passed to SFU cluster handlers:

```go
mySQLClient := data.NewMySQLClient(appConfig)
for _, v := range tenantClusterHandlerMap {
    v.SetDb(mySQLClient.MysqlDb)
}
```

### 4. OpenTelemetry Setup

Metrics and tracing are initialized:

```go
meter := common.InitializeOpenTelemetry(appConfig)
```

### 5. Dependency Injection

Application components are wired together using Google Wire:

```go
application, err := InitializeApp(appConfig, meter, tenantClusterHandlerMap)
```

## Dependency Injection with Wire

### Wire Setup

Wire is configured in `wire.go`:

```go
//go:build wireinject
// +build wireinject

func InitializeApp(
    appConfig *common.AppConfig,
    metric *common.Meter,
    tenantClusterHandlerMap map[string]sfu.IClusterHandler,
) (*app.Application, error) {
    wire.Build(
        clients.ProviderSet,    // Signaling platform client
        data.ProviderSet,        // MySQL, Redis clients
        handler.ProviderSet,    // HTTP handlers
        server.ProviderSet,      // HTTP server
        app.NewApplication,      // Application
    )
    return &app.Application{}, nil
}
```

### Provider Sets

Each package provides a `ProviderSet` that Wire uses:

- **`clients.ProviderSet`**: External service clients
- **`data.ProviderSet`**: Data access layer (MySQL, Redis)
- **`handler.ProviderSet`**: HTTP handlers
- **`server.ProviderSet`**: HTTP server and middleware

### Generating Wire Code

Wire generates `wire_gen.go` automatically. To regenerate:

```bash
# Install Wire
go install github.com/google/wire/cmd/wire@latest

# Generate code
cd cmd/doordarshan-kendra
wire
```

## Tenant Cluster Handlers

The application supports multiple tenants, each with its own SFU cluster handler:

```go
tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
    "default": &sfu.TenantDefaultClusterHandler{},
    // Add more tenants as needed
}
```

### Adding a New Tenant

1. Implement `sfu.IClusterHandler` interface
2. Add to `tenantClusterHandlerMap` in `main.go`
3. Configure tenant-specific settings

Example:
```go
tenantClusterHandlerMap := map[string]sfu.IClusterHandler{
    "default": &sfu.TenantDefaultClusterHandler{},
    "enterprise": &sfu.EnterpriseClusterHandler{},
}
```

## Configuration Mapping

Environment to configuration file mapping:

```go
envToConfigPathMap := map[string]string{
    "local": "configs/local.env",
    "dev":   "configs/dev.env",
    "prod":  "configs/prod.env",
}
```

Add new environments by:
1. Creating config file in `configs/`
2. Adding entry to `envToConfigPathMap`
3. Setting `ENV` environment variable

## Application Lifecycle

### Startup Sequence

1. Parse command-line flags
2. Detect environment
3. Load configuration
4. Initialize MySQL client
5. Initialize OpenTelemetry
6. Wire dependencies (Wire)
7. Start application

### Application Start

The `application.Start()` method:
- Initializes middleware
- Sets up routes
- Starts HTTP server
- Handles graceful shutdown

## Error Handling

### Startup Errors

Startup errors cause panic (fail-fast):
- Configuration loading failures
- Database connection failures
- Dependency injection failures

### Runtime Errors

Runtime errors are handled by:
- HTTP middleware (error responses)
- Logging (error logs)
- Graceful degradation where possible

## Adding New Dependencies

To add a new dependency:

1. **Create Provider Function**:
   ```go
   func NewMyService(config *common.AppConfig) *MyService {
       return &MyService{config: config}
   }
   ```

2. **Add to ProviderSet**:
   ```go
   var ProviderSet = wire.NewSet(NewMyService)
   ```

3. **Add to Wire Build**:
   ```go
   wire.Build(
       // ... existing providers
       mypackage.ProviderSet,
   )
   ```

4. **Regenerate Wire Code**:
   ```bash
   wire
   ```

## Testing

### Unit Tests

Test individual components:

```bash
go test ./cmd/doordarshan-kendra/...
```

### Integration Tests

Test full application startup:

```bash
# Set test environment
export ENV=test
go test ./...
```

## Environment Variables

Key environment variables:

- `ENV`: Environment name (default: "local")
- Configuration variables (see `configs/local.env`)

## Best Practices

1. **Fail Fast**: Panic on startup errors
2. **Configuration**: Use environment-based config files
3. **Dependency Injection**: Use Wire for all dependencies
4. **Error Handling**: Log errors with context
5. **Graceful Shutdown**: Handle SIGTERM/SIGINT signals

## Troubleshooting

### Wire Generation Fails

- Check for circular dependencies
- Verify all providers are exported
- Ensure build tags are correct (`//go:build wireinject`)

### Configuration Not Loading

- Verify `ENV` variable is set
- Check config file exists in `configs/`
- Verify file path in `envToConfigPathMap`

### Database Connection Issues

- Check MySQL is running
- Verify connection string format
- Check credentials are correct

### Dependency Injection Errors

- Review Wire error messages
- Check provider function signatures
- Verify all dependencies are provided

## Related Documentation

- [Google Wire Documentation](https://github.com/google/wire)
- [Application Package](../pkg/app/README.md) - Application lifecycle
- [Server Package](../pkg/server/README.md) - HTTP server setup
- [Main README](../../README.md) - Project overview
