# Signaling Platform Package

This package defines the data structures and models for communicating with the Signaling Platform service. **Note: The Signaling Platform itself is NOT open source** - this package only defines the message format.

## Overview

The Signaling Platform is a separate, proprietary service that:
- Consumes messages from Redis Streams
- Manages WebSocket connections with meeting participants
- Broadcasts signaling messages to connected clients

This package provides the request/response structures used to communicate with the Signaling Platform.

## Important Note

⚠️ **The Signaling Platform service is NOT part of this open-source repository.**

This service only **publishes** messages to Redis Streams. You must implement or integrate your own Signaling Platform that:
1. Consumes from Redis Streams (format: `room-stream:{meeting_id}`)
2. Manages WebSocket connections
3. Broadcasts messages to connected clients

## Message Flow

```
┌─────────────────────┐
│  DoorDarshan Kendra │
│   (This Service)    │
└──────────┬──────────┘
           │
           │ Publishes to Redis Streams
           ▼
┌─────────────────────┐
│   Redis Streams     │
│ room-stream:{id}    │
└──────────┬──────────┘
           │
           │ Consumed by
           ▼
┌─────────────────────┐
│ Signaling Platform  │  ← NOT OPEN SOURCE
│  (Your Implementation)│
└──────────┬──────────┘
           │
           │ WebSocket
           ▼
┌─────────────────────┐
│   Client Apps       │
└─────────────────────┘
```

## Data Structures

### BroadcastSignalRequest

Main structure for broadcasting signals to meeting participants:

```go
type BroadcastSignalRequest struct {
    Type        string        `json:"type"`         // Signal type
    SubType     string        `json:"sub_type"`     // Signal subtype
    ClientID    string        `json:"client_id"`   // Client identifier
    RequestID   string        `json:"request_id"`  // Unique request ID
    Details     SignalDetails `json:"details"`      // Signal payload
    SenderID    string        `json:"sender_id"`   // Sender participant ID
    ReceiverIDs []string      `json:"receiver_ids"` // Target participants (nil = all)
    RoomID      string        `json:"room_id"`     // Meeting/room ID
    VersionInfo VersionInfo   `json:"version_info"` // Version control
}
```

### SignalDetails

Contains the actual signal payload:

```go
type SignalDetails struct {
    Payload   json.RawMessage `json:"payload"`   // Signal-specific payload
    Timestamp int64           `json:"timestamp"` // Unix timestamp (nanoseconds)
}
```

### VersionInfo

Version control information:

```go
type VersionInfo struct {
    SkipVersion bool  `json:"skip_version"` // Skip version checking
    Version     int64 `json:"version"`     // Version number
}
```

### BulkBroadcastSignalRequest

For broadcasting multiple signals at once:

```go
type BulkBroadcastSignalRequest struct {
    BroadcastSignalRequest []BroadcastSignalRequest `json:"broadcast_signal_request"`
}
```

## Signal Types

The following signal types are supported:

### 1. DoordarshanStreamPublish

Published when a participant starts publishing a media stream.

**Payload Structure:**
```json
{
  "publishReason": "user_started_streaming",
  "producerMap": {
    "audio": "producer-id-1",
    "video": "producer-id-2"
  },
  "participantId": "participant-456"
}
```

### 2. DoordarshanStreamUnPublish

Published when a participant stops publishing a media stream.

**Payload Structure:**
```json
{
  "unPublishReason": "user_stopped_streaming",
  "producerMap": {
    "audio": "producer-id-1",
    "video": "producer-id-2"
  },
  "participantId": "participant-456"
}
```

### 3. DoordarshanUserJoined

Published when a participant joins a meeting.

**Payload Structure:**
```json
{
  "userJoinedReason": "participant_joined_meeting"
}
```

### 4. DoordarshanUserLeft

Published when a participant leaves a meeting.

**Payload Structure:**
```json
{
  "userLeftReason": "participant_left_meeting",
  "instanceId": "instance-789"
}
```

### 5. DoordarshanDisconnect

Published when a participant disconnects unexpectedly.

**Payload Structure:**
```json
{
  "disconnectReason": "connection_lost"
}
```

## Redis Stream Format

Messages are published to Redis Streams with the key: `room-stream:{meeting_id}`

### Stream Message Structure

```go
map[string]interface{}{
    "name":          "DoordarshanStreamPublish",  // Signal type
    "signal":        signalJSONString,             // JSON string of BroadcastSignalRequest
    "roomId":        "meeting-123",                // Meeting ID
    "participantId": "participant-456",           // Participant ID
    "messageType":   "signalMessage",              // Always "signalMessage"
    "requestId":     "unique-request-id",         // Unique request ID
}
```

### Example Stream Message

```json
{
  "name": "DoordarshanStreamPublish",
  "signal": "{\"type\":\"DoordarshanStreamPublish\",\"sub_type\":\"DoordarshanStreamPublish\",\"client_id\":\"DoordarshanKendra\",\"request_id\":\"participant-4561234567890\",\"details\":{\"payload\":{\"publishReason\":\"user_started_streaming\",\"producerMap\":{\"audio\":\"producer-1\"},\"participantId\":\"participant-456\"},\"timestamp\":1234567890000000000},\"sender_id\":\"participant-456\",\"receiver_ids\":null,\"room_id\":\"meeting-123\",\"version_info\":{\"skip_version\":true,\"version\":0}}",
  "roomId": "meeting-123",
  "participantId": "participant-456",
  "messageType": "signalMessage",
  "requestId": "participant-4561234567890"
}
```

## Implementation Guide

### For Signaling Platform Implementers

If you're implementing your own Signaling Platform, you need to:

1. **Consume Redis Streams**
   - Subscribe to streams matching pattern: `room-stream:*`
   - Or consume specific streams: `room-stream:{meeting_id}`

2. **Parse Messages**
   - Extract `signal` field (JSON string)
   - Parse `BroadcastSignalRequest` from JSON
   - Extract `Details.Payload` for signal-specific data

3. **Manage WebSocket Connections**
   - Maintain WebSocket connections per participant
   - Map participant IDs to WebSocket connections
   - Handle connection lifecycle (connect, disconnect, reconnect)

4. **Broadcast Messages**
   - If `ReceiverIDs` is nil/empty: broadcast to all participants in room
   - If `ReceiverIDs` is set: send only to specified participants
   - Format message appropriately for client consumption

### Example Consumer (Pseudocode)

```go
// Consume from Redis Stream
streamKey := fmt.Sprintf("room-stream:%s", meetingID)
messages, err := redisClient.XRead(ctx, &redis.XReadArgs{
    Streams: []string{streamKey, "0"},
    Count:   10,
    Block:   time.Second,
}).Result()

for _, message := range messages[0].Messages {
    // Extract signal
    signalJSON := message.Values["signal"].(string)

    // Parse BroadcastSignalRequest
    var request BroadcastSignalRequest
    json.Unmarshal([]byte(signalJSON), &request)

    // Get WebSocket connections for room
    connections := getWebSocketConnections(request.RoomID)

    // Broadcast to connections
    if request.ReceiverIDs == nil {
        // Broadcast to all
        broadcastToAll(connections, request)
    } else {
        // Broadcast to specific participants
        broadcastToParticipants(connections, request.ReceiverIDs, request)
    }
}
```

## HTTP API (Optional)

The Signaling Platform may also expose an HTTP API for direct broadcasting (used by this service):

### Endpoints

- `POST /room/{room_id}/broadcast` - Broadcast to a specific room
- `POST /room/broadcast/bulk` - Bulk broadcast to multiple rooms

See `pkg/clients/signaling_platform.go` for the HTTP client implementation.

## Client ID

The service uses `ClientID = "DoordarshanKendra"` when publishing signals. Your Signaling Platform should recognize this client ID.

## Request ID Format

Request IDs are generated as: `{participantId}{timestamp}`

Example: `participant-4561234567890`

## Timestamp Format

Timestamps are Unix timestamps in nanoseconds (int64).

Example: `1234567890000000000` (nanoseconds since epoch)

## Validation

Signal types are validated using the `oneof` validator:

```go
Type string `json:"type" validate:"required,oneof=DoordarshanStreamPublish DoordarshanStreamUnPublish DoordarshanUserJoined DoordarshanUserLeft DoordarshanDisconnect"`
```

Only the specified signal types are valid.

## Related Documentation

- [Handler Package](../handler/README.md) - Where signals are published
- [Data Layer](../data/README.md) - Redis Stream operations
- [Main README](../../README.md) - Overall architecture

## Troubleshooting

### Messages Not Being Consumed

1. **Check Redis Stream**: Verify messages are being published
   ```bash
   redis-cli XREAD STREAMS room-stream:meeting-123 0
   ```

2. **Verify Stream Key**: Ensure format is `room-stream:{meeting_id}`

3. **Check Message Format**: Verify all required fields are present

### Signal Type Validation Errors

- Ensure signal type matches one of the allowed values
- Check that `Type` and `SubType` are set correctly
- Verify payload structure matches signal type

### WebSocket Broadcasting Issues

- Verify WebSocket connections are maintained
- Check participant ID mapping
- Ensure proper message formatting for clients
