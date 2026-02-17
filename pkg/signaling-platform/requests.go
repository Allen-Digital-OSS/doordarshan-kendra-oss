package signaling_platform

import "encoding/json"

type BroadcastSignalRequest struct {
	Type        string        `json:"type" validate:"required,oneof=DoordarshanStreamPublish DoordarshanStreamUnPublish DoordarshanUserJoined DoordarshanUserLeft DoordarshanDisconnect"`
	SubType     string        `json:"sub_type"`
	ClientID    string        `json:"client_id" validate:"required"`
	RequestID   string        `json:"request_id"  validate:"required"`
	Details     SignalDetails `json:"details" validate:"required"`
	SenderID    string        `json:"sender_id" validate:"required"`
	ReceiverIDs []string      `json:"receiver_ids"`
	RoomID      string        `json:"room_id"`
	VersionInfo VersionInfo   `json:"version_info"`
}

type VersionInfo struct {
	SkipVersion bool  `json:"skip_version"`
	Version     int64 `json:"version"`
}

type BulkBroadcastSignalRequest struct {
	BroadcastSignalRequest []BroadcastSignalRequest `json:"broadcast_signal_request" validate:"required,dive"`
}

type SignalDetails struct {
	Payload   json.RawMessage `json:"payload" validate:"required"`
	Timestamp int64           `json:"timestamp"`
}
