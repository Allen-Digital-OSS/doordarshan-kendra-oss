## 🏗 High Level Architecture

```mermaid
flowchart LR

    subgraph Client Layer
        C[Thin Client]
    end

    subgraph Control Plane
        K[Doordarshan Kendra]
        DB[(Database)]
    end

    subgraph Media Plane
        MS[Media Server Instance]
    end

    C -->|API Calls| K

    K -->|Read and get the meta data of the media-server cluster| DB
    K -->|Allocate| MS


    MS -->|Write Runtime State| DB
```

## Meeting Creation Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Kendra
    participant DB
    participant MediaServer

    Client->>Kendra: Create Meeting(meetingId, tenant)

    Kendra->>DB: Fetch available media servers
    DB-->>Kendra: Server list + capacity metrics

    Kendra->>Kendra: Select server based on utilization

    Kendra->>MediaServer: Create Meeting(meetingId)

    MediaServer->>MediaServer: Allocate worker
    MediaServer->>MediaServer: Create router

    MediaServer->>DB: Write allocation metadata
    DB-->>MediaServer: Ack

    MediaServer-->>Kendra: Meeting Created (routerId)

    Kendra-->>Client: Meeting Created
```

## Join Meeting Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Kendra
    participant DB
    participant MediaServer

    Client->>Kendra: Join Meeting

    Kendra->>DB: Lookup meeting → mediaServer
    DB-->>Kendra: mediaServerId + routerId

    Kendra->>MediaServer: Fetch rtpCapabilities(routerId)

    MediaServer-->>Kendra: rtpCapabilities

    Kendra-->>Client: rtpCapabilities
```

## Recreate Producer / Consumer Transport Recreation (With Old Transport Cleanup)

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Kendra
    participant DB
    participant MediaServer

    Client->>Kendra: Recreate Transport(meetingId, participantId, instanceId)

    %% Step 1: Resolve media server
    Kendra->>DB: Lookup meeting → mediaServer
    DB-->>Kendra: mediaServerId

    %% Step 2: Fetch existing transports
    Kendra->>DB: Fetch existing transportIds(participantId)
    DB-->>Kendra: [oldTransportIds]

    %% Step 3: Request recreation
    Kendra->>MediaServer: Recreate Transport(oldTransportIds, participantId, instanceId)

    %% Step 4: Media server cleanup
    MediaServer->>MediaServer: Destroy old transports (if any)
    MediaServer->>MediaServer: Cleanup related producers / consumers

    %% Step 5: Create new transport
    MediaServer->>MediaServer: Create new WebRTC Transport

    %% Step 6: Persist new transport
    MediaServer->>DB: Delete old transportIds and insert new transportId
    DB-->>MediaServer: Ack

    %% Step 7: Return transport params
    MediaServer-->>Kendra: Transport Params (id, ice, dtls)
    Kendra-->>Client: Transport Params
```

## Create Consumer Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Kendra
    participant MediaServer
    participant DB

    Client->>Kendra: Create Consumer(meetingId, participantId)

    Kendra->>Kendra: Resolve assigned media-server

    Kendra->>MediaServer: Create Consumer

    MediaServer->>MediaServer: Create Consumer internally

    MediaServer->>DB: Persist consumer metadata (participantId, consumerId, producerId)
    DB-->>MediaServer: Ack

    MediaServer-->>Kendra: Success + metadata

    Kendra-->>Client: Success
```


## Create Producer Flow (With Redis Notification)

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant Kendra
    participant MediaServer
    participant DB
    participant Redis

    Client->>Kendra: Create Producer(meetingId, participantId)

    Kendra->>Kendra: Resolve assigned media-server

    Kendra->>MediaServer: Create Producer

    MediaServer->>MediaServer: Create Producer internally

    MediaServer->>DB: Persist producer metadata (participantId, producerId)
    DB-->>MediaServer: Ack

    MediaServer-->>Kendra: Producer Created (producerId)
    
    Kendra->>Redis: Push ProducerCreated Event

    Kendra-->>Client: Success

```

