package sfu

import "encoding/json"

type ProducerDetails struct {
	ProducerID    string `json:"producerId"`
	ParticipantID string `json:"participantId"`
	Kind          string `json:"kind"`
}

type CreateProducerResponse struct {
	ProducerDetails ProducerDetails `json:"producerDetails"`
}

type ResumeProducerResponse struct {
	ProducerDetails []ProducerDetails `json:"producerDetails"`
}

type PauseProducerResponse struct {
	ProducerDetails []ProducerDetails `json:"producerDetails"`
}

type CloseProducerResponse struct {
	ProducerDetails []ProducerDetails `json:"producerDetails"`
}

type CloseAllProducersResponse struct {
	ProducerDetails []ProducerDetails `json:"producerDetails"`
}

type CreateTransportResponse struct {
	TransportOptions json.RawMessage `json:"transportOptions"`
}

type RecreateProducerTransportResponse struct {
	TransportOptions      json.RawMessage   `json:"transportOptions"`
	ClosedProducerDetails []ProducerDetails `json:"closedProducerDetails"`
}

type RecreateBulkTransportResponse struct {
	ProducerTransportOptions json.RawMessage   `json:"producerTransportOptions"`
	ConsumerTransportOptions json.RawMessage   `json:"consumerTransportOptions"`
	ClosedProducerDetails    []ProducerDetails `json:"closedProducerDetails"`
}

type JoinMeetingResponse struct {
	ProducerRtpCapabilities json.RawMessage `json:"producerRtpCapabilities"`
	ConsumerRtpCapabilities json.RawMessage `json:"consumerRtpCapabilities"`
}

type LeaveMeetingResponse struct {
	Response   string `json:"response"`
	SendSignal bool   `json:"sendSignal"`
}
