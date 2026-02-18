package sfu

import (
	"context"
	"database/sql"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data"
	"github.com/labstack/gommon/log"
)

type IClusterHandler interface {
	GetTenant() string
	SetDb(db *sql.DB)
	GetDb() *sql.DB
	GetFreeContainer(self IClusterHandler, ctx context.Context, request FreeContainerRequest) (*FreeContainersResponse, error)
	GetReplaceContainers(self IClusterHandler, ctx context.Context) (string, string, []string, error)
	MonitorContainers(self IClusterHandler, ctx context.Context) (map[string]int64, map[string][]string, error)
	GetContainerIPForMeetingId(ctx context.Context, meetingId string) (string, error)
	GetContainerIdForMeetingId(ctx context.Context, meetingId string) (string, error)
	AllocateConsumerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, string, []string, []string, error)
	AllocateProducerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, string, []string, []string, error)
	AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, error)
	GetContainerIPProducerConsumerRouterIdForParticipantId(ctx context.Context, participantId string) (string, string, string, error)
	GetContainerIPProducerTransportIdForParticipantId(ctx context.Context, participantId string) (string, string, error)
	GetContainerIPConsumerTransportIdForParticipantId(ctx context.Context, participantId string) (string, string, error)
	GetContainerIPTransportIdsForParticipantId(ctx context.Context, participantId string) (string, []string, error)
	GetCreateProducerDetails(ctx context.Context, participantId string, meetingId string) (string, string, []string, string, error)
	GetCreateConsumerDetails(ctx context.Context, participantId string, meetingId string, targetParticipants map[string][]string) (string, string, map[string]map[string]string, error)
	GetLeaveMeetingDetails(ctx context.Context, participantId string, meetingId string) (string, string, string, []string, []string, error)
	GetConsumerIds(ctx context.Context, participantId string, meetingId string, participantKindMap map[string][]string) (string, []string, error)
	GetProducerIds(ctx context.Context, participantId string, meetingId string, kinds []string) (string, map[string]string, error)
	GetCompleteMeetingInfo(ctx context.Context, meetingId string) (string, []string, []string, []string, []string, []string, []string, error)
	MeetingConfiguration(ctx context.Context, containerId string) (map[string][]WorkerConfiguration, []string, error)
	GetCalculatedMeetingCapacity(capacity int, producerCapacity int, consumerCapacity int) (int, int, int, error)
}

func UniqueStrings(input []string) []string {
	set := make(map[string]struct{}, len(input))
	result := make([]string, 0, len(input))
	for _, val := range input {
		if _, ok := set[val]; !ok {
			set[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}

func NewBaseClusterHandler(client *data.MySQLClient) *BaseClusterHandler {
	return &BaseClusterHandler{Db: client.MysqlDb}
}

type BaseClusterHandler struct {
	Db *sql.DB
}
type UtilisedContainerInfo struct {
	ContainerId string
	Host        string
	WorkerId    string
	Capacity    int
}
type TempContainerCapacity struct {
	ContainerId       string
	Host              string
	WorkerIds         []string
	Capacity          int
	PerWorkerCapacity []int
}

func (b *BaseClusterHandler) queryDb(ctx context.Context, query string) (*sql.Rows, error) {
	rows, err := b.Db.QueryContext(ctx, query)
	if err != nil {
		log.Error("Error while querying with query %s ", query, err)
		return nil, err
	}
	return rows, nil
}

func (b *BaseClusterHandler) MonitorContainers() (map[string]int64, map[string][]string, map[string]string, error) {
	ctx := context.Background()
	query := "select distinct t4.container_id, t4.heartbeat, t3.meeting_id, t4.tenant " +
		"from worker_container t1 " +
		"JOIN router_worker t2 ON t1.worker_id = t2.worker_id " +
		"JOIN router_meeting t3 ON t2.router_id = t3.router_id " +
		"JOIN container_heartbeat t4 ON t1.container_id = t4.container_id "
	//log.Info("Query is ", query)
	rows, err := b.queryDb(ctx, query)
	if err != nil {
		log.Error("Error while querying db", err)
		return nil, nil, nil, err
	}
	defer rows.Close()
	containerHeartbeatMap := make(map[string]int64)
	containerMeetingsMap := make(map[string][]string)
	containerTenantMap := make(map[string]string)
	for rows.Next() {
		var containerId string
		var timeStamp int64
		var meetingId string
		var tenant string
		err = rows.Scan(&containerId, &timeStamp, &meetingId, &tenant)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return nil, nil, nil, err
		}
		if _, ok := containerHeartbeatMap[containerId]; !ok {
			containerHeartbeatMap[containerId] = timeStamp
		}
		if _, ok := containerMeetingsMap[containerId]; !ok {
			containerTenantMap[containerId] = tenant
		}
		if containerMeetingsMap[containerId] == nil {
			containerMeetingsMap[containerId] = []string{}
		}
		containerMeetingsMap[containerId] = append(containerMeetingsMap[containerId], meetingId)
	}
	return containerHeartbeatMap, containerMeetingsMap, containerTenantMap, nil
}

type FreeContainerRequest struct {
	Capacity         int
	ProducerCapacity int
	ConsumerCapacity int
}
type WorkerCapacity struct {
	WorkerId string
	Capacity int
}

type FreeContainersResponse struct {
	ContainerId            string
	Host                   string
	WorkerCapacity         *WorkerCapacity
	ProducerWorkerCapacity *WorkerCapacity
	ConsumerWorkerCapacity []WorkerCapacity
}

type WorkerConfiguration struct {
	WorkerId         string
	Capacity         int
	ProducerCapacity int
	ConsumerCapacity int
}
