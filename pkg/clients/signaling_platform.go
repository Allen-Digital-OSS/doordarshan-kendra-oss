package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	signaling_platform "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/signaling-platform"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
	"github.com/google/wire"
	"github.com/labstack/gommon/log"
	"net/http"
)

const bulkBroadcastUrl string = "/room/broadcast/bulk"
const broadcastUrl string = "/room/%s/broadcast"
const POST string = "POST"

var ProviderSet = wire.NewSet(NewSignalingPlatformClient)

type SignalingPlatformClientInterface interface {
	BulkBroadcastHandler(request signaling_platform.BulkBroadcastSignalRequest) (*http.Response, error)
	BroadcastHandler(request signaling_platform.BroadcastSignalRequest) (*http.Response, error)
}

type SignalingPlatformClient struct {
	client *http.Client
	config common.SignalingPlatformConfig
}

func NewSignalingPlatformClient(appConfig *common.AppConfig) *SignalingPlatformClient {
	/*retryPolicy := failsafehttp.RetryPolicyBuilder().
		WithDelay(time.Second).
		WithMaxRetries(3).
		Build()
	circuitBreaker := circuitbreaker.Builder[*http.Response]().
		WithFailureThreshold(5).
		WithDelay(time.Minute).
		WithSuccessThreshold(2).Build()
	timeoutPolicy := timeout.With[*http.Response](time.Duration(config.Timeout) * time.Millisecond)
	roundTripper := failsafehttp.NewRoundTripper(nil, retryPolicy, circuitBreaker, timeoutPolicy)*/
	roundTripper := failsafehttp.NewRoundTripper(nil)
	client := &http.Client{
		Transport: roundTripper,
	}
	return &SignalingPlatformClient{
		client: client,
		config: appConfig.SignalingPlatformConfig,
	}
}

func (s *SignalingPlatformClient) BulkBroadcastHandler(
	request signaling_platform.BulkBroadcastSignalRequest) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Error("Error while marshalling request: %v", err)
		return nil, err
	}
	httpRequest, err := http.NewRequest(POST,
		fmt.Sprintf("%s%s", s.config.Endpoint, bulkBroadcastUrl),
		bytes.NewBuffer(requestBytes))
	if err != nil {
		log.Error("Error while creating request: %v", err)
		return nil, err
	}
	log.Infof("Request is %v", httpRequest)
	failsafeRequest := failsafehttp.NewRequest(httpRequest, s.client)
	httpRequest.Header.Add("Content-Type", "application/json")
	response, err := failsafeRequest.Do()
	if err != nil {
		log.Error("Error while sending request: %v", err)
		return nil, err
	}
	return response, nil
}

func (s *SignalingPlatformClient) BroadcastHandler(
	request signaling_platform.BroadcastSignalRequest) (*http.Response, error) {
	requestBytes, err := json.Marshal(request)
	if err != nil {
		log.Error("Error while marshalling request: %v", err)
		return nil, err
	}
	roomBroadcastUrl := fmt.Sprintf(broadcastUrl, request.RoomID)
	httpRequest, err := http.NewRequest(POST,
		fmt.Sprintf("%s%s", s.config.Endpoint, roomBroadcastUrl),
		bytes.NewBuffer(requestBytes))
	if err != nil {
		log.Error("Error while creating request: %v", err)
		return nil, err
	}
	log.Infof("Request is %v", httpRequest)
	failsafeRequest := failsafehttp.NewRequest(httpRequest, s.client)
	httpRequest.Header.Add("Content-Type", "application/json")
	response, err := failsafeRequest.Do()
	if err != nil {
		log.Error("Error while sending request: %v", err)
		return nil, err
	}
	return response, nil
}
