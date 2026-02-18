package sfu

import (
	"encoding/json"
)

type PayloadGenerator struct {
}

func (p PayloadGenerator) GetCreateMeetingPayload(meetingId string, freeContainerResponse FreeContainersResponse) CreateMeetingRequest {
	if freeContainerResponse.WorkerCapacity != nil {
		return CreateMeetingRequest{
			Routers: &RouterDetails{
				RouterCapacity: freeContainerResponse.WorkerCapacity.Capacity,
				WorkerId:       freeContainerResponse.WorkerCapacity.WorkerId,
			},
			MeetingID:   meetingId,
			ContainerID: freeContainerResponse.ContainerId,
		}
	} else {
		createMeetingRequest := CreateMeetingRequest{}
		createMeetingRequest.MeetingID = meetingId
		createMeetingRequest.ContainerID = freeContainerResponse.ContainerId
		for _, consumerWorkerCapacity := range freeContainerResponse.ConsumerWorkerCapacity {
			createMeetingRequest.ConsumerRouterDetails = append(createMeetingRequest.ConsumerRouterDetails, RouterDetails{
				WorkerId:       consumerWorkerCapacity.WorkerId,
				RouterCapacity: consumerWorkerCapacity.Capacity,
			})
		}
		producerRouterDetails := RouterDetails{
			RouterCapacity: freeContainerResponse.ProducerWorkerCapacity.Capacity,
			WorkerId:       freeContainerResponse.ProducerWorkerCapacity.WorkerId,
		}
		createMeetingRequest.ProducerRouterDetails = []RouterDetails{producerRouterDetails}
		return createMeetingRequest
	}
}

type ReplaceMeetingPayloadInput struct {
	WorkerId               string
	RouterCapacity         int
	ProducerRouterCapacity int
	ConsumerRouterCapacity int
}

func (p PayloadGenerator) GetReplaceMeetingPayload(meetingId string, oldContainerId string, newContainerId string, inputs []ReplaceMeetingPayloadInput) ReplaceMeetingRequest {
	var routerDetails *RouterDetails
	producerRouterDetails := make([]RouterDetails, 0)
	consumerRouterDetails := make([]RouterDetails, 0)
	for _, input := range inputs {
		if input.RouterCapacity != 0 {
			routerDetails = &RouterDetails{}
			routerDetails.RouterCapacity = input.RouterCapacity
			routerDetails.WorkerId = input.WorkerId
		} else if input.ProducerRouterCapacity != 0 {
			tempRouterDetails := RouterDetails{
				RouterCapacity: input.ProducerRouterCapacity,
				WorkerId:       input.WorkerId,
			}
			producerRouterDetails = append(producerRouterDetails, tempRouterDetails)
		} else if input.ConsumerRouterCapacity != 0 {
			tempRouterDetails := RouterDetails{
				RouterCapacity: input.ConsumerRouterCapacity,
				WorkerId:       input.WorkerId,
			}
			consumerRouterDetails = append(consumerRouterDetails, tempRouterDetails)
		}
	}
	return ReplaceMeetingRequest{
		Routers:               routerDetails,
		ProducerRouterDetails: producerRouterDetails,
		ConsumerRouterDetails: consumerRouterDetails,
		MeetingID:             meetingId,
		OldContainerId:        oldContainerId,
		NewContainerId:        newContainerId,
	}
}

func (p PayloadGenerator) GetJoinMeetingPayload(participantId string, instanceId string, producerRouterId string,
	consumerRouterId string, meetingId string) JoinMeetingRequest {
	return JoinMeetingRequest{
		ProducerRouterId: producerRouterId,
		ConsumerRouterId: consumerRouterId,
		ParticipantId:    participantId,
		MeetingId:        meetingId,
		InstanceId:       instanceId,
	}
}

func (p PayloadGenerator) GetGetProducersOfMeetingPayload(meetingId string) GetProducersOfMeetingRequest {
	return GetProducersOfMeetingRequest{
		MeetingID: meetingId,
	}
}

type GetCreateTransportRequestInput struct {
	ParticipantId          string
	InstanceId             string
	RouterId               string
	MeetingId              string
	OldProducerTransportId string
	OldConsumerTransportId string
	OldProducerIds         []string
	OldConsumerIds         []string
}

func (p PayloadGenerator) GetCreateTransportRequest(input GetCreateTransportRequestInput) CreateTransportRequest {
	return CreateTransportRequest{
		MeetingID:              input.MeetingId,
		ParticipantId:          input.ParticipantId,
		InstanceId:             input.InstanceId,
		RouterId:               input.RouterId,
		OldProducerTransportId: input.OldProducerTransportId,
		OldConsumerTransportId: input.OldConsumerTransportId,
		OldProducerIds:         input.OldProducerIds,
		OldConsumerIds:         input.OldConsumerIds,
	}
}

func (p PayloadGenerator) GetRestartIceRequest(participantId string, transportIds []string) RestartIceRequest {
	return RestartIceRequest{
		ParticipantId: participantId,
		TransportIds:  transportIds,
	}
}

type CreateProducerInput struct {
	ParticipantId       string
	Kind                string
	RtpParameters       json.RawMessage
	MeetingID           string
	ProducerRouterId    string
	ProducerTransportId string
	ConsumerRouterIds   []string
	StartRecording      bool
	InstanceId          string
}

func (p PayloadGenerator) CreateProduceRequest(input CreateProducerInput) CreateProduceRequest {
	return CreateProduceRequest{
		ParticipantId:       input.ParticipantId,
		RtpParameters:       input.RtpParameters,
		Kind:                input.Kind,
		MeetingID:           input.MeetingID,
		ProducerRouterId:    input.ProducerRouterId,
		ProducerTransportId: input.ProducerTransportId,
		ConsumerRouterIds:   input.ConsumerRouterIds,
		StartRecording:      input.StartRecording,
		InstanceId:          input.InstanceId,
	}
}

func (p PayloadGenerator) ConnectTransportRequest(participantId string, dtlsParameters json.RawMessage, transportId string, meetingId string) ConnectTransportRequest {
	return ConnectTransportRequest{
		ParticipantId:  participantId,
		DtlsParameters: dtlsParameters,
		MeetingID:      meetingId,
		TransportId:    transportId,
	}
}

func (p PayloadGenerator) CreateConsumeRequest(participantId string, rtpCapabilities json.RawMessage,
	consumerTransportId string, targetParticipants map[string]map[string]string, meetingId string, instanceId string) CreateConsumeRequest {
	return CreateConsumeRequest{
		ParticipantId:      participantId,
		RtpCapabilities:    rtpCapabilities,
		TransportId:        consumerTransportId,
		TargetParticipants: targetParticipants,
		MeetingID:          meetingId,
		InstanceId:         instanceId,
	}
}

func (p PayloadGenerator) ResumeConsumeRequest(participantId string, consumerIds []string, meetingId string) ResumeConsumeRequest {
	return ResumeConsumeRequest{
		ParticipantId: participantId,
		ConsumerIds:   consumerIds,
		MeetingID:     meetingId,
	}
}

func (p PayloadGenerator) ResumeProduceRequest(participantId string, producerIdKindMap map[string]string, meetingId string, recordingEnabled bool) ResumeProduceRequest {
	return ResumeProduceRequest{
		ParticipantId:    participantId,
		ProducerKindMap:  producerIdKindMap,
		MeetingID:        meetingId,
		RecordingEnabled: recordingEnabled,
	}
}

func (p PayloadGenerator) PauseProducerRequest(participantId string, producerIdKindMap map[string]string, meetingId string, recordingEnabled bool) PauseProducerRequest {
	return PauseProducerRequest{
		ParticipantId:    participantId,
		ProducerKindMap:  producerIdKindMap,
		MeetingID:        meetingId,
		RecordingEnabled: recordingEnabled,
	}
}

func (p PayloadGenerator) PauseConsumerRequest(participantId string, consumerIds []string, meetingId string) PauseConsumerRequest {
	return PauseConsumerRequest{
		ParticipantId: participantId,
		ConsumerIds:   consumerIds,
		MeetingID:     meetingId,
	}
}

func (p PayloadGenerator) CloseConsumerRequest(participantId string, consumerIds []string, meetingId string) CloseConsumerRequest {
	return CloseConsumerRequest{
		ParticipantId: participantId,
		ConsumerIds:   consumerIds,
		MeetingID:     meetingId,
	}
}

func (p PayloadGenerator) CloseProducerRequest(participantId string, producerIdKindMap map[string]string, meetingId string) CloseProducerRequest {
	return CloseProducerRequest{
		ParticipantId:   participantId,
		ProducerKindMap: producerIdKindMap,
		MeetingID:       meetingId,
	}
}

func (p PayloadGenerator) CloseAllProducersRequest(participantId string, meetingId string) CloseAllProducersRequest {
	return CloseAllProducersRequest{
		ParticipantId: participantId,
		MeetingID:     meetingId,
	}
}

func (p PayloadGenerator) CloseAllConsumersForProducerRequest(participantId string, targetParticipantId string, meetingId string) CloseAllConsumersForProducerRequest {
	return CloseAllConsumersForProducerRequest{
		ParticipantId:       participantId,
		TargetParticipantId: targetParticipantId,
		MeetingID:           meetingId,
	}
}

func (p PayloadGenerator) LeaveMeetingRequest(participantId string, instanceId string, meetingId string, producerTransportId string,
	consumerTransportId string, producerIds []string, consumerIds []string) LeaveMeetingRequest {
	return LeaveMeetingRequest{
		ParticipantId:       participantId,
		InstanceId:          instanceId,
		MeetingId:           meetingId,
		ProducerTransportId: producerTransportId,
		ConsumerTransportId: consumerTransportId,
		ProducerIds:         producerIds,
		ConsumerIds:         consumerIds,
	}
}
func (p PayloadGenerator) EndMeetingRequest(meetingId string, routerIds []string,
	transportIds []string, producerIds []string, consumerIds []string, participantIds []string) EndMeetingRequest {
	return EndMeetingRequest{
		MeetingId:      meetingId,
		RouterIds:      routerIds,
		TransportIds:   transportIds,
		ProducerIds:    producerIds,
		ConsumerIds:    consumerIds,
		ParticipantIds: participantIds,
	}
}

func (p PayloadGenerator) GetRTPCapabilitiesRequest(meetingId, producerRouterId, consumerRouterId string) GetRTPCapabilitiesRequest {
	return GetRTPCapabilitiesRequest{
		MeetingId:        meetingId,
		ProducerRouterId: producerRouterId,
		ConsumerRouterId: consumerRouterId,
	}
}

func (p PayloadGenerator) PreMeetingDetailsRequest(meetingId, producerRouterId, consumerRouterId, participantID string) PreClassDetailsRequest {
	return PreClassDetailsRequest{
		MeetingId:        meetingId,
		ProducerRouterId: producerRouterId,
		ConsumerRouterId: consumerRouterId,
		ParticipantId:    participantID,
	}
}
