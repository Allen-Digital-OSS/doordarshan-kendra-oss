package request

import "encoding/json"

// CreateMeetingRequestV1 represents the request to create a new meeting
type CreateMeetingRequestV1 struct {
	MeetingId               string `json:"meeting_id" validate:"required" example:"meeting-123"` // Unique identifier for the meeting
	MeetingCapacity         int    `json:"meeting_capacity" example:"100"`                       // Overall meeting capacity (total participants)
	ProducerMeetingCapacity int    `json:"producer_meeting_capacity" example:"50"`               // Maximum number of participants who can produce (send) media
	ConsumerMeetingCapacity int    `json:"consumer_meeting_capacity" example:"100"`              // Maximum number of participants who can consume (receive) media
	Tenant                  string `json:"tenant" validate:"required" example:"default"`         // Tenant identifier for multi-tenancy support
}

// JoinMeetingRequestV1 represents the request to join an existing meeting
type JoinMeetingRequestV1 struct {
	ParticipantId string `json:"participant_id" validate:"required" example:"participant-456"` // Unique identifier for the participant
	InstanceId    string `json:"instance_id" validate:"required" example:"instance-789"`       // Instance identifier for the participant's client
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier to join
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
}

// GetProducersOfMeetingRequestV1 represents the request to get all producers in a meeting
type GetProducersOfMeetingRequestV1 struct {
	MeetingId string `json:"meeting_id" validate:"required" example:"meeting-123"` // Meeting identifier
	Tenant    string `json:"tenant" validate:"required" example:"default"`         // Tenant identifier
}

// CreateTransportRequestV1 represents the request to create/recreate a transport
type CreateTransportRequestV1 struct {
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	ParticipantId string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	InstanceId    string `json:"instance_id" validate:"required" example:"instance-789"`       // Instance identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
}

// RestartIceRequestV1 represents the request to restart ICE connection
type RestartIceRequestV1 struct {
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	ParticipantId string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
}

// ConnectTransportRequestV1 represents the request to connect a transport with DTLS parameters
type ConnectTransportRequestV1 struct {
	ParticipantId  string          `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	DtlsParameters json.RawMessage `json:"dtls_parameters" validate:"required"`                          // DTLS parameters for WebRTC connection
	Tenant         string          `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId      string          `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
}

// CreateProducerRequestV1 represents the request to create a media producer
type CreateProducerRequestV1 struct {
	ParticipantId  string          `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	RtpParameters  json.RawMessage `json:"rtp_parameters" validate:"required"`                           // RTP parameters for the media stream
	Kind           string          `json:"kind" validate:"required" example:"audio"`                     // Media kind: "audio" or "video"
	Tenant         string          `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId      string          `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	StartRecording bool            `json:"start_recording" example:"false"`                              // Whether to start recording this producer
	InstanceId     string          `json:"instance_id" validate:"required" example:"instance-789"`       // Instance identifier
}

// CreateConsumerRequestV1 represents the request to create a media consumer
type CreateConsumerRequestV1 struct {
	ParticipantId      string              `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	RtpCapabilities    json.RawMessage     `json:"rtp_capabilities" validate:"required"`                         // RTP capabilities for consuming media
	TargetParticipants map[string][]string `json:"target_participants" validate:"required"`                      // Map of target participant IDs to their producer kinds to consume
	Tenant             string              `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId          string              `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	InstanceId         string              `json:"instance_id" validate:"required" example:"instance-789"`       // Instance identifier
}

// ResumeConsumerRequestV1 represents the request to resume a paused consumer
type ResumeConsumerRequestV1 struct {
	ParticipantId      string              `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	TargetParticipants map[string][]string `json:"target_participants" validate:"required"`                      // Map of target participant IDs to their consumer IDs to resume
	Tenant             string              `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId          string              `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
}

// ResumeProducerRequestV1 represents the request to resume a paused producer
type ResumeProducerRequestV1 struct {
	ParticipantId    string   `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	Kinds            []string `json:"kinds" validate:"required" example:"audio,video"`              // List of producer kinds to resume (e.g., ["audio", "video"])
	Tenant           string   `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId        string   `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	RecordingEnabled bool     `json:"recording_enabled" example:"false"`                            // Whether recording should be enabled
}

// PauseProducerRequestV1 represents the request to pause a producer
type PauseProducerRequestV1 struct {
	ParticipantId    string   `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	Kinds            []string `json:"kinds" validate:"required" example:"audio,video"`              // List of producer kinds to pause (e.g., ["audio", "video"])
	Tenant           string   `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId        string   `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	RecordingEnabled bool     `json:"recording_enabled" example:"false"`                            // Whether recording should be enabled
}

// PauseConsumerRequestV1 represents the request to pause a consumer
type PauseConsumerRequestV1 struct {
	ParticipantId      string              `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	TargetParticipants map[string][]string `json:"target_participants" validate:"required"`                      // Map of target participant IDs to their consumer IDs to pause
	Tenant             string              `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId          string              `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
}

// CloseConsumerRequestV1 represents the request to close a consumer
type CloseConsumerRequestV1 struct {
	ParticipantId      string              `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	TargetParticipants map[string][]string `json:"target_participants" validate:"required"`                      // Map of target participant IDs to their consumer IDs to close
	Tenant             string              `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId          string              `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
}

// CloseProducerRequestV1 represents the request to close a producer
type CloseProducerRequestV1 struct {
	ParticipantId string   `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	Kinds         []string `json:"kinds" validate:"required" example:"audio,video"`              // List of producer kinds to close (e.g., ["audio", "video"])
	Tenant        string   `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	MeetingId     string   `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
}

type CloseAllProducersRequest struct {
	ParticipantId string `json:"participant_id" validate:"required"`
	Tenant        string `json:"tenant" validate:"required"`
	MeetingId     string `json:"meeting_id" validate:"required"`
}

type CloseAllConsumersForProducerRequest struct {
	ParticipantId       string `json:"participant_id" validate:"required"`
	TargetParticipantId string `json:"target_participant_id" validate:"required"`
	Tenant              string `json:"tenant" validate:"required"`
	MeetingId           string `json:"meeting_id" validate:"required"`
}

// LeaveMeetingRequest represents the request to leave a meeting
type LeaveMeetingRequest struct {
	ParticipantId string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	InstanceId    string `json:"instance_id" validate:"required" example:"instance-789"`       // Instance identifier
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
}

// EndMeetingRequest represents the request to end a meeting
type EndMeetingRequest struct {
	MeetingId string `json:"meeting_id" validate:"required" example:"meeting-123"` // Meeting identifier
	Tenant    string `json:"tenant" validate:"required" example:"default"`         // Tenant identifier
}

// GetRTPCapabilitiesRequest represents the request to get RTP capabilities
type GetRTPCapabilitiesRequest struct {
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	ParticipantId string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
}

// PreMeetingDetailsRequest represents the request to get pre-meeting details
type PreMeetingDetailsRequest struct {
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	ParticipantID string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
}

// GetActiveContainerOfMeetingRequest represents the request to get active container information
type GetActiveContainerOfMeetingRequest struct {
	MeetingId     string `json:"meeting_id" validate:"required" example:"meeting-123"`         // Meeting identifier
	Tenant        string `json:"tenant" validate:"required" example:"default"`                 // Tenant identifier
	ParticipantID string `json:"participant_id" validate:"required" example:"participant-456"` // Participant identifier
}
