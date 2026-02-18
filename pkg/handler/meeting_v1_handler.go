package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/handler/request"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/handler/response"
	appLog "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/log"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/myerrors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/sfu"
	signaling_platform "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/signaling-platform"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"strconv"
	"time"
)

const createMeetingUrl string = "/v1/createMeeting"
const replaceMeetingsUrl string = "/v1/replaceMeetings"
const joinMeetingUrl string = "/v1/joinMeeting"
const getProducersOfMeetingUrl string = "/v1/getProducersOfMeeting"
const recreateProducerTransportUrl string = "/v1/recreateProducerTransport"
const connectTransportUrl string = "/v1/connectTransport"
const recreateConsumerTransportUrl string = "/v1/recreateConsumerTransport"
const recreateTransportsUrl string = "/v1/recreateTransports"
const restartProducerIceUrl string = "/v1/restartProducerIce"
const restartConsumerIceUrl string = "/v1/restartConsumerIce"
const restartIceUrl string = "/v1/restartIce"
const createConsumerUrl string = "/v1/createConsumer"
const createProducerUrl string = "/v1/createProducer"
const resumeConsumerUrl string = "/v1/resumeConsumer"
const resumeProducerUrl string = "/v1/resumeProducer"
const pauseProducerUrl string = "/v1/pauseProducer"
const pauseConsumerUrl string = "/v1/pauseConsumer"
const closeProducerUrl string = "/v1/closeProducer"
const closeAllProducersUrl string = "/v1/closeAllProducers"
const closeAllConsumersForProducerUrl string = "/v1/closeAllConsumersForProducer"
const closeConsumerUrl string = "/v1/closeConsumer"
const leaveMeetingUrl string = "/v1/leaveMeeting"
const endMeetingUrl string = "/v1/endMeeting"
const getRTPCapabilitiesUrl string = "/v1/getRTPCapabilities"
const DDKTenantForSignalingPlatform = "DoordarshanKendra"
const DoordarshanDisconnectSignal = "DoordarshanDisconnect"
const DoordarshanStreamPublishSignal = "DoordarshanStreamPublish"
const DoordarshanStreamUnPublishSignal = "DoordarshanStreamUnPublish"
const DoordarshanUserJoinedSignal = "DoordarshanUserJoined"
const DoordarshanUserLeftSignal = "DoordarshanUserLeft"
const preMeetingDetailsUrl string = "/v1/preMeetingDetails"
const DefaultRateLimitQps = 10

const (
	RedisPayloadNameKey          = "name"
	RedisPayloadSignalKey        = "signal"
	RedisPayloadRoomIDKey        = "roomId"
	RedisPayloadParticipantIDKey = "participantId"
	RedisMessageTypeKey          = "messageType"
	RedisSignalMessageType       = "signalMessage"

	RedisRequestIDType  = "requestId"
	RoomStreamKeyFormat = "room-stream:%s"
)

type DoordarshanStreamPublishPayload struct {
	PublishReason string            `json:"publishReason"`
	ProducerMap   map[string]string `json:"producerMap"`
	ParticipantID string            `json:"participantId"`
}

type DoordarshanStreamUnPublishPayload struct {
	UnPublishReason string            `json:"unPublishReason"`
	ProducerMap     map[string]string `json:"producerMap"`
	ParticipantID   string            `json:"participantId"`
}

type DoordarshanUserJoinedPayload struct {
	UserJoinedReason string `json:"userJoinedReason"`
}

type DoordarshanUserLeftPayload struct {
	DoordarshanUserLeftPayload string `json:"userLeftReason"`
	InstanceID                 string `json:"instanceId"`
}

type ContainerDownSignal struct {
	ContainerId string   `json:"container_id"`
	MeetingId   []string `json:"meeting_id"`
}

type ProducerSignal struct {
	ParticipantId string          `json:"participant_id"`
	MeetingId     string          `json:"meeting_id"`
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload"`
}

type DoordarshanRecreateBulkTransportResponse struct {
	ProducerTransportOptions json.RawMessage `json:"producerTransportOptions"`
	ConsumerTransportOptions json.RawMessage `json:"consumerTransportOptions"`
}

type ApiRateLimiter struct {
	createMeetingRateLimiter                *rate.Limiter
	joinMeetingRateLimiter                  *rate.Limiter
	getProducersOfMeetingRateLimiter        *rate.Limiter
	createProducerTransportRateLimiter      *rate.Limiter
	connectProducerTransportRateLimiter     *rate.Limiter
	createProducerRateLimiter               *rate.Limiter
	createConsumerTransportRateLimiter      *rate.Limiter
	connectConsumerTransportRateLimiter     *rate.Limiter
	recreateProducerTransportRateLimiter    *rate.Limiter
	recreateConsumerTransportRateLimiter    *rate.Limiter
	recreateTransportsRateLimiter           *rate.Limiter
	restartProducerIceRateLimiter           *rate.Limiter
	restartConsumerIceRateLimiter           *rate.Limiter
	restartIceRateLimiter                   *rate.Limiter
	createConsumerRateLimiter               *rate.Limiter
	resumeConsumerRateLimiter               *rate.Limiter
	resumeProducerRateLimiter               *rate.Limiter
	pauseProducerRateLimiter                *rate.Limiter
	pauseConsumerRateLimiter                *rate.Limiter
	closeProducerRateLimiter                *rate.Limiter
	closeAllProducersRateLimiter            *rate.Limiter
	closeAllConsumersForProducerRateLimiter *rate.Limiter
	closeConsumerRateLimiter                *rate.Limiter
	leaveMeetingRateLimiter                 *rate.Limiter
	endMeetingRateLimiter                   *rate.Limiter
	getRTPCapabilitiesRateLimiter           *rate.Limiter
	preMeetingDetailsRateLimiter            *rate.Limiter
}

type SignalIncoming struct {
	Type        string          `json:"type"`
	SubType     string          `json:"subType"`
	VersionID   int64           `json:"version_id"`
	RecipientID []string        `json:"recipient_id"`
	Payload     json.RawMessage `json:"payload"`
}

type meetingV1Handler struct {
	validator               *request.Validator
	client                  *http.Client
	tenantClusterHandlerMap map[string]sfu.IClusterHandler
	baseClusterHandler      *sfu.BaseClusterHandler
	requestGenerator        sfu.PayloadGenerator
	containerDownChannel    chan ContainerDownSignal
	producerSignalChannel   chan ProducerSignal
	apiRateLimiter          *ApiRateLimiter
	logger                  *appLog.Logger
	signallingRedisRepo     data.RedisRepository
}

func NewMeetingV1Handler(
	tenantClusterHandlerMap map[string]sfu.IClusterHandler,
	baseClusterHandler *sfu.BaseClusterHandler,
	appConfig *common.AppConfig,
	logger *appLog.Logger,
	redisRepo data.RedisRepository) IMeetingV1Handler {
	/*retryPolicy := failsafehttp.RetryPolicyBuilder().
	WithDelay(time.Second).
	WithMaxRetries(3).
	Build()*/
	/*circuitBreaker := circuitbreaker.Builder[*http.Response]().
	WithFailureThreshold(5).
	WithDelay(time.Minute).
	WithSuccessThreshold(2).Build()*/
	config := appConfig.RateLimitConfig
	client := &http.Client{
		Timeout: time.Millisecond * 1500,
		//Transport: failsafehttp.NewRoundTripper(http.DefaultTransport, circuitBreaker),
	}
	apiRateLimiter := ApiRateLimiter{
		createMeetingRateLimiter:                rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		joinMeetingRateLimiter:                  rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		getProducersOfMeetingRateLimiter:        rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		createProducerTransportRateLimiter:      rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		connectProducerTransportRateLimiter:     rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		createProducerRateLimiter:               rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		createConsumerTransportRateLimiter:      rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		connectConsumerTransportRateLimiter:     rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		recreateProducerTransportRateLimiter:    rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		recreateConsumerTransportRateLimiter:    rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		recreateTransportsRateLimiter:           rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		restartProducerIceRateLimiter:           rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		restartConsumerIceRateLimiter:           rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		restartIceRateLimiter:                   rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		createConsumerRateLimiter:               rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		resumeConsumerRateLimiter:               rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		resumeProducerRateLimiter:               rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		pauseProducerRateLimiter:                rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		pauseConsumerRateLimiter:                rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		closeProducerRateLimiter:                rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		closeAllProducersRateLimiter:            rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		closeAllConsumersForProducerRateLimiter: rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		closeConsumerRateLimiter:                rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		leaveMeetingRateLimiter:                 rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		endMeetingRateLimiter:                   rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		getRTPCapabilitiesRateLimiter:           rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
		preMeetingDetailsRateLimiter:            rate.NewLimiter(rate.Every(time.Second), config.DefaultApiRateLimitQps),
	}
	m := &meetingV1Handler{
		validator:               request.NewValidator(),
		client:                  client,
		tenantClusterHandlerMap: tenantClusterHandlerMap,
		baseClusterHandler:      baseClusterHandler,
		requestGenerator:        sfu.PayloadGenerator{},
		containerDownChannel:    make(chan ContainerDownSignal, 100),
		producerSignalChannel:   make(chan ProducerSignal, 40000),
		apiRateLimiter:          &apiRateLimiter,
		logger:                  logger,
		signallingRedisRepo:     redisRepo,
	}
	go m.CheckHealthOfContainers()
	go m.RoutineOnContainerDownChannel()
	go m.RoutineOnProducerChannel()
	return m
}

func (m *meetingV1Handler) validate(ec echo.Context, payload interface{}) error {
	if err := ec.Bind(payload); err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while binding the Request: %v", err)
		return HandleCommonError(err)
	}
	if err := m.validator.Validate(payload); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return echo.NewHTTPError(http.StatusBadRequest, validationErrors)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (m *meetingV1Handler) RoutineOnProducerChannel() {
	ctx := context.Background() // Background context - trace IDs won't be available in goroutines
	for {
		msg, ok := <-m.producerSignalChannel
		if !ok {
			m.logger.Info(ctx, "Channel closed")
			return
		}
		signalingDetails := signaling_platform.SignalDetails{
			Payload:   msg.Payload,
			Timestamp: time.Now().UnixNano(),
		}
		broadcastSignal := signaling_platform.BroadcastSignalRequest{
			Type:        msg.Type,
			SubType:     msg.Type,
			ClientID:    DDKTenantForSignalingPlatform,
			RequestID:   msg.ParticipantId + strconv.FormatInt(time.Now().UnixNano(), 10),
			Details:     signalingDetails,
			SenderID:    msg.ParticipantId,
			ReceiverIDs: nil,
			VersionInfo: signaling_platform.VersionInfo{
				SkipVersion: true,
			},
			RoomID: msg.MeetingId,
		}
		startTime := time.Now()
		err := m.handleBroadcastSignal(ctx, msg.MeetingId, &broadcastSignal)
		if err != nil {
			m.logger.Errorf(ctx, "Error while broadcasting signal to signaling platform: %v", err)
			continue
		}
		m.logger.Infof(ctx, "Time taken to broadcast signal to signaling platform: %v", time.Since(startTime))
	}

}

func (m *meetingV1Handler) RoutineOnContainerDownChannel() {
	ctx := context.Background()
	for {
		msg, ok := <-m.containerDownChannel
		if !ok {
			m.logger.Info(ctx, "Channel closed")
			return
		}
		for _, meetingId := range msg.MeetingId {
			broadcastSignal := signaling_platform.BroadcastSignalRequest{
				Type:        DoordarshanDisconnectSignal,
				SubType:     DoordarshanDisconnectSignal,
				ClientID:    DDKTenantForSignalingPlatform,
				RequestID:   msg.ContainerId,
				Details:     signaling_platform.SignalDetails{},
				SenderID:    "system",
				ReceiverIDs: nil,
				VersionInfo: signaling_platform.VersionInfo{
					SkipVersion: true,
				},
				RoomID: meetingId,
			}
			err := m.handleBroadcastSignal(ctx, meetingId, &broadcastSignal)
			if err != nil {
				m.logger.Errorf(ctx, "Error while broadcasting container down signal to signaling platform: %v", err)
				continue
			}
		}
		m.logger.Infof(ctx, "Container %s is down", msg.ContainerId)
	}
}

func (m *meetingV1Handler) CheckHealthOfContainers() {
	ctx := context.Background()
	for {
		containerHeartbeats, _, containerTenantMap, err := m.baseClusterHandler.MonitorContainers()
		if err != nil {
			m.logger.Errorf(ctx, "Error while monitoring containers: %v", err)
			continue
		}
		borderTime := time.Now().Add(-10 * time.Second).UnixNano()
		for k, v := range containerHeartbeats {
			if v < borderTime {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				m.logger.Errorf(ctx, "Container %s is dead", k)
				clusterHandler, err := m.getClusterHandler(containerTenantMap[k])
				// Create new resources for the meetings
				host, newContainerId, newWorkerIds, err := clusterHandler.GetReplaceContainers(clusterHandler, ctx)
				if err != nil {
					m.logger.Errorf(ctx, "Error while getting replace containers: %v", err)
					continue
				}
				meetingConfigurations, existingWorkerIds, err := clusterHandler.MeetingConfiguration(ctx, k)
				if err != nil {
					m.logger.Errorf(ctx, "Error while getting meeting configuration: %v", err)
					continue
				}
				if len(newWorkerIds) < len(existingWorkerIds) {
					m.logger.Error(ctx, "New worker ids are less than existing worker ids")
					continue
				}
				var workerIdMap = make(map[string]string, 0)
				for i := 0; i < len(existingWorkerIds); i++ {
					workerIdMap[existingWorkerIds[i]] = newWorkerIds[i]
				}
				replaceInputs := make(map[string][]sfu.ReplaceMeetingPayloadInput, 0)
				replaceMeetingRequests := make([]sfu.ReplaceMeetingRequest, 0)
				replacedMeetingIds := make([]string, 0)
				for meetingId, workerConfigurations := range meetingConfigurations {
					for _, workerConfiguration := range workerConfigurations {
						if workerConfiguration.Capacity != 0 {
							replaceInput := sfu.ReplaceMeetingPayloadInput{
								WorkerId:       workerIdMap[workerConfiguration.WorkerId],
								RouterCapacity: workerConfiguration.Capacity,
							}
							replaceInputs[meetingId] = append(replaceInputs[meetingId], replaceInput)
						} else if workerConfiguration.ProducerCapacity != 0 {
							replaceInput := sfu.ReplaceMeetingPayloadInput{
								WorkerId:               workerIdMap[workerConfiguration.WorkerId],
								ProducerRouterCapacity: workerConfiguration.ProducerCapacity,
							}
							replaceInputs[meetingId] = append(replaceInputs[meetingId], replaceInput)
						} else if workerConfiguration.ConsumerCapacity != 0 {
							replaceInput := sfu.ReplaceMeetingPayloadInput{
								WorkerId:               workerIdMap[workerConfiguration.WorkerId],
								ConsumerRouterCapacity: workerConfiguration.ConsumerCapacity,
							}
							replaceInputs[meetingId] = append(replaceInputs[meetingId], replaceInput)
						}
					}
					if len(replaceInputs[meetingId]) > 0 {
						replacedMeetingIds = append(replacedMeetingIds, meetingId)
						payload := m.requestGenerator.GetReplaceMeetingPayload(meetingId, k, newContainerId, replaceInputs[meetingId])
						replaceMeetingRequests = append(replaceMeetingRequests, payload)
					} else {
						m.logger.Errorf(ctx, "While replacing no worker configuration found for meeting: %s", meetingId)
					}
				}
				if len(replaceMeetingRequests) > 0 {
					replaceMeetingsPayload := sfu.ReplaceMeetingsRequest{
						ReplaceMeetingRequests: replaceMeetingRequests,
					}
					if err != nil {
						m.logger.Errorf(ctx, "Error while marshalling the payload: %v", err)
						continue
					}
					callCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
					defer cancel()
					_, err = m.doDDCall(callCtx, replaceMeetingsPayload, host, replaceMeetingsUrl)
					if err != nil {
						m.logger.Errorf(ctx, "Error while replacing meetings: %v", err)
						continue
					}
					m.containerDownChannel <- ContainerDownSignal{
						ContainerId: k,
						MeetingId:   replacedMeetingIds,
					}
				}
				cancel()
			}
		}
		time.Sleep(5 * time.Second)
	}
}

// CreateMeeting creates a new meeting with specified capacity
// @Summary      Create a new meeting
// @Description  Creates a new meeting with the specified meeting ID and capacity. The capacity can be specified as overall meeting capacity, or separately as producer and consumer capacities.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateMeetingRequestV1  true  "Create Meeting Request"
// @Success      200      {object}  map[string]interface{}         "Meeting created successfully"
// @Failure      400      {object}  map[string]interface{}         "Bad request - validation failed or invalid tenant"
// @Failure      500      {object}  map[string]interface{}         "Internal server error"
// @Router       /v1/createMeeting [post]
func (m *meetingV1Handler) CreateMeeting(ec echo.Context) error {
	createMeetingRequest := new(request.CreateMeetingRequestV1)
	err := m.validate(ec, createMeetingRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(createMeetingRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant for meetingID: %s error: %s", createMeetingRequest.MeetingId, err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if (createMeetingRequest.ProducerMeetingCapacity != 0 && createMeetingRequest.ConsumerMeetingCapacity == 0) ||
		(createMeetingRequest.ConsumerMeetingCapacity != 0 && createMeetingRequest.ProducerMeetingCapacity == 0) {
		m.logger.Errorf(ec.Request().Context(), "Producer and Consumer capacity should be present for meetingID: %s", createMeetingRequest.MeetingId)
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("producer and Consumer capacity should be present"))
	}
	if createMeetingRequest.ProducerMeetingCapacity == 0 &&
		createMeetingRequest.ConsumerMeetingCapacity == 0 && createMeetingRequest.MeetingCapacity == 0 {
		m.logger.Errorf(ec.Request().Context(), "Capacity should be present for meetingID: %s", createMeetingRequest.MeetingId)
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("capacity should be present"))
	}
	meetingCapacity, meetingProducerCapacity, meetingConsumerCapacity, err := clusterHandler.GetCalculatedMeetingCapacity(
		createMeetingRequest.MeetingCapacity, createMeetingRequest.ProducerMeetingCapacity, createMeetingRequest.ConsumerMeetingCapacity)
	startTime := time.Now()
	freeContainerResponse, err := clusterHandler.GetFreeContainer(clusterHandler, ec.Request().Context(), sfu.FreeContainerRequest{
		Capacity:         meetingCapacity,
		ProducerCapacity: meetingProducerCapacity,
		ConsumerCapacity: meetingConsumerCapacity,
	})
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get free container: %v for meetingID: %s", time.Since(startTime), createMeetingRequest.MeetingId)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while getting free container for meetingID: %s error: %v", createMeetingRequest.MeetingId, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	payload := m.requestGenerator.GetCreateMeetingPayload(createMeetingRequest.MeetingId, *freeContainerResponse)
	err = m.wrapperForDDCall(ec, payload, freeContainerResponse.Host, createMeetingUrl)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while creating meetingID: %s error: %v", createMeetingRequest.MeetingId, err)
	}
	return err
}

// JoinMeeting allows a participant to join an existing meeting
// @Summary      Join an existing meeting
// @Description  Allows a participant to join an existing meeting. The system allocates producer and consumer router IDs for the participant.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.JoinMeetingRequestV1  true  "Join Meeting Request"
// @Success      200      {object}  map[string]interface{}         "Participant joined successfully"
// @Failure      400      {object}  map[string]interface{}         "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}         "Meeting not found"
// @Failure      500      {object}  map[string]interface{}         "Internal server error"
// @Router       /v1/joinMeeting [post]
func (m *meetingV1Handler) JoinMeeting(ec echo.Context) error {
	joinMeetingRequest := new(request.JoinMeetingRequestV1)
	err := m.validate(ec, joinMeetingRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(joinMeetingRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant for meetingID: %s err: %v", joinMeetingRequest.MeetingId, err)
		return err
	}
	startTime := time.Now()
	host, producerRouterId, consumerRouterId, err :=
		clusterHandler.AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId(ec.Request().Context(),
			joinMeetingRequest.MeetingId, joinMeetingRequest.ParticipantId)
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to allocate producer and consumer router: %v", time.Since(startTime))
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Errorf(ec.Request().Context(), "Meeting not found for meetingID: %s error: %v", joinMeetingRequest.MeetingId, err)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			m.logger.Errorf(ec.Request().Context(), "Error while allocating producer and consumer router for meetingID: %s error: %v", joinMeetingRequest.MeetingId, err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	payload := m.requestGenerator.GetJoinMeetingPayload(joinMeetingRequest.ParticipantId, joinMeetingRequest.InstanceId,
		producerRouterId, consumerRouterId, joinMeetingRequest.MeetingId)
	err = m.wrapperForDDCall(ec, payload, host, joinMeetingUrl)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while joining meetingID: %s error: %v", joinMeetingRequest.MeetingId, err)
	}
	return err
}

// GetProducersOfMeeting retrieves all producers in a meeting
// @Summary      Get all producers of a meeting
// @Description  Retrieves a list of all active producers (audio/video streams) in the specified meeting.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.GetProducersOfMeetingRequestV1  true  "Get Producers Request"
// @Success      200      {object}  map[string]interface{}                "List of producers"
// @Failure      400      {object}  map[string]interface{}                "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}                "Meeting not found"
// @Failure      500      {object}  map[string]interface{}                "Internal server error"
// @Router       /v1/getProducersOfMeeting [post]
func (m *meetingV1Handler) GetProducersOfMeeting(ec echo.Context) error {
	getProducersOfMeetingRequest := new(request.GetProducersOfMeetingRequestV1)
	err := m.validate(ec, getProducersOfMeetingRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(getProducersOfMeetingRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
	}
	startTime := time.Now()
	host, err := clusterHandler.GetContainerIPForMeetingId(ec.Request().Context(), getProducersOfMeetingRequest.MeetingId)
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Error(ec.Request().Context(), "Meeting not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP: %v", time.Since(startTime))
	payload := m.requestGenerator.GetGetProducersOfMeetingPayload(getProducersOfMeetingRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, getProducersOfMeetingUrl)

}

// ConnectProducerTransport connects the producer transport with DTLS parameters
// @Summary      Connect producer transport
// @Description  Connects the producer transport using DTLS parameters. This establishes the WebRTC transport connection for sending media.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.ConnectTransportRequestV1  true  "Connect Transport Request"
// @Success      200      {object}  map[string]interface{}              "Transport connected successfully"
// @Failure      400      {object}  map[string]interface{}              "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}              "Router not found"
// @Failure      500      {object}  map[string]interface{}              "Internal server error"
// @Router       /v1/connectProducerTransport [post]
func (m *meetingV1Handler) ConnectProducerTransport(ec echo.Context) error {
	connectTransportRequest := new(request.ConnectTransportRequestV1)
	err := m.validate(ec, connectTransportRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(connectTransportRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, transportId, err := clusterHandler.GetContainerIPProducerTransportIdForParticipantId(ec.Request().Context(), connectTransportRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP and transport ID: %v", time.Since(startTime))
	payload := m.requestGenerator.ConnectTransportRequest(connectTransportRequest.ParticipantId,
		connectTransportRequest.DtlsParameters, transportId, connectTransportRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, connectTransportUrl)
}

// ConnectConsumerTransport connects the consumer transport with DTLS parameters
// @Summary      Connect consumer transport
// @Description  Connects the consumer transport using DTLS parameters. This establishes the WebRTC transport connection for receiving media.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.ConnectTransportRequestV1  true  "Connect Transport Request"
// @Success      200      {object}  map[string]interface{}              "Transport connected successfully"
// @Failure      400      {object}  map[string]interface{}              "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}              "Router not found"
// @Failure      500      {object}  map[string]interface{}              "Internal server error"
// @Router       /v1/connectConsumerTransport [post]
func (m *meetingV1Handler) ConnectConsumerTransport(ec echo.Context) error {
	connectTransportRequest := new(request.ConnectTransportRequestV1)
	err := m.validate(ec, connectTransportRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(connectTransportRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, transportId, err := clusterHandler.GetContainerIPConsumerTransportIdForParticipantId(ec.Request().Context(), connectTransportRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP and transport ID: %v", time.Since(startTime))
	payload := m.requestGenerator.ConnectTransportRequest(connectTransportRequest.ParticipantId,
		connectTransportRequest.DtlsParameters, transportId, connectTransportRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, connectTransportUrl)
}

// RecreateProducerTransport recreates the producer transport for a participant
// @Summary      Recreate producer transport
// @Description  Recreates the producer transport for a participant. This is useful when the transport connection needs to be re-established.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateTransportRequestV1  true  "Recreate Transport Request"
// @Success      200      {object}  map[string]interface{}              "Transport recreated successfully"
// @Failure      400      {object}  map[string]interface{}              "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}              "Router not found"
// @Failure      500      {object}  map[string]interface{}              "Internal server error"
// @Router       /v1/recreateProducerTransport [post]
func (m *meetingV1Handler) RecreateProducerTransport(ec echo.Context) error {
	createTransportRequest := new(request.CreateTransportRequestV1)
	err := m.validate(ec, createTransportRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(createTransportRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant for meetingID: %s err: %v", createTransportRequest.MeetingId, err)
		return err
	}
	startTime := time.Now()
	host, routerId, oldProducerTransportId, oldConsumerTransportId, oldProducerIds, oldConsumerIds, err :=
		clusterHandler.AllocateProducerRouterIdForMeetingAndParticipantId(ec.Request().Context(),
			createTransportRequest.MeetingId, createTransportRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Errorf(ec.Request().Context(), "Router not found for meetingID: %s", createTransportRequest.MeetingId)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			m.logger.Errorf(ec.Request().Context(), "Error while allocating producer router for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to allocate producer router: %v", time.Since(startTime))
	payload := m.requestGenerator.GetCreateTransportRequest(sfu.GetCreateTransportRequestInput{
		MeetingId:              createTransportRequest.MeetingId,
		ParticipantId:          createTransportRequest.ParticipantId,
		InstanceId:             createTransportRequest.InstanceId,
		RouterId:               routerId,
		OldProducerTransportId: oldProducerTransportId,
		OldConsumerTransportId: oldConsumerTransportId,
		OldProducerIds:         oldProducerIds,
		OldConsumerIds:         oldConsumerIds,
	})
	body, err := m.doDDCall(ec.Request().Context(), payload, host, recreateProducerTransportUrl)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	m.logger.Infof(ec.Request().Context(), "Response from DD: %s", string(body))
	recreateProducerTransportResponse, err := m.requestGenerator.RecreateProducerTransportResponse(body)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	m.logger.Infof(ec.Request().Context(), "Unmarshalled Response from DD: %v", recreateProducerTransportResponse)
	if recreateProducerTransportResponse == nil {
		m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response for meetingID: %s", createTransportRequest.MeetingId)
		return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while unmarshalling response"))
	}
	if len(recreateProducerTransportResponse.ClosedProducerDetails) != 0 {
		var producerIDMap = make(map[string]string)
		for _, data := range recreateProducerTransportResponse.ClosedProducerDetails {
			producerIDMap[data.Kind] = data.ProducerID
		}
		recreateProducerPayload, err := json.Marshal(DoordarshanStreamUnPublishPayload{
			ProducerMap:     producerIDMap,
			ParticipantID:   createTransportRequest.ParticipantId,
			UnPublishReason: "Recreating producer transport",
		})
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while marshalling recreate producer payload for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		recreateProducerRawMessage := json.RawMessage(recreateProducerPayload)
		producerSignal := ProducerSignal{
			ParticipantId: createTransportRequest.ParticipantId,
			MeetingId:     createTransportRequest.MeetingId,
			Type:          DoordarshanStreamUnPublishSignal,
			Payload:       recreateProducerRawMessage,
		}
		m.producerSignalChannel <- producerSignal
	}
	byteData, err := json.Marshal(&recreateProducerTransportResponse.TransportOptions)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while marshalling response to byte array for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
		return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while marshalling response to byte array"))
	}
	return ec.JSONBlob(http.StatusOK, byteData)
}

// RecreateConsumerTransport recreates the consumer transport for a participant
// @Summary      Recreate consumer transport
// @Description  Recreates the consumer transport for a participant. This is useful when the transport connection needs to be re-established.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateTransportRequestV1  true  "Recreate Transport Request"
// @Success      200      {object}  map[string]interface{}              "Transport recreated successfully"
// @Failure      400      {object}  map[string]interface{}              "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}              "Router not found"
// @Failure      500      {object}  map[string]interface{}              "Internal server error"
// @Router       /v1/recreateConsumerTransport [post]
func (m *meetingV1Handler) RecreateConsumerTransport(ec echo.Context) error {
	createTransportRequest := new(request.CreateTransportRequestV1)
	err := m.validate(ec, createTransportRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(createTransportRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant for meetingID: %s err: %v", createTransportRequest.MeetingId, err)
		return err
	}
	startTime := time.Now()
	host, routerId, oldProducerTransportId, oldConsumerTransportId, oldProducerIds, oldConsumerIds, err :=
		clusterHandler.AllocateConsumerRouterIdForMeetingAndParticipantId(ec.Request().Context(),
			createTransportRequest.MeetingId, createTransportRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Errorf(ec.Request().Context(), "Router not found for meetingID: %s", createTransportRequest.MeetingId)
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			m.logger.Errorf(ec.Request().Context(), "Error while allocating consumer router for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to allocate consumer router: %v", time.Since(startTime))
	payload := m.requestGenerator.GetCreateTransportRequest(sfu.GetCreateTransportRequestInput{
		ParticipantId:          createTransportRequest.ParticipantId,
		InstanceId:             createTransportRequest.InstanceId,
		RouterId:               routerId,
		MeetingId:              createTransportRequest.MeetingId,
		OldProducerTransportId: oldProducerTransportId,
		OldConsumerTransportId: oldConsumerTransportId,
		OldProducerIds:         oldProducerIds,
		OldConsumerIds:         oldConsumerIds,
	})
	err = m.wrapperForDDCall(ec, payload, host, recreateConsumerTransportUrl)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while recreating consumer transport for meetingID: %s error: %v", createTransportRequest.MeetingId, err)
	}
	return err
}

// RestartProducerIce restarts ICE connection for producer transport
// @Summary      Restart producer ICE
// @Description  Restarts the ICE (Interactive Connectivity Establishment) connection for the producer transport. Used to recover from connection issues.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.RestartIceRequestV1  true  "Restart ICE Request"
// @Success      200      {object}  map[string]interface{}       "ICE restarted successfully"
// @Failure      400      {object}  map[string]interface{}       "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}       "Router not found"
// @Failure      500      {object}  map[string]interface{}       "Internal server error"
// @Router       /v1/restartProducerIce [post]
func (m *meetingV1Handler) RestartProducerIce(ec echo.Context) error {
	restartIceRequest := new(request.RestartIceRequestV1)
	err := m.validate(ec, restartIceRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(restartIceRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerTransportId, err := clusterHandler.GetContainerIPProducerTransportIdForParticipantId(ec.Request().Context(), restartIceRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP and transport ID: %v", time.Since(startTime))
	payload := m.requestGenerator.GetRestartIceRequest(restartIceRequest.ParticipantId, []string{producerTransportId})
	return m.wrapperForDDCall(ec, payload, host, restartProducerIceUrl)
}

// RestartConsumerIce restarts ICE connection for consumer transport
// @Summary      Restart consumer ICE
// @Description  Restarts the ICE (Interactive Connectivity Establishment) connection for the consumer transport. Used to recover from connection issues.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.RestartIceRequestV1  true  "Restart ICE Request"
// @Success      200      {object}  map[string]interface{}       "ICE restarted successfully"
// @Failure      400      {object}  map[string]interface{}       "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}       "Router not found"
// @Failure      500      {object}  map[string]interface{}       "Internal server error"
// @Router       /v1/restartConsumerIce [post]
func (m *meetingV1Handler) RestartConsumerIce(ec echo.Context) error {
	restartIceRequest := new(request.RestartIceRequestV1)
	err := m.validate(ec, restartIceRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(restartIceRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, transportId, err := clusterHandler.GetContainerIPConsumerTransportIdForParticipantId(ec.Request().Context(), restartIceRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP and transport ID: %v", time.Since(startTime))
	payload := m.requestGenerator.GetRestartIceRequest(restartIceRequest.ParticipantId, []string{transportId})
	return m.wrapperForDDCall(ec, payload, host, restartConsumerIceUrl)
}

// RestartIce restarts ICE connection for both producer and consumer transports
// @Summary      Restart ICE for all transports
// @Description  Restarts the ICE (Interactive Connectivity Establishment) connection for both producer and consumer transports of a participant.
// @Tags         Transport
// @Accept       json
// @Produce      json
// @Param        request  body      request.RestartIceRequestV1  true  "Restart ICE Request"
// @Success      200      {object}  map[string]interface{}       "ICE restarted successfully"
// @Failure      400      {object}  map[string]interface{}       "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}       "Router not found"
// @Failure      500      {object}  map[string]interface{}       "Internal server error"
// @Router       /v1/restartIce [post]
func (m *meetingV1Handler) RestartIce(ec echo.Context) error {
	restartIceRequest := new(request.RestartIceRequestV1)
	err := m.validate(ec, restartIceRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(restartIceRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, transportIds, err := clusterHandler.GetContainerIPTransportIdsForParticipantId(ec.Request().Context(), restartIceRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container IP and transport ID: %v", time.Since(startTime))
	payload := m.requestGenerator.GetRestartIceRequest(restartIceRequest.ParticipantId, transportIds)
	return m.wrapperForDDCall(ec, payload, host, restartIceUrl)
}

// CreateProducer creates a new media producer (audio/video stream)
// @Summary      Create a media producer
// @Description  Creates a new media producer for a participant. This allows the participant to publish their audio/video stream to the meeting.
// @Tags         Producer
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateProducerRequestV1  true  "Create Producer Request"
// @Success      200      {object}  map[string]interface{}            "Producer created successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/createProducer [post]
func (m *meetingV1Handler) CreateProducer(ec echo.Context) error {
	createProducerRequest := new(request.CreateProducerRequestV1)
	err := m.validate(ec, createProducerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(createProducerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, routerId, consumerRouterIds, producerTransportId, err := clusterHandler.GetCreateProducerDetails(ec.Request().Context(),
		createProducerRequest.ParticipantId, createProducerRequest.MeetingId)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get create producer details: %v", time.Since(startTime))
	input := sfu.CreateProducerInput{
		ParticipantId:       createProducerRequest.ParticipantId,
		Kind:                createProducerRequest.Kind,
		RtpParameters:       createProducerRequest.RtpParameters,
		MeetingID:           createProducerRequest.MeetingId,
		ProducerRouterId:    routerId,
		ProducerTransportId: producerTransportId,
		ConsumerRouterIds:   consumerRouterIds,
		StartRecording:      createProducerRequest.StartRecording,
		InstanceId:          createProducerRequest.InstanceId,
	}
	payload := m.requestGenerator.CreateProduceRequest(input)
	startTime = time.Now()
	body, err := m.doDDCall(ec.Request().Context(), payload, host, createProducerUrl)
	m.logger.Infof(ec.Request().Context(), "Time taken to create producer on DD: %v", time.Since(startTime))
	if err == nil {
		startTime = time.Now()
		createProducerResponse, err := m.requestGenerator.ProducerDetails(body)
		m.logger.Infof(ec.Request().Context(), "Time taken to unmarshal producer details: %v", time.Since(startTime))
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if createProducerResponse == nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while unmarshalling response"))
		}
		var producerIDMap = make(map[string]string)
		producerIDMap[createProducerResponse.Kind] = createProducerResponse.ProducerID
		startTime = time.Now()
		createProducerPayload, err := json.Marshal(DoordarshanStreamPublishPayload{
			ProducerMap:   producerIDMap,
			ParticipantID: createProducerRequest.ParticipantId,
			PublishReason: "User created the stream",
		})
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while marshalling recreate producer payload: %v", err)
			return err
		}
		m.logger.Infof(ec.Request().Context(), "Time taken to marshal producer details: %v", time.Since(startTime))
		streamPublishRawMessage := json.RawMessage(createProducerPayload)
		startTime = time.Now()
		m.producerSignalChannel <- ProducerSignal{
			ParticipantId: createProducerRequest.ParticipantId,
			MeetingId:     createProducerRequest.MeetingId,
			Type:          DoordarshanStreamPublishSignal,
			Payload:       streamPublishRawMessage,
		}
		m.logger.Infof(ec.Request().Context(), "Time taken to send producer signal: %v", time.Since(startTime))
		return ec.JSONBlob(http.StatusOK, body)
	}
	return err
}

// CreateConsumer creates a new media consumer to receive streams from other participants
// @Summary      Create a media consumer
// @Description  Creates a new media consumer for a participant to receive audio/video streams from target participants in the meeting.
// @Tags         Consumer
// @Accept       json
// @Produce      json
// @Param        request  body      request.CreateConsumerRequestV1  true  "Create Consumer Request"
// @Success      200      {object}  map[string]interface{}            "Consumer created successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/createConsumer [post]
func (m *meetingV1Handler) CreateConsumer(ec echo.Context) error {
	createConsumerRequest := new(request.CreateConsumerRequestV1)
	err := m.validate(ec, createConsumerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(createConsumerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, consumerTransportId, producerInfo, err := clusterHandler.GetCreateConsumerDetails(ec.Request().Context(), createConsumerRequest.ParticipantId,
		createConsumerRequest.MeetingId, createConsumerRequest.TargetParticipants)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get create consumer details: %v", time.Since(startTime))
	payload := m.requestGenerator.CreateConsumeRequest(createConsumerRequest.ParticipantId,
		createConsumerRequest.RtpCapabilities,
		consumerTransportId,
		producerInfo,
		createConsumerRequest.MeetingId,
		createConsumerRequest.InstanceId)
	return m.wrapperForDDCall(ec, payload, host, createConsumerUrl)
}

// ResumeConsumer resumes a paused consumer
// @Summary      Resume a consumer
// @Description  Resumes a previously paused consumer, allowing the participant to receive media streams again.
// @Tags         Consumer
// @Accept       json
// @Produce      json
// @Param        request  body      request.ResumeConsumerRequestV1  true  "Resume Consumer Request"
// @Success      200      {object}  map[string]interface{}            "Consumer resumed successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/resumeConsumer [post]
func (m *meetingV1Handler) ResumeConsumer(ec echo.Context) error {
	resumeConsumerRequest := new(request.ResumeConsumerRequestV1)
	err := m.validate(ec, resumeConsumerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(resumeConsumerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, consumerIds, err := clusterHandler.GetConsumerIds(ec.Request().Context(),
		resumeConsumerRequest.ParticipantId, resumeConsumerRequest.MeetingId, resumeConsumerRequest.TargetParticipants)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get consumer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.ResumeConsumeRequest(resumeConsumerRequest.ParticipantId,
		consumerIds,
		resumeConsumerRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, resumeConsumerUrl)
}

// ResumeProducer resumes a paused producer
// @Summary      Resume a producer
// @Description  Resumes a previously paused producer, allowing the participant to publish their media stream again.
// @Tags         Producer
// @Accept       json
// @Produce      json
// @Param        request  body      request.ResumeProducerRequestV1  true  "Resume Producer Request"
// @Success      200      {object}  map[string]interface{}            "Producer resumed successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/resumeProducer [post]
func (m *meetingV1Handler) ResumeProducer(ec echo.Context) error {
	resumeProducerRequest := new(request.ResumeProducerRequestV1)
	err := m.validate(ec, resumeProducerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(resumeProducerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerIdKindMap, err := clusterHandler.GetProducerIds(ec.Request().Context(),
		resumeProducerRequest.ParticipantId, resumeProducerRequest.MeetingId, resumeProducerRequest.Kinds)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get producer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.ResumeProduceRequest(resumeProducerRequest.ParticipantId,
		producerIdKindMap, resumeProducerRequest.MeetingId, resumeProducerRequest.RecordingEnabled)
	body, err := m.doDDCall(ec.Request().Context(), payload, host, resumeProducerUrl)
	if err == nil {
		resumeProducerResponse, err := m.requestGenerator.ResumeProducerResponse(body)
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if resumeProducerResponse == nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while unmarshalling response"))
		}
		var producerIDKindMap = make(map[string]string)
		for _, data := range resumeProducerResponse.ProducerDetails {
			producerIDKindMap[data.Kind] = data.ProducerID
		}
		resumeProducer, err := json.Marshal(DoordarshanStreamPublishPayload{
			ProducerMap:   producerIDKindMap,
			ParticipantID: resumeProducerRequest.ParticipantId,
			PublishReason: "User resumed the stream",
		})
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while marshalling user joined payload: %v", err)
			return err
		}
		resumeProducerRawMessage := json.RawMessage(resumeProducer)
		m.producerSignalChannel <- ProducerSignal{
			ParticipantId: resumeProducerRequest.ParticipantId,
			MeetingId:     resumeProducerRequest.MeetingId,
			Type:          DoordarshanStreamPublishSignal,
			Payload:       resumeProducerRawMessage,
		}
	}
	return err
}

// PauseProducer pauses a producer temporarily
// @Summary      Pause a producer
// @Description  Pauses a producer temporarily, stopping the media stream from being published without closing it.
// @Tags         Producer
// @Accept       json
// @Produce      json
// @Param        request  body      request.PauseProducerRequestV1  true  "Pause Producer Request"
// @Success      200      {object}  map[string]interface{}            "Producer paused successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/pauseProducer [post]
func (m *meetingV1Handler) PauseProducer(ec echo.Context) error {
	pauseProducerRequest := new(request.PauseProducerRequestV1)
	err := m.validate(ec, pauseProducerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(pauseProducerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerIds, err := clusterHandler.GetProducerIds(ec.Request().Context(),
		pauseProducerRequest.ParticipantId, pauseProducerRequest.MeetingId, pauseProducerRequest.Kinds)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get producer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.PauseProducerRequest(pauseProducerRequest.ParticipantId,
		producerIds, pauseProducerRequest.MeetingId, pauseProducerRequest.RecordingEnabled)
	body, err := m.doDDCall(ec.Request().Context(), payload, host, pauseProducerUrl)
	if err == nil {
		pauseProducerResponse, err := m.requestGenerator.PauseProducerResponse(body)
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if pauseProducerResponse == nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while unmarshalling response"))
		}
		var producerIDKindMap = make(map[string]string)
		for _, data := range pauseProducerResponse.ProducerDetails {
			producerIDKindMap[data.Kind] = data.ProducerID
		}
		pauseProducerPayload, err := json.Marshal(DoordarshanStreamUnPublishPayload{
			ProducerMap:     producerIDKindMap,
			ParticipantID:   pauseProducerRequest.ParticipantId,
			UnPublishReason: "User paused the stream",
		})
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while marshalling user joined payload: %v", err)
			return err
		}
		pauseProducerRawMessage := json.RawMessage(pauseProducerPayload)
		m.producerSignalChannel <- ProducerSignal{
			ParticipantId: pauseProducerRequest.ParticipantId,
			MeetingId:     pauseProducerRequest.MeetingId,
			Type:          DoordarshanStreamUnPublishSignal,
			Payload:       pauseProducerRawMessage,
		}
	}
	return err
}

// PauseConsumer pauses a consumer temporarily
// @Summary      Pause a consumer
// @Description  Pauses a consumer temporarily, stopping the reception of media streams without closing it.
// @Tags         Consumer
// @Accept       json
// @Produce      json
// @Param        request  body      request.PauseConsumerRequestV1  true  "Pause Consumer Request"
// @Success      200      {object}  map[string]interface{}            "Consumer paused successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/pauseConsumer [post]
func (m *meetingV1Handler) PauseConsumer(ec echo.Context) error {
	pauseConsumerRequest := new(request.PauseConsumerRequestV1)
	err := m.validate(ec, pauseConsumerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(pauseConsumerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, consumerIds, err := clusterHandler.GetConsumerIds(ec.Request().Context(), pauseConsumerRequest.ParticipantId,
		pauseConsumerRequest.MeetingId, pauseConsumerRequest.TargetParticipants)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get consumer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.PauseConsumerRequest(pauseConsumerRequest.ParticipantId,
		consumerIds, pauseConsumerRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, pauseConsumerUrl)
}

// CloseConsumer closes a consumer permanently
// @Summary      Close a consumer
// @Description  Closes a consumer permanently, stopping the reception of media streams and freeing resources.
// @Tags         Consumer
// @Accept       json
// @Produce      json
// @Param        request  body      request.CloseConsumerRequestV1  true  "Close Consumer Request"
// @Success      200      {object}  map[string]interface{}            "Consumer closed successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/closeConsumer [post]
func (m *meetingV1Handler) CloseConsumer(ec echo.Context) error {
	closeConsumerRequest := new(request.CloseConsumerRequestV1)
	err := m.validate(ec, closeConsumerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(closeConsumerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, consumerIds, err := clusterHandler.GetConsumerIds(ec.Request().Context(), closeConsumerRequest.ParticipantId,
		closeConsumerRequest.MeetingId, closeConsumerRequest.TargetParticipants)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get consumer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.CloseConsumerRequest(closeConsumerRequest.ParticipantId,
		consumerIds, closeConsumerRequest.MeetingId)
	return m.wrapperForDDCall(ec, payload, host, closeConsumerUrl)
}

// CloseProducer closes a producer permanently
// @Summary      Close a producer
// @Description  Closes a producer permanently, stopping the media stream and freeing resources.
// @Tags         Producer
// @Accept       json
// @Produce      json
// @Param        request  body      request.CloseProducerRequestV1  true  "Close Producer Request"
// @Success      200      {object}  map[string]interface{}            "Producer closed successfully"
// @Failure      400      {object}  map[string]interface{}            "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}            "Router not found"
// @Failure      500      {object}  map[string]interface{}            "Internal server error"
// @Router       /v1/closeProducer [post]
func (m *meetingV1Handler) CloseProducer(ec echo.Context) error {
	closeProducerRequest := new(request.CloseProducerRequestV1)
	err := m.validate(ec, closeProducerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(closeProducerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerIdKindMap, err := clusterHandler.GetProducerIds(ec.Request().Context(),
		closeProducerRequest.ParticipantId, closeProducerRequest.MeetingId, closeProducerRequest.Kinds)
	if err != nil {
		if errors.Is(err, myerrors.RouterNotFound) {
			m.logger.Error(ec.Request().Context(), "Router not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get producer ids: %v", time.Since(startTime))
	payload := m.requestGenerator.CloseProducerRequest(closeProducerRequest.ParticipantId,
		producerIdKindMap, closeProducerRequest.MeetingId)
	body, err := m.doDDCall(ec.Request().Context(), payload, host, closeProducerUrl)
	if err == nil {
		closeProducerResponse, err := m.requestGenerator.CloseProducerResponse(body)
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if closeProducerResponse == nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("Error while unmarshalling response"))
		}
		var producerIDMap = make(map[string]string)
		for _, data := range closeProducerResponse.ProducerDetails {
			producerIDMap[data.Kind] = data.ProducerID
		}
		closeProducerPayload, err := json.Marshal(DoordarshanStreamUnPublishPayload{
			ProducerMap:     producerIDMap,
			ParticipantID:   closeProducerRequest.ParticipantId,
			UnPublishReason: "User left the meeting",
		})
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while marshalling user joined payload: %v", err)
			return err
		}
		closeProducerRawMessage := json.RawMessage(closeProducerPayload)
		m.producerSignalChannel <- ProducerSignal{
			ParticipantId: closeProducerRequest.ParticipantId,
			MeetingId:     closeProducerRequest.MeetingId,
			Type:          DoordarshanStreamUnPublishSignal,
			Payload:       closeProducerRawMessage,
		}
		/*closeProducerUserLeftPayload, err := json.Marshal(DoordarshanUserLeftPayload{
			DoordarshanUserLeftPayload: "User left the meeting",
		})
		if err != nil {
			 m.logger.Errorf(ec.Request().Context(), "Error while marshalling user joined payload: %v", err) }
			return err
		}		closeProducerUserLeftRawMessage := json.RawMessage(closeProducerUserLeftPayload)
		m.producerSignalChannel <- ProducerSignal{
			ParticipantId: closeProducerRequest.ParticipantId,
			MeetingId:     closeProducerRequest.MeetingId,
			Type:          DoordarshanUserLeftSignal,
			Payload:       closeProducerUserLeftRawMessage,
		}*/
	}
	return err
}

// LeaveMeeting allows a participant to leave a meeting
// @Summary      Leave a meeting
// @Description  Allows a participant to leave a meeting. This closes all their producers and consumers and frees associated resources.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.LeaveMeetingRequest  true  "Leave Meeting Request"
// @Success      200      {object}  map[string]interface{}      "Participant left successfully"
// @Failure      400      {object}  map[string]interface{}      "Bad request - validation failed"
// @Failure      500      {object}  map[string]interface{}      "Internal server error"
// @Router       /v1/leaveMeeting [post]
func (m *meetingV1Handler) LeaveMeeting(ec echo.Context) error {
	leaveMeetingRequest := new(request.LeaveMeetingRequest)
	err := m.validate(ec, leaveMeetingRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(leaveMeetingRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerTransportId, consumerTransportId, producerIds, consumerIds, err :=
		clusterHandler.GetLeaveMeetingDetails(ec.Request().Context(), leaveMeetingRequest.ParticipantId, leaveMeetingRequest.MeetingId)
	if err != nil {
		if errors.Is(err, myerrors.ParticipantNotFound) {
			return ec.JSON(http.StatusOK, "Participant not present in the meeting")
		} else {
			return ec.JSON(http.StatusInternalServerError, "Internal server error")
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get leave meeting details: %v", time.Since(startTime))
	payload := m.requestGenerator.LeaveMeetingRequest(leaveMeetingRequest.ParticipantId, leaveMeetingRequest.InstanceId, leaveMeetingRequest.MeetingId,
		producerTransportId, consumerTransportId, producerIds, consumerIds)
	body, err := m.doDDCall(ec.Request().Context(), payload, host, leaveMeetingUrl)
	if err == nil {
		leaveMeetingResponse, err := m.requestGenerator.LeaveMeetingResponse(body)
		if err != nil {
			m.logger.Errorf(ec.Request().Context(), "Error while unmarshalling response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("error while unmarshalling response"))
		}
		if leaveMeetingResponse == nil {
			m.logger.Error(ec.Request().Context(), "leaveMeetingResponse is nil")
			return echo.NewHTTPError(http.StatusInternalServerError, errors.New("leaveMeetingResponse is nil"))
		}
		if leaveMeetingResponse.SendSignal {
			userLeftPayload, err := json.Marshal(DoordarshanUserLeftPayload{
				DoordarshanUserLeftPayload: "User left the meeting",
				InstanceID:                 leaveMeetingRequest.InstanceId,
			})
			if err != nil {
				m.logger.Errorf(ec.Request().Context(), "Error while marshalling user left payload: %v", err)
				return err
			}
			userLeftRawMessage := json.RawMessage(userLeftPayload)
			m.producerSignalChannel <- ProducerSignal{
				ParticipantId: leaveMeetingRequest.ParticipantId,
				MeetingId:     leaveMeetingRequest.MeetingId,
				Type:          DoordarshanUserLeftSignal,
				Payload:       userLeftRawMessage,
			}
		}
	}
	return err
}

// EndMeeting ends a meeting and closes all resources
// @Summary      End a meeting
// @Description  Ends a meeting permanently, closing all participants, producers, consumers, and freeing all associated resources.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.EndMeetingRequest  true  "End Meeting Request"
// @Success      200      {object}  map[string]interface{}     "Meeting ended successfully"
// @Failure      400      {object}  map[string]interface{}     "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}     "Meeting not found"
// @Failure      500      {object}  map[string]interface{}     "Internal server error"
// @Router       /v1/endMeeting [post]
func (m *meetingV1Handler) EndMeeting(ec echo.Context) error {
	endMeetingRequest := new(request.EndMeetingRequest)
	err := m.validate(ec, endMeetingRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(endMeetingRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, routerIds, producerTransportIds, consumerTransportIds, producers, consumers, participantIds, err :=
		clusterHandler.GetCompleteMeetingInfo(ec.Request().Context(), endMeetingRequest.MeetingId)
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Error(ec.Request().Context(), "Meeting not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get complete meeting info: %v", time.Since(startTime))
	var transportIds []string
	transportIds = append(transportIds, producerTransportIds...)
	transportIds = append(transportIds, consumerTransportIds...)
	payload := m.requestGenerator.EndMeetingRequest(endMeetingRequest.MeetingId, routerIds, transportIds, producers, consumers, participantIds)
	return m.wrapperForDDCall(ec, payload, host, endMeetingUrl)
}

// GetRTPCapabilities retrieves RTP capabilities for a meeting
// @Summary      Get RTP capabilities
// @Description  Retrieves the RTP (Real-time Transport Protocol) capabilities for a meeting, which are required for WebRTC negotiation.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.GetRTPCapabilitiesRequest  true  "Get RTP Capabilities Request"
// @Success      200      {object}  map[string]interface{}              "RTP capabilities"
// @Failure      400      {object}  map[string]interface{}              "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}              "Meeting not found"
// @Failure      500      {object}  map[string]interface{}              "Internal server error"
// @Router       /v1/getRTPCapabilities [post]
func (m *meetingV1Handler) GetRTPCapabilities(ec echo.Context) error {
	getRTPCapabilitiesRequest := new(request.GetRTPCapabilitiesRequest)
	err := m.validate(ec, getRTPCapabilitiesRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(getRTPCapabilitiesRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerRouterID, consumerRouterId, err := clusterHandler.AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId(
		ec.Request().Context(), getRTPCapabilitiesRequest.MeetingId, getRTPCapabilitiesRequest.ParticipantId)
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Error(ec.Request().Context(), "Meeting not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to allocate producer and consumer router: %v", time.Since(startTime))
	payload := m.requestGenerator.GetRTPCapabilitiesRequest(getRTPCapabilitiesRequest.MeetingId, producerRouterID, consumerRouterId)
	return m.wrapperForDDCall(ec, payload, host, getRTPCapabilitiesUrl)
}

// PreMeetingDetails retrieves pre-meeting details for a participant
// @Summary      Get pre-meeting details
// @Description  Retrieves pre-meeting details including router information for a participant before they join.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.PreMeetingDetailsRequest  true  "Pre Meeting Details Request"
// @Success      200      {object}  map[string]interface{}                "Pre-meeting details"
// @Failure      400      {object}  map[string]interface{}                "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}                "Meeting not found"
// @Failure      500      {object}  map[string]interface{}                "Internal server error"
// @Router       /v1/preMeetingDetails [post]
func (m *meetingV1Handler) PreMeetingDetails(ec echo.Context) error {
	preMeetingDetailsRequest := new(request.PreMeetingDetailsRequest)
	err := m.validate(ec, preMeetingDetailsRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(preMeetingDetailsRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	host, producerRouterId, consumerRouterId, err := clusterHandler.AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId(
		ec.Request().Context(), preMeetingDetailsRequest.MeetingId, preMeetingDetailsRequest.ParticipantID)
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Error(ec.Request().Context(), "Meeting not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to allocate producer and consumer router: %v", time.Since(startTime))
	payload := m.requestGenerator.PreMeetingDetailsRequest(preMeetingDetailsRequest.MeetingId, producerRouterId, consumerRouterId, preMeetingDetailsRequest.ParticipantID)
	return m.wrapperForDDCall(ec, payload, host, preMeetingDetailsUrl)
}

// GetActiveContainerOfMeeting retrieves the active container ID for a meeting
// @Summary      Get active container of meeting
// @Description  Retrieves the container ID where the meeting is currently hosted. Useful for monitoring and debugging.
// @Tags         Meeting
// @Accept       json
// @Produce      json
// @Param        request  body      request.GetActiveContainerOfMeetingRequest  true  "Get Active Container Request"
// @Success      200      {object}  response.GetActiveContainerOfMeetingResponse  "Active container information"
// @Failure      400      {object}  map[string]interface{}                        "Bad request - validation failed"
// @Failure      404      {object}  map[string]interface{}                        "Meeting not found"
// @Failure      500      {object}  map[string]interface{}                        "Internal server error"
// @Router       /v1/activeContainer [post]
func (m *meetingV1Handler) GetActiveContainerOfMeeting(ec echo.Context) error {
	containerRequest := new(request.GetActiveContainerOfMeetingRequest)
	err := m.validate(ec, containerRequest)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Validation failed : %v", err)
		return err
	}
	clusterHandler, err := m.getClusterHandler(containerRequest.Tenant)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Invalid tenant: %v", err)
		return err
	}
	startTime := time.Now()
	containerId, err := clusterHandler.GetContainerIdForMeetingId(ec.Request().Context(), containerRequest.MeetingId)
	if err != nil {
		if errors.Is(err, myerrors.MeetingNotFound) {
			m.logger.Error(ec.Request().Context(), "Meeting not found")
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	m.logger.Infof(ec.Request().Context(), "Mysql Time taken to get container ID: %v", time.Since(startTime))
	activeContainerResponse := &response.GetActiveContainerOfMeetingResponse{
		ContainerId: containerId,
		MeetingId:   containerRequest.MeetingId,
	}
	jsonResponse, err := json.Marshal(activeContainerResponse)
	if err != nil {
		m.logger.Error(ec.Request().Context(), "Error marshalling JSON")
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return ec.JSONBlob(http.StatusOK, jsonResponse)
}

type IMeetingV1Handler interface {
	CreateMeeting(ec echo.Context) error
	JoinMeeting(ec echo.Context) error
	GetProducersOfMeeting(ec echo.Context) error
	RecreateProducerTransport(ec echo.Context) error
	ConnectProducerTransport(ec echo.Context) error
	RecreateConsumerTransport(ec echo.Context) error
	ConnectConsumerTransport(ec echo.Context) error
	RestartProducerIce(ec echo.Context) error
	RestartConsumerIce(ec echo.Context) error
	RestartIce(ec echo.Context) error
	CreateProducer(ec echo.Context) error
	CreateConsumer(ec echo.Context) error
	ResumeConsumer(ec echo.Context) error
	ResumeProducer(ec echo.Context) error
	PauseProducer(ec echo.Context) error
	PauseConsumer(ec echo.Context) error
	CloseConsumer(ec echo.Context) error
	CloseProducer(ec echo.Context) error
	LeaveMeeting(ec echo.Context) error
	EndMeeting(ec echo.Context) error
	GetRTPCapabilities(ec echo.Context) error
	PreMeetingDetails(ec echo.Context) error
	GetActiveContainerOfMeeting(ec echo.Context) error
}

func (m *meetingV1Handler) getClusterHandler(tenant string) (sfu.IClusterHandler, error) {
	handler, ok := m.tenantClusterHandlerMap[tenant]
	if ok {
		return handler, nil
	}
	return nil, errors.New("tenant not found")
}

func (m *meetingV1Handler) wrapperForDDCall(ec echo.Context, payload interface{}, host string, url string) error {
	body, err := m.doDDCall(ec.Request().Context(), payload, host, url)
	if err != nil {
		m.logger.Errorf(ec.Request().Context(), "Error while making request: %v", err)
		return err
	}
	return ec.JSONBlob(http.StatusOK, body)
}

func (m *meetingV1Handler) doDDCall(ctx context.Context, payload interface{}, host string, url string) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		m.logger.Errorf(ctx, "Error while marshalling the Request: %v", err)
		return nil, HandleCommonError(err)
	}
	m.logger.Infof(ctx, "Url, Payload: %s, %s", url, string(payloadBytes))
	ddRequest, err := http.NewRequestWithContext(ctx,
		http.MethodPost, "http://"+host+":3001"+url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		m.logger.Errorf(ctx, "Error while creating request: %v", err)
		return nil, err
	}
	startTime := time.Now()

	ddResponse, err := m.client.Do(ddRequest)
	defer func() {
		if ddResponse != nil && ddResponse.Body != nil {
			if err := ddResponse.Body.Close(); err != nil {
				m.logger.Warn(ctx, "Error closing response body", zap.Error(err))
			}
		}
	}()
	if err != nil {
		m.logger.Errorf(ctx, "Error while making request: %v", err)
		return nil, err
	}
	endTime := time.Since(startTime)
	m.logger.Infof(ctx, "Time taken for making request %s: %v", url, endTime)
	if ddResponse.StatusCode >= 400 {
		m.logger.Errorf(ctx, "Error while making request: %v", err)
		return nil, errors.New("error while making request")
	}
	body, err := io.ReadAll(ddResponse.Body)
	return body, nil
}

func (sh *meetingV1Handler) handleBroadcastSignal(ctx context.Context, roomIDStr string, broadcastRequest *signaling_platform.BroadcastSignalRequest) error {
	participantID := broadcastRequest.SenderID
	signalType := broadcastRequest.Type
	signalSubType := broadcastRequest.SubType

	var (
		roomSignalVersionID int64 = 0
		err                 error
	)

	additionalParams := make(map[string]string)
	additionalParams[RedisRequestIDType] = broadcastRequest.RequestID

	broadcastPayloadStr, err := common.ToJSONString(broadcastRequest.Details.Payload)
	if err != nil {
		log.Errorf("[BroadcastSignal] Error while converting the broadcastRequest to JSON string for for roomId "+
			"%s signal type %s subType %s: %v", roomIDStr, signalType, signalSubType, err)
		return errors.New("error while marshalling signal payload")
	}

	incomingSignal := SignalIncoming{
		Type:        signalType,
		SubType:     signalSubType,
		VersionID:   roomSignalVersionID,
		RecipientID: broadcastRequest.ReceiverIDs,
		Payload:     []byte(broadcastPayloadStr),
	}

	incomingSignalStr, err := common.ToJSONString(incomingSignal)
	if err != nil {
		log.Errorf("[BroadcastSignal] Error while converting the incomingSignal to JSON string for for roomId "+
			"%s signal type %s subType %s: %v", roomIDStr, signalType, signalSubType, err)
		return errors.New("error while marshalling signal payload")
	}

	streamKey := fmt.Sprintf(RoomStreamKeyFormat, roomIDStr)
	messageID, err := sh.signallingRedisRepo.PushToStream(
		ctx,
		streamKey, NewDirectedSignalRedisPayload(
			signalType,
			incomingSignalStr,
			roomIDStr,
			participantID,
			additionalParams,
		).Value,
	)
	if err != nil {
		log.Errorf("[BroadcastSignal] Error while pushing the payload to the stream: %v", err)
		return errors.New("error while pushing the payload to the stream")
	}
	log.Debugf("[BroadcastSignal] Successfully pushed the payload to the stream: %s", messageID)

	log.Infof("Broadcasting signal to all participants in roomID: %s and signal type %s and subType %s versionId %d", roomIDStr, signalType,
		signalSubType, roomSignalVersionID)
	return nil
}

type RedisPayload struct {
	Value map[string]interface{}
}

func NewDirectedSignalRedisPayload(name, jsonPayloadString, roomID, participantID string, extraParameters map[string]string) *RedisPayload {
	payloadMap := make(map[string]interface{})
	payloadMap[RedisPayloadNameKey] = name
	payloadMap[RedisMessageTypeKey] = RedisSignalMessageType
	if roomID != constant.EmptyStr {
		payloadMap[RedisPayloadRoomIDKey] = roomID
	}
	if participantID != constant.EmptyStr {
		payloadMap[RedisPayloadParticipantIDKey] = participantID
	}

	payloadMap[RedisPayloadSignalKey] = jsonPayloadString
	for key, value := range extraParameters {
		payloadMap[key] = value
	}

	return &RedisPayload{
		Value: payloadMap,
	}
}
