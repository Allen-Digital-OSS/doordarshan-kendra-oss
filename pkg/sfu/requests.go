package sfu

import (
	"encoding/json"
)

type RouterDetails struct {
	RouterCapacity int    `json:"routerCapacity"`
	WorkerId       string `json:"workerId"`
}

type ReplaceMeetingsRequest struct {
	ReplaceMeetingRequests []ReplaceMeetingRequest `json:"replaceMeetingRequests"`
}

type ReplaceMeetingRequest struct {
	ProducerRouterDetails []RouterDetails `json:"producerRouterDetails,omitempty"`
	ConsumerRouterDetails []RouterDetails `json:"consumerRouterDetails,omitempty"`
	Routers               *RouterDetails  `json:"routers,omitempty"`
	MeetingID             string          `json:"meetingId"`
	OldContainerId        string          `json:"oldContainerId"`
	NewContainerId        string          `json:"newContainerId"`
}

type CreateMeetingRequest struct {
	ProducerRouterDetails []RouterDetails `json:"producerRouterDetails,omitempty"`
	ConsumerRouterDetails []RouterDetails `json:"consumerRouterDetails,omitempty"`
	Routers               *RouterDetails  `json:"routers,omitempty"`
	MeetingID             string          `json:"meetingId"`
	ContainerID           string          `json:"containerId"`
}

type JoinMeetingRequest struct {
	ProducerRouterId string `json:"producerRouterId"`
	ConsumerRouterId string `json:"consumerRouterId"`
	ParticipantId    string `json:"participantId"`
	MeetingId        string `json:"meetingId"`
	InstanceId       string `json:"instanceId"`
}

type GetProducersOfMeetingRequest struct {
	MeetingID string `json:"meetingId"`
}

type CreateTransportRequest struct {
	MeetingID              string   `json:"meetingId"`
	InstanceId             string   `json:"instanceId"`
	ParticipantId          string   `json:"participantId"`
	RouterId               string   `json:"routerId"`
	OldProducerTransportId string   `json:"oldProducerTransportId"`
	OldConsumerTransportId string   `json:"oldConsumerTransportId"`
	OldProducerIds         []string `json:"oldProducerIds"`
	OldConsumerIds         []string `json:"oldConsumerIds"`
}

type RestartIceRequest struct {
	ParticipantId string   `json:"participantId"`
	TransportIds  []string `json:"transportIds"`
}

type CreateProduceRequest struct {
	ParticipantId       string          `json:"participantId"`
	RtpParameters       json.RawMessage `json:"rtpParameters"`
	Kind                string          `json:"kind"`
	MeetingID           string          `json:"meetingId"`
	ProducerRouterId    string          `json:"producerRouterId"`
	ProducerTransportId string          `json:"producerTransportId"`
	ConsumerRouterIds   []string        `json:"consumerRouterIds"`
	StartRecording      bool            `json:"startRecording"`
	InstanceId          string          `json:"instanceId"`
}

type ConnectTransportRequest struct {
	ParticipantId  string          `json:"participantId"`
	DtlsParameters json.RawMessage `json:"dtlsParameters"`
	MeetingID      string          `json:"meetingId"`
	TransportId    string          `json:"transportId"`
}

type CreateConsumeRequest struct {
	ParticipantId      string                       `json:"participantId"`
	RtpCapabilities    json.RawMessage              `json:"rtpCapabilities"`
	MeetingID          string                       `json:"meetingId"`
	TransportId        string                       `json:"transportId"`
	TargetParticipants map[string]map[string]string `json:"targetParticipants"`
	InstanceId         string                       `json:"instanceId"`
}

type ResumeConsumeRequest struct {
	ParticipantId string   `json:"participantId"`
	MeetingID     string   `json:"meetingId"`
	ConsumerIds   []string `json:"consumerIds"`
}

type ResumeProduceRequest struct {
	ParticipantId    string            `json:"participantId"`
	MeetingID        string            `json:"meetingId"`
	ProducerKindMap  map[string]string `json:"producerKindMap"`
	RecordingEnabled bool              `json:"recordingEnabled"`
}

type PauseProducerRequest struct {
	ParticipantId    string            `json:"participantId"`
	MeetingID        string            `json:"meetingId"`
	ProducerKindMap  map[string]string `json:"producerKindMap"`
	RecordingEnabled bool              `json:"recordingEnabled"`
}

type PauseConsumerRequest struct {
	ParticipantId string   `json:"participantId"`
	MeetingID     string   `json:"meetingId"`
	ConsumerIds   []string `json:"consumerIds"`
}

type CloseConsumerRequest struct {
	ParticipantId string   `json:"participantId"`
	MeetingID     string   `json:"meetingId"`
	ConsumerIds   []string `json:"consumerIds"`
}

type CloseProducerRequest struct {
	ParticipantId   string            `json:"participantId"`
	MeetingID       string            `json:"meetingId"`
	ProducerKindMap map[string]string `json:"producerKindMap"`
}

type CloseAllProducersRequest struct {
	ParticipantId string `json:"participantId"`
	MeetingID     string `json:"meetingId"`
}

type CloseAllConsumersForProducerRequest struct {
	ParticipantId       string `json:"participantId"`
	TargetParticipantId string `json:"targetParticipantId"`
	MeetingID           string `json:"meetingId"`
}

type LeaveMeetingRequest struct {
	ParticipantId       string   `json:"participantId"`
	InstanceId          string   `json:"instanceId"`
	MeetingId           string   `json:"meetingId"`
	ProducerTransportId string   `json:"producerTransportId"`
	ConsumerTransportId string   `json:"consumerTransportId"`
	ProducerIds         []string `json:"producerIds"`
	ConsumerIds         []string `json:"consumerIds"`
}

type EndMeetingRequest struct {
	MeetingId      string   `json:"meetingId"`
	RouterIds      []string `json:"routerIds"`
	TransportIds   []string `json:"transportIds"`
	ProducerIds    []string `json:"producerIds"`
	ConsumerIds    []string `json:"consumerIds"`
	ParticipantIds []string `json:"participantIds"`
}

type GetRTPCapabilitiesRequest struct {
	MeetingId        string `json:"meetingId"`
	ProducerRouterId string `json:"producerRouterId"`
	ConsumerRouterId string `json:"consumerRouterId"`
}

type PreClassDetailsRequest struct {
	MeetingId        string `json:"meetingId"`
	ProducerRouterId string `json:"producerRouterId"`
	ConsumerRouterId string `json:"consumerRouterId"`
	ParticipantId    string `json:"participantId"`
}
