# API Documentation

This directory contains the API documentation for DoorDarshan Kendra in OpenAPI/Swagger format.

## Files

- **`swagger.yaml`**: OpenAPI 2.0 specification in YAML format
- **`swagger.json`**: OpenAPI 2.0 specification in JSON format
- **`swagger-ui.html`**: Standalone Swagger UI HTML page
- **`docs.go`**: Generated Go code from Swagger specification (used by swaggo)

## Viewing Documentation

### Option 1: Swagger UI (Recommended)

When the server is running, access the interactive Swagger UI:

1. Start the server:
   ```bash
   go run main.go
   ```

2. Open in browser:
   - **Swagger UI**: http://localhost:8000/swagger
   - **Root URL**: http://localhost:8000/ (redirects to Swagger UI)

The Swagger UI provides:
- Interactive API documentation
- Try-it-out functionality
- Request/response schemas
- Example values

### Option 2: Direct File Access

You can also view the raw specification files:
- **YAML**: http://localhost:8000/swagger.yaml
- **JSON**: http://localhost:8000/swagger.json

### Option 3: External Tools

Use external tools to view the specification:

- **Swagger Editor**: https://editor.swagger.io/
  - Copy contents of `swagger.yaml` or `swagger.json`
  - Paste into the editor

- **Postman**: Import `swagger.json` to generate a Postman collection

- **Insomnia**: Import `swagger.json` to generate requests

## API Endpoints

The documentation includes all API endpoints organized by tags:

### Health
- `GET /health` - Health check endpoint

### Meeting
- `POST /v1/createMeeting` - Create a new meeting
- `POST /v1/joinMeeting` - Join an existing meeting
- `POST /v1/leaveMeeting` - Leave a meeting
- `POST /v1/endMeeting` - End a meeting
- `POST /v1/preMeetingDetails` - Get pre-meeting details
- `POST /v1/activeContainer` - Get active container for a meeting
- `POST /v1/getProducersOfMeeting` - Get all producers in a meeting
- `POST /v1/getRTPCapabilities` - Get RTP capabilities

### Producer
- `POST /v1/createProducer` - Create a media producer
- `POST /v1/pauseProducer` - Pause a producer
- `POST /v1/resumeProducer` - Resume a producer
- `POST /v1/closeProducer` - Close a producer

### Consumer
- `POST /v1/createConsumer` - Create a media consumer
- `POST /v1/pauseConsumer` - Pause a consumer
- `POST /v1/resumeConsumer` - Resume a consumer
- `POST /v1/closeConsumer` - Close a consumer

### Transport
- `POST /v1/connectProducerTransport` - Connect producer transport
- `POST /v1/connectConsumerTransport` - Connect consumer transport
- `POST /v1/recreateProducerTransport` - Recreate producer transport
- `POST /v1/recreateConsumerTransport` - Recreate consumer transport
- `POST /v1/restartIce` - Restart ICE for all transports
- `POST /v1/restartProducerIce` - Restart producer ICE
- `POST /v1/restartConsumerIce` - Restart consumer ICE

## Request/Response Schemas

All request and response schemas are defined in the `definitions` section of the Swagger specification.

### Request Schemas

- `request.CreateMeetingRequestV1`
- `request.JoinMeetingRequestV1`
- `request.LeaveMeetingRequest`
- `request.EndMeetingRequest`
- `request.CreateProducerRequestV1`
- `request.CreateConsumerRequestV1`
- `request.ConnectTransportRequestV1`
- And more...

### Response Schemas

- `response.GetActiveContainerOfMeetingResponse`
- And more (most endpoints return generic objects)

## Updating Documentation

### Manual Updates

1. Edit `swagger.yaml` directly
2. Convert to JSON (optional):
   ```bash
   # Using yq (install: brew install yq)
   yq eval -o=json swagger.yaml > swagger.json
   ```

### Using Swaggo (Future)

If you want to generate documentation from code comments:

1. Install swaggo:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```

2. Add Swagger annotations to handlers:
   ```go
   // @Summary      Create a new meeting
   // @Description  Creates a new meeting with the specified meeting ID and capacity
   // @Tags         Meeting
   // @Accept       json
   // @Produce      json
   // @Param        request body request.CreateMeetingRequestV1 true "Create Meeting Request"
   // @Success      200  {object}  response.MeetingResponse
   // @Failure      400  {object}  map[string]interface{}
   // @Router       /v1/createMeeting [post]
   func (h *Handler) CreateMeeting(c echo.Context) error {
       // ...
   }
   ```

3. Generate documentation:
   ```bash
   swag init -g main.go -o docs
   ```

## Swagger UI Customization

The `swagger-ui.html` file uses Swagger UI from CDN. To customize:

1. Edit `swagger-ui.html`
2. Modify SwaggerUIBundle configuration:
   ```javascript
   const ui = SwaggerUIBundle({
       url: "/swagger.yaml",
       dom_id: '#swagger-ui',
       deepLinking: true,
       presets: [
           SwaggerUIBundle.presets.apis,
           SwaggerUIStandalonePreset
       ],
       // Add custom configuration here
   });
   ```

## OpenAPI Specification Version

This project uses **OpenAPI 2.0** (Swagger 2.0). The specification is defined at the top of `swagger.yaml`:

```yaml
swagger: "2.0"
info:
  contact: {}
```

## Integration with Server

The Swagger UI is integrated into the Echo server via routes in `pkg/server/routes.go`:

- `GET /swagger` - Serves Swagger UI HTML
- `GET /swagger-ui` - Alternative route for Swagger UI
- `GET /swagger.yaml` - Serves YAML specification
- `GET /swagger.json` - Serves JSON specification

## Best Practices

1. **Keep Documentation Updated**: Update `swagger.yaml` when adding/modifying endpoints
2. **Use Examples**: Include example values in request/response schemas
3. **Descriptions**: Add clear descriptions for all endpoints and fields
4. **Tags**: Organize endpoints using tags (Meeting, Producer, Consumer, Transport)
5. **Validation**: Document required fields and validation rules

## Related Documentation

- [Handler Package](../pkg/handler/README.md) - API handler implementation
- [Main README](../README.md) - Project overview
- [OpenAPI Specification](https://swagger.io/specification/v2/) - OpenAPI 2.0 reference

## Troubleshooting

### Swagger UI Not Loading

1. **Check Server**: Ensure server is running
2. **Check Routes**: Verify routes are registered in `pkg/server/routes.go`
3. **Check Files**: Ensure `swagger-ui.html` and `swagger.yaml` exist in `docs/`
4. **Browser Console**: Check for JavaScript errors in browser console

### YAML Format Errors

- Validate YAML syntax:
  ```bash
  # Using yamllint
  yamllint swagger.yaml
  ```

- Use online validators:
  - https://editor.swagger.io/
  - https://www.swagger.io/tools/swagger-editor/

### Missing Endpoints

- Ensure all endpoints are documented in `swagger.yaml`
- Check that routes match the specification
- Verify handler methods exist
