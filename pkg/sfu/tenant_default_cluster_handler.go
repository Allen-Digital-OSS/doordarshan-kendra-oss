package sfu

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/myerrors"
	"github.com/labstack/gommon/log"
	"hash/fnv"
	"math"
	"slices"
	"sort"
	"strings"
)

const MaxConsumerCapacity = 500

func (t *TenantDefaultClusterHandler) queryDb(ctx context.Context, query string) (*sql.Rows, error) {
	rows, err := t.Db.QueryContext(ctx, query)
	if err != nil {
		log.Error("Error while querying with query %s ", query, err)
		return nil, err
	}
	return rows, nil
}

type TenantDefaultClusterHandler struct {
	Db *sql.DB
}

func (t *TenantDefaultClusterHandler) GetTenant() string {
	return "default"
}

func (t *TenantDefaultClusterHandler) SetDb(db *sql.DB) {
	t.Db = db
}

func (t *TenantDefaultClusterHandler) GetDb() *sql.DB {
	return t.Db
}

func (t *TenantDefaultClusterHandler) GetFreeContainer(self IClusterHandler, ctx context.Context, request FreeContainerRequest) (*FreeContainersResponse, error) {
	consumerCapacity := request.ConsumerCapacity
	producerCapacity := request.ProducerCapacity
	query := "select t1.container_id AS consumer_container_id, " +
		"t1.host AS consumer_host, " +
		"t2.worker_id AS consumer_worker_id, " +
		"COALESCE(SUM(t4.consumer_capacity),0) AS total_consumer_capacity " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t1.tenant = ' " + self.GetTenant() + "' AND FROM_UNIXTIME(t1.heartbeat/1000000000) > now() - interval 10 second " +
		"GROUP BY t1.container_id, t1.host, t2.worker_id " +
		"HAVING total_consumer_capacity < 500;"
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error while querying db", err)
		return nil, err
	}
	defer rows.Close()
	var utilisedConsumersCapacityInfo = make([]UtilisedContainerInfo, 0)
	for rows.Next() {
		var consumerContainerId string
		var consumerHost string
		var consumerWorkerId string
		var totalConsumerCapacity int
		err = rows.Scan(&consumerContainerId, &consumerHost, &consumerWorkerId, &totalConsumerCapacity)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return nil, err
		}

		utilisedConsumersCapacityInfo = append(utilisedConsumersCapacityInfo, UtilisedContainerInfo{
			ContainerId: consumerContainerId,
			Host:        consumerHost,
			WorkerId:    consumerWorkerId,
			Capacity:    totalConsumerCapacity,
		})
	}
	response := FreeContainersResponse{}
	var containerIds = make(map[string][]UtilisedContainerInfo)
	for _, consumer := range utilisedConsumersCapacityInfo {
		if containerIds[consumer.ContainerId] == nil {
			containerIds[consumer.ContainerId] = []UtilisedContainerInfo{}
		}
		containerIds[consumer.ContainerId] = append(containerIds[consumer.ContainerId], consumer)
	}
	var (
		tempContainerCapacities        = make([]TempContainerCapacity, 0)
		maxAllocatedContainerWorkerMap = make(map[string]string, 0)
	)

	for containerId, consumers := range containerIds {
		maxAllocatedCapacity := math.MinInt
		maxAllocatedCapacityWorkerId := ""
		for _, c := range consumers {
			if c.Capacity > maxAllocatedCapacity {
				maxAllocatedCapacity = c.Capacity
				maxAllocatedCapacityWorkerId = c.WorkerId
			}
		}
		var filteredConsumers []UtilisedContainerInfo
		for _, consumer := range consumers {
			if consumer.WorkerId == maxAllocatedCapacityWorkerId {
				log.Infof("Skipping the worker %s as it is the max allocated capacity", consumer.WorkerId)
				maxAllocatedContainerWorkerMap[consumer.ContainerId] = consumer.WorkerId
				continue
			}
			filteredConsumers = append(filteredConsumers, consumer)
		}
		for i := 0; i < len(filteredConsumers); i++ {
			tempWorkerIds := make([]string, 0)
			tempCapacity := 0
			tempPerWorkerCapacity := make([]int, 0)

			for j := i; j < len(filteredConsumers); j++ {
				tempCapacity = tempCapacity + filteredConsumers[j].Capacity
				tempWorkerIds = append(tempWorkerIds, filteredConsumers[j].WorkerId)
				tempPerWorkerCapacity = append(tempPerWorkerCapacity, filteredConsumers[j].Capacity)
				tempContainerCapacities = append(tempContainerCapacities, TempContainerCapacity{
					ContainerId:       containerId,
					Host:              filteredConsumers[j].Host,
					WorkerIds:         tempWorkerIds,
					Capacity:          tempCapacity,
					PerWorkerCapacity: tempPerWorkerCapacity,
				})
			}
		}
	}
	// Sort by Capacity in descending order
	sort.Slice(tempContainerCapacities, func(i, j int) bool {
		return tempContainerCapacities[i].Capacity > tempContainerCapacities[j].Capacity
	})
	for _, tempContainerCapacity := range tempContainerCapacities {
		if (len(tempContainerCapacity.WorkerIds)*MaxConsumerCapacity)-tempContainerCapacity.Capacity >= consumerCapacity {
			response.ContainerId = tempContainerCapacity.ContainerId
			response.Host = tempContainerCapacity.Host
			for index, workerId := range tempContainerCapacity.WorkerIds {
				if consumerCapacity <= 0 {
					log.Debugf("Consumer capacity is exhausted")
					break
				}
				if MaxConsumerCapacity-tempContainerCapacity.PerWorkerCapacity[index] > consumerCapacity {
					response.ConsumerWorkerCapacity = append(response.ConsumerWorkerCapacity, WorkerCapacity{
						WorkerId: workerId,
						Capacity: consumerCapacity,
					})
					consumerCapacity = 0
				} else {
					response.ConsumerWorkerCapacity = append(response.ConsumerWorkerCapacity, WorkerCapacity{
						WorkerId: workerId,
						Capacity: MaxConsumerCapacity - tempContainerCapacity.PerWorkerCapacity[index],
					})
					consumerCapacity = consumerCapacity - (MaxConsumerCapacity - tempContainerCapacity.PerWorkerCapacity[index])
				}
			}
			if consumerCapacity <= 0 {
				log.Debugf("Consumer capacity is exhausted")
				break
			}
		}
	}
	if response.ContainerId == "" {
		log.Error("No free container found")
		err = errors.New("no free container found")
		return nil, err
	}

	producerWorkerId, ok := maxAllocatedContainerWorkerMap[response.ContainerId]
	if !ok || producerWorkerId == "" {
		log.Error("No producer worker found")
		err = errors.New("no producer worker found")
		return nil, err
	}

	response.ProducerWorkerCapacity = &WorkerCapacity{
		WorkerId: producerWorkerId,
		Capacity: producerCapacity,
	}
	log.Infof("Response for free container: %v", response)
	return &response, nil
}

func (t *TenantDefaultClusterHandler) GetReplaceContainers(self IClusterHandler, ctx context.Context) (string, string, []string, error) {
	query := "select t1.container_id, " +
		"t1.host, " +
		"JSON_ARRAYAGG(t2.worker_id), " +
		//"COALESCE(SUM(t4.capacity),0) AS total_capacity, " +
		//"COALESCE(SUM(t4.producer_capacity),0) AS total_producer_capacity, " +
		"COALESCE(SUM(t4.consumer_capacity),0) AS total_consumer_capacity " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t1.tenant = ' " + self.GetTenant() + "' AND FROM_UNIXTIME(t1.heartbeat/1000000000) > now() - interval 10 second " +
		"GROUP BY t1.container_id, t1.host " +
		"ORDER BY total_consumer_capacity ASC LIMIT 1;"
	//log.Info("Query is ", query)
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error while querying db", err)
		return "", "", nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		log.Error("No container found with total capacity 0")
		err = errors.New("no container found with total capacity 0")
		return "", "", nil, err
	}
	var containerId string
	var host string
	var workerIdsJson string
	var capacity int
	err = rows.Scan(&containerId, &host, &workerIdsJson, &capacity)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", "", nil, err
	}
	var workerIds []string
	if err := json.Unmarshal([]byte(workerIdsJson), &workerIds); err != nil {
		log.Error("Error while unmarshalling workerIdsJson", err)
		return "", "", nil, err
	}
	var unique = UniqueStrings(workerIds)
	return host, containerId, unique, nil
}

func (t *TenantDefaultClusterHandler) MeetingConfiguration(ctx context.Context, containerId string) (map[string][]WorkerConfiguration, []string, error) {
	query := "select " +
		"t3.meeting_id as meeting_id, " +
		"t3.router_id as router_id, " +
		"t3.producer_capacity as producer_capacity, " +
		"t3.consumer_capacity as consumer_capacity, " +
		"t3.capacity as capacity, " +
		"t2.worker_id as worker_id " +
		"from worker_container AS t1 " +
		"JOIN router_worker AS t2 ON t1.worker_id = t2.worker_id " +
		"JOIN router_meeting AS t3 ON t2.router_id = t3.router_id " +
		"WHERE t1.container_id='" + containerId + "';"
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error while querying db", err)
		return nil, nil, err
	}
	defer rows.Close()
	meetingWorkerConfig := make(map[string][]WorkerConfiguration)
	workerIds := make([]string, 0)
	for rows.Next() {
		var meetingId string
		var routerId string
		var tempProducerCapacity sql.NullInt64
		var tempConsumerCapacity sql.NullInt64
		var tempCapacity sql.NullInt64
		var workerId string
		err := rows.Scan(&meetingId, &routerId, &tempProducerCapacity, &tempConsumerCapacity, &tempCapacity, &workerId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return nil, nil, err
		}
		var capacity int
		var producerCapacity int
		var consumerCapacity int
		if tempCapacity.Valid {
			capacity = int(tempCapacity.Int64)
		}
		if tempProducerCapacity.Valid {
			producerCapacity = int(tempProducerCapacity.Int64)
		}
		if tempConsumerCapacity.Valid {
			consumerCapacity = int(tempConsumerCapacity.Int64)
		}
		if meetingWorkerConfig[meetingId] == nil {
			meetingWorkerConfig[meetingId] = []WorkerConfiguration{}
		}
		meetingWorkerConfig[meetingId] = append(meetingWorkerConfig[meetingId], WorkerConfiguration{
			WorkerId:         workerId,
			Capacity:         capacity,
			ProducerCapacity: producerCapacity,
			ConsumerCapacity: consumerCapacity,
		})
		if !slices.Contains(workerIds, workerId) {
			workerIds = append(workerIds, workerId)
		}
	}
	return meetingWorkerConfig, workerIds, nil
}

func (t *TenantDefaultClusterHandler) MonitorContainers(self IClusterHandler, ctx context.Context) (map[string]int64, map[string][]string, error) {
	query := "select distinct t4.container_id,t4.heartbeat,t3.meeting_id " +
		"from worker_container t1 " +
		"JOIN router_worker t2 ON t1.worker_id = t2.worker_id " +
		"JOIN router_meeting t3 ON t2.router_id = t3.router_id " +
		"JOIN container_heartbeat t4 ON t1.container_id = t4.container_id AND t4.tenant = '" + self.GetTenant() + "'"
	//log.Info("Query is ", query)
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error while querying db", err)
		return nil, nil, err
	}
	defer rows.Close()
	containerHeartbeatMap := make(map[string]int64)
	containerMeetingsMap := make(map[string][]string)
	for rows.Next() {
		var containerId string
		var timeStamp int64
		var meetingId string
		err = rows.Scan(&containerId, &timeStamp, &meetingId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return nil, nil, err
		}
		if _, ok := containerHeartbeatMap[containerId]; !ok {
			containerHeartbeatMap[containerId] = timeStamp
		}
		if containerMeetingsMap[containerId] == nil {
			containerMeetingsMap[containerId] = []string{}
		}
		containerMeetingsMap[containerId] = append(containerMeetingsMap[containerId], meetingId)
	}
	return containerHeartbeatMap, containerMeetingsMap, nil
}

func (t *TenantDefaultClusterHandler) GetContainerIPForMeetingId(ctx context.Context, meetingId string) (string, error) {
	query := "select t1.host " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id WHERE " +
		"t4.meeting_id='" + meetingId + "'"
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", err
	}
	var host string
	var routerId string
	if !rows.Next() {
		log.Errorf("Meeting not found %s", meetingId)
		return "", myerrors.MeetingNotFound
	}
	err = rows.Scan(&host, &routerId)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", err
	}
	return host, nil
}

func (t *TenantDefaultClusterHandler) GetContainerIdForMeetingId(ctx context.Context, meetingId string) (string, error) {
	query := "select t1.container_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id WHERE " +
		"t4.meeting_id='" + meetingId + "'"
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", err
	}
	var containerId string
	if !rows.Next() {
		log.Errorf("Meeting not found %s", meetingId)
		return "", myerrors.MeetingNotFound
	}
	err = rows.Scan(&containerId)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", err
	}
	return containerId, nil
}

func (t *TenantDefaultClusterHandler) AllocateProducerRouterIdConsumerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, error) {
	consumerQuery := "select t1.host, t3.router_id AS consumer_router_id, 'NA' as producer_router_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id WHERE t4.consumer_capacity IS NOT NULL AND " +
		"t4.meeting_id='" + meetingId + "'"
	producerQuery := "select t1.host, 'NA' as consumer_router_id, t3.router_id AS producer_router_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id WHERE t4.producer_capacity IS NOT NULL AND " +
		"t4.meeting_id='" + meetingId + "'"
	query := consumerQuery + " UNION " + producerQuery
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", "", "", err
	}
	defer rows.Close()
	var producerRouterIds = make([]string, 0)
	var consumerRouterIds = make([]string, 0)
	var host string
	for rows.Next() {
		var tempHost string
		var producerRouterId string
		var consumerRouterId string
		err = rows.Scan(&tempHost, &producerRouterId, &consumerRouterId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", err
		}
		host = tempHost
		if producerRouterId != "NA" {
			producerRouterIds = append(producerRouterIds, producerRouterId)
		}
		if consumerRouterId != "NA" {
			consumerRouterIds = append(consumerRouterIds, consumerRouterId)
		}
	}
	if len(producerRouterIds) == 0 || len(consumerRouterIds) == 0 {
		log.Error("No router found for meeting ", meetingId)
		return "", "", "", myerrors.RouterNotFound
	}
	if len(consumerRouterIds) > 1 {
		h := fnv.New32a()
		_, err := h.Write([]byte(participantId))
		if err != nil {
			log.Error("Error while writing to hash ", err)
			return "", "", "", err
		}
		index := h.Sum32() % uint32(len(consumerRouterIds))
		return host, producerRouterIds[0], consumerRouterIds[index], nil
	} else {
		return host, producerRouterIds[0], consumerRouterIds[0], nil
	}
}

func (t *TenantDefaultClusterHandler) getContainerIPTransportIdForParticipantId(ctx context.Context, participantId string, transportTable string, routerTable string) (string, string, error) {
	query := "select t1.host, t5.transport_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN " + routerTable + " t4 ON t3.router_id = t4.router_id " +
		"LEFT JOIN " + transportTable + " t5 ON t4.participant_id = t5.participant_id " +
		"WHERE t4.participant_id='" + participantId + "'"
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", "", err
	}
	var host string
	var transportId string
	if !rows.Next() {
		log.Errorf("Incorrect producer config for participantId %s", participantId)
		return "", "", myerrors.RouterNotFound
	}
	err = rows.Scan(&host, &transportId)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", "", err
	}
	return host, transportId, err
}

func (t *TenantDefaultClusterHandler) GetContainerIPProducerTransportIdForParticipantId(ctx context.Context, participantId string) (string, string, error) {
	return t.getContainerIPTransportIdForParticipantId(ctx, participantId, "participant_producer_transport", "participant_producer_router")
}

func (t *TenantDefaultClusterHandler) GetContainerIPConsumerTransportIdForParticipantId(ctx context.Context, participantId string) (string, string, error) {
	return t.getContainerIPTransportIdForParticipantId(ctx, participantId, "participant_consumer_transport", "participant_consumer_router")
}

func (t *TenantDefaultClusterHandler) GetContainerIPTransportIdsForParticipantId(ctx context.Context, participantId string) (string, []string, error) {
	query := "select t1.host, t5.transport_id AS producer_transport_id, t6.transport_id AS consumer_transport_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN participant_producer_router t4 ON t3.router_id = t4.router_id " +
		"LEFT JOIN participant_producer_transport t5 ON t4.participant_id = t5.participant_id " +
		"LEFT JOIN participant_consumer_transport t6 ON t4.participant_id = t6.participant_id " +
		"WHERE t4.participant_id='" + participantId + "'"
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", nil, err
	}
	var host string
	var producerTransportId sql.NullString
	var consumerTransportId sql.NullString
	if !rows.Next() {
		log.Errorf("Incorrect producer config for participantId %s", participantId)
		return "", nil, myerrors.RouterNotFound
	}
	err = rows.Scan(&host, &producerTransportId, &consumerTransportId)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", nil, err
	}
	transportIds := make([]string, 0)
	if producerTransportId.Valid {
		transportIds = append(transportIds, producerTransportId.String)
	}
	if consumerTransportId.Valid {
		transportIds = append(transportIds, consumerTransportId.String)
	}
	return host, transportIds, err
}

func (t *TenantDefaultClusterHandler) GetCreateProducerDetails(ctx context.Context, participantId string, meetingId string) (string, string, []string, string, error) {
	query := "select t1.host AS host, t4.router_id AS router_id, " +
		"producer_capacity AS producer_capacity, consumer_capacity AS consumer_capacity, NULL AS producer_transport_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	producerTransportQuery := "select NULL AS host, NULL AS router_id, " +
		"NULL AS producer_capacity, NULL AS consumer_capacity, transport_id AS producer_transport_id " +
		" from participant_producer_transport where participant_id='" + participantId + "'"
	finalQuery := query + " UNION " + producerTransportQuery
	//"' AND FROM_UNIXTIME(t1.heartbeat/1000000000) > (now() - interval 5 minute)"
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", "", nil, "", nil
	}
	var host string
	var producerRouterId string
	var consumerRouterIds = make([]string, 0)
	var producerTransportId string
	for rows.Next() {
		var tempHost sql.NullString
		var tempRouterId sql.NullString
		var tempProducerCapacity sql.NullString
		var tempConsumerCapacity sql.NullString
		var tempProducerTransportId sql.NullString
		err = rows.Scan(&tempHost, &tempRouterId, &tempProducerCapacity, &tempConsumerCapacity, &tempProducerTransportId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", nil, "", err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempProducerCapacity.Valid {
			producerRouterId = tempRouterId.String
		}
		if tempConsumerCapacity.Valid {
			consumerRouterIds = append(consumerRouterIds, tempRouterId.String)
		}
		if tempProducerTransportId.Valid {
			producerTransportId = tempProducerTransportId.String
		}
	}
	if host == "" || producerRouterId == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", "", nil, "", myerrors.MeetingNotFound
	}
	return host, producerRouterId, consumerRouterIds, producerTransportId, err
}

func (t *TenantDefaultClusterHandler) GetCreateConsumerDetails(ctx context.Context, participantId string, meetingId string, targetParticipants map[string][]string) (string, string, map[string]map[string]string, error) {
	query := "select t1.host AS host, NULL AS consumer_transport_id, NULL AS participant_id, NULL AS kind, NULL AS producer_id " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	consumerTransportQuery := "select NULL AS host, transport_id AS consumer_transport_id, NULL AS participant_id, NULL AS kind, NULL AS producer_id " +
		"from participant_consumer_transport where participant_id='" + participantId + "'"
	targetProducerIdQuery := "select NULL AS host, NULL AS consumer_transport_id, participant_id, kind, producer_id " +
		"from participant_producers WHERE "
	for targetParticipant, targetKinds := range targetParticipants {
		for _, kind := range targetKinds {
			targetProducerIdQuery += "(participant_id='" + targetParticipant + "' AND kind='" + kind + "') OR "
		}
	}
	targetProducerIdQuery = strings.TrimSuffix(targetProducerIdQuery, " OR ")
	finalQuery := query + " UNION " + consumerTransportQuery + " UNION " + targetProducerIdQuery
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", "", nil, err
	}
	var host string
	var consumerTransportId string
	consumerDetails := make(map[string]map[string]string)
	for rows.Next() {
		var tempHost sql.NullString
		var tempConsumerTransportId sql.NullString
		var tempParticipantId sql.NullString
		var tempKind sql.NullString
		var tempProducerId sql.NullString
		err = rows.Scan(&tempHost, &tempConsumerTransportId, &tempParticipantId, &tempKind, &tempProducerId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempConsumerTransportId.Valid {
			consumerTransportId = tempConsumerTransportId.String
		}
		if tempParticipantId.Valid {
			if consumerDetails[tempParticipantId.String] == nil {
				consumerDetails[tempParticipantId.String] = make(map[string]string)
			}
			consumerDetails[tempParticipantId.String][tempKind.String] = tempProducerId.String
		}
	}
	if host == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", "", nil, myerrors.MeetingNotFound
	}
	return host, consumerTransportId, consumerDetails, err
}

func (t *TenantDefaultClusterHandler) GetConsumerIds(ctx context.Context, participantId string, meetingId string, participantKindMap map[string][]string) (string, []string, error) {
	query := "select t1.host AS host, NULL AS consumer_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	consumerIdsQuery := "select NULL AS host, group_concat(consumer_id) AS consumer_ids " +
		"from participant_consumers WHERE participant_id='" + participantId + "' AND "
	for targetParticipant, targetKinds := range participantKindMap {
		for _, kind := range targetKinds {
			consumerIdsQuery += "(target_participant_id='" + targetParticipant + "' AND kind='" + kind + "') OR "
		}
	}
	consumerIdsQuery = strings.TrimSuffix(consumerIdsQuery, " OR ")
	finalQuery := query + " UNION " + consumerIdsQuery
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", nil, err
	}
	var host string
	consumerIds := make([]string, 0)
	for rows.Next() {
		var tempHost sql.NullString
		var tempConsumerIds sql.NullString
		err = rows.Scan(&tempHost, &tempConsumerIds)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempConsumerIds.Valid {
			consumerIds = strings.Split(tempConsumerIds.String, ",")
		}
	}
	if host == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", nil, myerrors.MeetingNotFound
	}
	return host, consumerIds, err
}

func (t *TenantDefaultClusterHandler) GetProducerIds(ctx context.Context, participantId string, meetingId string, kinds []string) (string, map[string]string, error) {
	query := "select t1.host AS host, NULL AS producer_id, NULL AS kind " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	producerIdsQuery := "select NULL AS host, producer_id AS producer_id, kind AS kind " +
		"from participant_producers WHERE participant_id='" + participantId + "' AND " + "kind IN ('" + strings.Join(kinds, "','") + "')"
	finalQuery := query + " UNION " + producerIdsQuery
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", nil, err
	}
	var host string
	var producerIds = make(map[string]string)
	for rows.Next() {
		var tempHost sql.NullString
		var tempProducerId sql.NullString
		var tempKind sql.NullString
		err = rows.Scan(&tempHost, &tempProducerId, &tempKind)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempProducerId.Valid {
			producerIds[tempProducerId.String] = tempKind.String
		}
	}
	return host, producerIds, err
}

func (t *TenantDefaultClusterHandler) GetLeaveMeetingDetails(ctx context.Context, participantId string, meetingId string) (string, string, string, []string, []string, error) {
	query := "select MAX(t1.host) as host, t5.transport_id AS producer_transport_id, t6.transport_id AS consumer_transport_id, " +
		"group_concat(t7.producer_id) AS producer_ids, group_concat(t8.consumer_id) AS consumer_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN participant_producer_router t4 ON t3.router_id = t4.router_id " +
		"LEFT JOIN participant_producer_transport t5 ON t4.participant_id = t5.participant_id " +
		"LEFT JOIN participant_consumer_transport t6 ON t4.participant_id = t6.participant_id " +
		"LEFT JOIN participant_producers t7 ON t6.participant_id = t7.participant_id " +
		"LEFT JOIN participant_consumers t8 ON t7.participant_id = t8.participant_id " +
		"WHERE t4.participant_id='" + participantId + "'"
	log.Infof(query)
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error while querying Db", err)
		return "", "", "", nil, nil, err
	}
	var host string
	var producerTransportId string
	var consumerTransportId string
	var producerIds []string
	var consumerIds []string
	for rows.Next() {
		var tempHost sql.NullString
		var tempProducerTransportId sql.NullString
		var tempConsumerTransportId sql.NullString
		var tempProducerIds sql.NullString
		var tempConsumerIds sql.NullString
		err = rows.Scan(&tempHost, &tempProducerTransportId, &tempConsumerTransportId, &tempProducerIds, &tempConsumerIds)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", nil, nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempProducerTransportId.Valid {
			producerTransportId = tempProducerTransportId.String
		}
		if tempConsumerTransportId.Valid {
			consumerTransportId = tempConsumerTransportId.String
		}
		if tempProducerIds.Valid {
			producerIds = strings.Split(tempProducerIds.String, ",")
		}
		if tempConsumerIds.Valid {
			consumerIds = strings.Split(tempConsumerIds.String, ",")
		}
	}
	if host == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", "", "", nil, nil, myerrors.MeetingNotFound
	}
	return host, producerTransportId, consumerTransportId, producerIds, consumerIds, nil
}

func (t *TenantDefaultClusterHandler) GetCompleteMeetingInfo(ctx context.Context, meetingId string) (string, []string, []string, []string, []string, []string, []string, error) {
	query := "select MAX(t1.host), " +
		"JSON_ARRAYAGG(t4.router_id) AS router_ids, " +
		"JSON_ARRAYAGG(t6.transport_id) AS producer_transport_ids, " +
		"JSON_ARRAYAGG(t7.transport_id) AS consumer_transport_ids, " +
		"JSON_ARRAYAGG(t8.producer_id) AS producer_ids, " +
		"JSON_ARRAYAGG(t9.consumer_id) AS consumer_ids, " +
		"JSON_ARRAYAGG(t5.participant_id) AS participant_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"LEFT JOIN participant_meeting t5 ON t5.meeting_id = t4.meeting_id " +
		"LEFT JOIN participant_producer_transport t6 ON t6.participant_id = t5.participant_id " +
		"LEFT JOIN participant_consumer_transport t7 ON t7.participant_id = t5.participant_id " +
		"LEFT JOIN participant_producers t8 ON t8.participant_id = t5.participant_id " +
		"LEFT JOIN participant_consumers t9 ON t9.participant_id = t5.participant_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	log.Info(query)
	rows, err := t.queryDb(ctx, query)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", nil, nil, nil, nil, nil, nil, err
	}
	var host string
	if !rows.Next() {
		log.Errorf("Meeting not found %s", meetingId)
		return "", nil, nil, nil, nil, nil, nil, myerrors.MeetingNotFound
	}
	var tempRouterIds sql.NullString
	var tempProducerTransportIds sql.NullString
	var tempConsumerTransportIds sql.NullString
	var tempProducerIds sql.NullString
	var tempConsumerIds sql.NullString
	var tempParticipantIds sql.NullString
	err = rows.Scan(&host, &tempRouterIds, &tempProducerTransportIds, &tempConsumerTransportIds, &tempProducerIds, &tempConsumerIds, &tempParticipantIds)
	if err != nil {
		log.Error("Error while scanning rows", err)
		return "", nil, nil, nil, nil, nil, nil, err
	}
	var routerIds []string
	var producerTransportIds []string
	var consumerTransportIds []string
	var producerIds []string
	var consumerIds []string
	var participantIds []string
	if tempRouterIds.Valid {
		err = json.Unmarshal([]byte(tempRouterIds.String), &routerIds)
		if err != nil {
			log.Error("Error while unmarshalling routerIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	if tempProducerTransportIds.Valid {
		err = json.Unmarshal([]byte(tempProducerTransportIds.String), &producerTransportIds)
		if err != nil {
			log.Error("Error while unmarshalling producerTransportIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	if tempConsumerTransportIds.Valid {
		err = json.Unmarshal([]byte(tempConsumerTransportIds.String), &consumerTransportIds)
		if err != nil {
			log.Error("Error while unmarshalling consumerTransportIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	if tempProducerIds.Valid {
		err = json.Unmarshal([]byte(tempProducerIds.String), &producerIds)
		if err != nil {
			log.Error("Error while unmarshalling producerIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	if tempConsumerIds.Valid {
		err = json.Unmarshal([]byte(tempConsumerIds.String), &consumerIds)
		if err != nil {
			log.Error("Error while unmarshalling consumerIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	if tempParticipantIds.Valid {
		err = json.Unmarshal([]byte(tempParticipantIds.String), &participantIds)
		if err != nil {
			log.Error("Error while unmarshalling participantIds", err)
			return "", nil, nil, nil, nil, nil, nil, err
		}
	}
	return host,
		UniqueStrings(routerIds),
		UniqueStrings(producerTransportIds),
		UniqueStrings(consumerTransportIds),
		UniqueStrings(producerIds),
		UniqueStrings(consumerIds),
		UniqueStrings(participantIds),
		nil
}

func (t *TenantDefaultClusterHandler) getContainerIpRouterIdExistingResourcesForMeetingParticipantId(ctx context.Context, meetingId string,
	participantId string, transportTable string) (string, string, string, []string, []string, error) {
	query := "select t1.host AS new_host, t3.router_id AS new_router_id, NULL AS old_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "'"
	existingTransportQuery := "select NULL AS new_host, NULL AS new_router_id, transport_id AS old_transport_id,'NA' AS old_producer_ids, NULL AS old_consumer_ids " +
		"from " + transportTable + " where participant_id='" + participantId + "'"
	existingProducers := "select NULL AS new_host, NULL AS new_router_id, NULL AS old_transport_id, group_concat(producer_id) AS old_producer_ids,NULL AS old_consumer_ids " +
		"from participant_producers " + "where participant_id='" + participantId + "'"
	existingConsumers := "select NULL AS new_host, NULL AS new_router_id, NULL AS old_transport_id, NULL AS old_producer_ids, group_concat(consumer_id) AS old_consumer_ids " +
		"from participant_consumers " + "where participant_id='" + participantId + "'"
	finalQuery := query + " UNION " + existingTransportQuery + " UNION " + existingProducers + " UNION " + existingConsumers
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", "", "", nil, nil, err
	}
	var host string
	var routerId string
	var oldTransportId string
	var oldProducerIds []string
	var oldConsumerIds []string
	for rows.Next() {
		var tempHost sql.NullString
		var tempRouterId sql.NullString
		var tempOldTransportId sql.NullString
		var tempOldProducerIds sql.NullString
		var tempOldConsumerIds sql.NullString
		err = rows.Scan(&tempHost, &tempRouterId, &tempOldTransportId, &tempOldProducerIds, &tempOldConsumerIds)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", nil, nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempRouterId.Valid {
			routerId = tempRouterId.String
		}
		if tempOldTransportId.Valid {
			oldTransportId = tempOldTransportId.String
		}
		if tempOldProducerIds.Valid {
			oldProducerIds = strings.Split(tempOldProducerIds.String, ",")
		}
		if tempOldConsumerIds.Valid {
			oldConsumerIds = strings.Split(tempOldConsumerIds.String, ",")
		}
	}
	if host == "" || routerId == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", "", "", nil, nil, myerrors.MeetingNotFound
	}
	return host, routerId, oldTransportId, oldProducerIds, oldConsumerIds, nil
}

func (t *TenantDefaultClusterHandler) AllocateConsumerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (string, string, string, string, []string, []string, error) {
	query := "select group_concat(t1.host) AS new_host, group_concat(t3.router_id) AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "' AND t4.producer_capacity IS NOT NULL"
	existingProducerTransportQuery := "select NULL AS new_host, NULL AS new_router_id, " +
		"transport_id AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from participant_producer_transport where participant_id='" + participantId + "'"
	existingConsumerTransportQuery := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, transport_id AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from participant_consumer_transport where participant_id='" + participantId + "'"
	existingProducers := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, group_concat(producer_id) AS old_producer_ids,NULL AS old_consumer_ids " +
		"from participant_producers " + "where participant_id='" + participantId + "'"
	existingConsumers := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, group_concat(consumer_id) AS old_consumer_ids " +
		"from participant_consumers " + "where participant_id='" + participantId + "'"
	finalQuery := query + " UNION " + existingProducerTransportQuery + " UNION " +
		existingConsumerTransportQuery + " UNION " + existingProducers + " UNION " + existingConsumers
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", "", "", "", nil, nil, err
	}
	var host string
	var routerIds []string
	var oldProducerTransportId string
	var oldConsumerTransportId string
	var oldProducerIds []string
	var oldConsumerIds []string
	for rows.Next() {
		var tempHost sql.NullString
		var tempRouterId sql.NullString
		var tempOldProducerTransportId sql.NullString
		var tempOldConsumerTransportId sql.NullString
		var tempOldProducerIds sql.NullString
		var tempOldConsumerIds sql.NullString
		err = rows.Scan(&tempHost, &tempRouterId, &tempOldProducerTransportId, &tempOldConsumerTransportId, &tempOldProducerIds, &tempOldConsumerIds)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", "", nil, nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempRouterId.Valid {
			routerIds = UniqueStrings(strings.Split(tempRouterId.String, ",")) // Ensure unique router IDs
		}
		if tempOldProducerTransportId.Valid {
			oldProducerTransportId = tempOldProducerTransportId.String
		}
		if tempOldConsumerTransportId.Valid {
			oldConsumerTransportId = tempOldConsumerTransportId.String
		}
		if tempOldProducerIds.Valid {
			oldProducerIds = UniqueStrings(strings.Split(tempOldProducerIds.String, ","))
		}
		if tempOldConsumerIds.Valid {
			oldConsumerIds = UniqueStrings(strings.Split(tempOldConsumerIds.String, ","))
		}
	}
	if host == "" || len(routerIds) == 0 {
		log.Errorf("Meeting not found for meetingID %s participantID: %s", meetingId, participantId)
		return "", "", "", "", nil, nil, myerrors.MeetingNotFound
	}
	var choseConsumerRouterId string
	if len(routerIds) == 1 {
		choseConsumerRouterId = routerIds[0]
	} else {
		h := fnv.New32a()
		_, err = h.Write([]byte(participantId))
		if err != nil {
			log.Error("Error while writing to hash ", err)
			return "", "", "", "", nil, nil, err
		}
		index := h.Sum32() % uint32(len(routerIds))
		choseConsumerRouterId = routerIds[index]
	}
	return host, choseConsumerRouterId, oldProducerTransportId, oldConsumerTransportId, oldProducerIds, oldConsumerIds, nil
}

func (t *TenantDefaultClusterHandler) AllocateProducerRouterIdForMeetingAndParticipantId(ctx context.Context, meetingId string, participantId string) (
	string, string, string, string, []string, []string, error) {
	query := "select t1.host AS new_host, t3.router_id AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from container_heartbeat t1 " +
		"LEFT JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"LEFT JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"LEFT JOIN router_meeting t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.meeting_id='" + meetingId + "' AND t4.producer_capacity IS NOT NULL"
	existingProducerTransportQuery := "select NULL AS new_host, NULL AS new_router_id, " +
		"transport_id AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from participant_producer_transport where participant_id='" + participantId + "'"
	existingConsumerTransportQuery := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, transport_id AS old_consumer_transport_id, NULL AS old_producer_ids, NULL AS old_consumer_ids " +
		"from participant_consumer_transport where participant_id='" + participantId + "'"
	existingProducers := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, group_concat(producer_id) AS old_producer_ids,NULL AS old_consumer_ids " +
		"from participant_producers " + "where participant_id='" + participantId + "'"
	existingConsumers := "select NULL AS new_host, NULL AS new_router_id, " +
		"NULL AS old_producer_transport_id, NULL AS old_consumer_transport_id, NULL AS old_producer_ids, group_concat(consumer_id) AS old_consumer_ids " +
		"from participant_consumers " + "where participant_id='" + participantId + "'"
	finalQuery := query + " UNION " + existingProducerTransportQuery + " UNION " +
		existingConsumerTransportQuery + " UNION " + existingProducers + " UNION " + existingConsumers
	rows, err := t.queryDb(ctx, finalQuery)
	defer rows.Close()
	if err != nil {
		log.Error("Error wile executing the query ", err)
		return "", "", "", "", nil, nil, err
	}
	var host string
	var routerId string
	var oldProducerTransportId string
	var oldConsumerTransportId string
	var oldProducerIds []string
	var oldConsumerIds []string
	for rows.Next() {
		var tempHost sql.NullString
		var tempRouterId sql.NullString
		var tempOldProducerTransportId sql.NullString
		var tempOldConsumerTransportId sql.NullString
		var tempOldProducerIds sql.NullString
		var tempOldConsumerIds sql.NullString
		err = rows.Scan(&tempHost, &tempRouterId, &tempOldProducerTransportId, &tempOldConsumerTransportId, &tempOldProducerIds, &tempOldConsumerIds)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", "", nil, nil, err
		}
		if tempHost.Valid {
			host = tempHost.String
		}
		if tempRouterId.Valid {
			routerId = tempRouterId.String
		}
		if tempOldProducerTransportId.Valid {
			oldProducerTransportId = tempOldProducerTransportId.String
		}
		if tempOldConsumerTransportId.Valid {
			oldConsumerTransportId = tempOldConsumerTransportId.String
		}
		if tempOldProducerIds.Valid {
			oldProducerIds = strings.Split(tempOldProducerIds.String, ",")
		}
		if tempOldConsumerIds.Valid {
			oldConsumerIds = strings.Split(tempOldConsumerIds.String, ",")
		}
	}
	if host == "" || routerId == "" {
		log.Errorf("Meeting not found %s", meetingId)
		return "", "", "", "", nil, nil, myerrors.MeetingNotFound
	}
	return host, routerId, oldProducerTransportId, oldConsumerTransportId, oldProducerIds, oldConsumerIds, nil
}

func (t *TenantDefaultClusterHandler) GetContainerIPProducerConsumerRouterIdForParticipantId(ctx context.Context, participantId string) (string, string, string, error) {
	producerQuery := "select t1.host, t4.router_id as producer_router_id, 'NA' as consumer_router_id " +
		"from container_heartbeat t1 " +
		"JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"RIGHT JOIN participant_producer_router t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.participant_id='" + participantId + "'"
	consumerQuery := "select t1.host, 'NA' as producer_router_id, t4.router_id as consumer_router_id " +
		"from container_heartbeat t1 " +
		"JOIN worker_container t2 ON t1.container_id = t2.container_id " +
		"JOIN router_worker t3 ON t2.worker_id = t3.worker_id " +
		"RIGHT JOIN participant_consumer_router t4 ON t3.router_id = t4.router_id " +
		"WHERE t4.participant_id='" + participantId + "'"
	query := producerQuery + " UNION " + consumerQuery
	rows, err := t.queryDb(ctx, query)
	if err != nil {
		log.Error("Error in querying Db", err)
		return "", "", "", err
	}
	defer rows.Close()
	var producerHost string
	var producerRouterId string
	var consumerRouterId string
	var consumerHost string
	for rows.Next() {
		var tempHost string
		var tempProducerRouterId string
		var tempConsumerRouterId string
		err = rows.Scan(&tempHost, &tempProducerRouterId, &tempConsumerRouterId)
		if err != nil {
			log.Error("Error while scanning rows", err)
			return "", "", "", err
		}
		if tempProducerRouterId != "NA" {
			producerHost = tempHost
			producerRouterId = tempProducerRouterId
		} else if tempConsumerRouterId != "NA" {
			consumerHost = tempHost
			consumerRouterId = tempConsumerRouterId
		}
	}
	if producerHost == "" || consumerHost == "" || producerHost != consumerHost || producerRouterId == "" || consumerRouterId == "" {
		log.Errorf("Incorrect config for participantId %s", participantId)
		return "", "", "", myerrors.RouterNotFound
	}
	return producerHost, producerRouterId, consumerRouterId, nil
}

func (t *TenantDefaultClusterHandler) GetCalculatedMeetingCapacity(capacity int, producerCapacity int, consumerCapacity int) (int, int, int, error) {
	if capacity > 0 {
		log.Error("Capacity is not supported")
		return 0, 0, 0, myerrors.InValidCapacity
	}
	return 0, producerCapacity * 2, 2 * producerCapacity * (producerCapacity + consumerCapacity), nil
}
