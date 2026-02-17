package sfu

import (
	"encoding/json"
	"github.com/labstack/gommon/log"
)

func (p PayloadGenerator) ProducerDetails(body []byte) (*ProducerDetails, error) {
	var producerDetails *ProducerDetails
	err := json.Unmarshal(body, &producerDetails)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return producerDetails, nil
}

func (p PayloadGenerator) PauseProducerResponse(body []byte) (*PauseProducerResponse, error) {
	var pauseProducerResponse *PauseProducerResponse
	err := json.Unmarshal(body, &pauseProducerResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return pauseProducerResponse, nil
}

func (p PayloadGenerator) ResumeProducerResponse(body []byte) (*ResumeProducerResponse, error) {
	var resumeProducerResponse *ResumeProducerResponse
	err := json.Unmarshal(body, &resumeProducerResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return resumeProducerResponse, nil
}

func (p PayloadGenerator) CloseProducerResponse(body []byte) (*CloseProducerResponse, error) {
	var closeProducerResponse *CloseProducerResponse
	err := json.Unmarshal(body, &closeProducerResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return closeProducerResponse, nil
}

func (p PayloadGenerator) CloseAllProducersResponse(body []byte) (*CloseAllProducersResponse, error) {
	var closeAllProducersResponse *CloseAllProducersResponse
	err := json.Unmarshal(body, &closeAllProducersResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return closeAllProducersResponse, nil
}

func (p PayloadGenerator) RecreateProducerTransportResponse(body []byte) (*RecreateProducerTransportResponse, error) {
	var recreateProducerTransportResponse *RecreateProducerTransportResponse
	err := json.Unmarshal(body, &recreateProducerTransportResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return recreateProducerTransportResponse, nil
}

func (p PayloadGenerator) RecreateBulkTransportResponse(body []byte) (*RecreateBulkTransportResponse, error) {
	var recreateBulkTransportResponse *RecreateBulkTransportResponse
	err := json.Unmarshal(body, &recreateBulkTransportResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return recreateBulkTransportResponse, nil
}

func (p PayloadGenerator) LeaveMeetingResponse(body []byte) (*LeaveMeetingResponse, error) {
	var leaveMeetingResponse *LeaveMeetingResponse
	err := json.Unmarshal(body, &leaveMeetingResponse)
	if err != nil {
		log.Error("Error while unmarshalling response: ", err)
		return nil, err
	}
	return leaveMeetingResponse, nil
}
