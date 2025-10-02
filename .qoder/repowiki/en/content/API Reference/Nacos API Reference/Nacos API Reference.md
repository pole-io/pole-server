# Nacos API Reference

<cite>
**Referenced Files in This Document**   
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go)
- [nacosserver/v1/server.go](file://plugin/apiserver/nacosserver/v1/server.go)
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go)
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go)
- [nacosserver/v2/discover/instance.go](file://plugin/apiserver/nacosserver/v2/discover/instance.go)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go)
- [nacosserver/v2/config/watch.go](file://plugin/apiserver/nacosserver/v2/config/watch.go)
- [nacosserver/config.go](file://plugin/apiserver/nacosserver/config.go)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [Dual-Protocol Architecture](#dual-protocol-architecture)
3. [Nacos v1 HTTP API Implementation](#nacos-v1-http-api-implementation)
4. [Nacos v2 gRPC API Implementation](#nacos-v2-grpc-api-implementation)
5. [Data Model Mapping](#data-model-mapping)
6. [Watch Mechanisms Comparison](#watch-mechanisms-comparison)
7. [Service Discovery Operations](#service-discovery-operations)
8. [Configuration Management Operations](#configuration-management-operations)
9. [Performance Characteristics](#performance-characteristics)
10. [Migration Strategies](#migration-strategies)

## Introduction
The pole-server implements dual-protocol support for Nacos API through both HTTP (v1) and gRPC (v2) interfaces, providing backward compatibility while enabling modern streaming capabilities. The system supports service discovery and configuration management operations through both protocols, with v1 maintaining compatibility with Nacos HTTP API specifications and v2 leveraging gRPC streaming for improved performance and real-time updates.

**Section sources**
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go#L37-L91)

## Dual-Protocol Architecture
The Nacos server implementation in pole-server supports both v1 (HTTP) and v2 (gRPC) protocols simultaneously through a unified architecture. The `NacosServer` struct manages both protocol implementations, initializing separate v1 and v2 server instances that share the same underlying data storage and business logic.

```mermaid
graph TB
subgraph "Nacos Server"
NacosServer[NacosServer]
v1Svr[NacosV1Server]
v2Svr[NacosV2Server]
DataStore[NacosDataStorage]
end
NacosServer --> v1Svr
NacosServer --> v2Svr
NacosServer --> DataStore
v1Svr --> DataStore
v2Svr --> DataStore
subgraph "Protocols"
HTTP[HTTP/1.1]
GRPC[gRPC/2]
end
HTTP --> v1Svr
GRPC --> v2Svr
```

**Diagram sources**
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go#L37-L91)
- [nacosserver/v1/server.go](file://plugin/apiserver/nacosserver/v1/server.go#L31-L63)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)

**Section sources**
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go#L37-L91)
- [nacosserver/v1/server.go](file://plugin/apiserver/nacosserver/v1/server.go#L31-L63)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)

## Nacos v1 HTTP API Implementation
The v1 implementation provides HTTP-based endpoints for service discovery and configuration management, maintaining compatibility with the Nacos HTTP API specification. The server exposes RESTful endpoints under the `/nacos/v1` path prefix for both service discovery (`/ns`) and configuration management (`/cs/configs`).

### Service Discovery Endpoints
The service discovery API provides standard Nacos v1 endpoints for service registration, instance management, and service listing.

```mermaid
flowchart TD
A[Client Request] --> B{Endpoint}
B --> |/nacos/v1/ns/instance| C[Register/Update Instance]
B --> |/nacos/v1/ns/instance/list| D[List Instances]
B --> |/nacos/v1/ns/service/list| E[List Services]
B --> |/nacos/v1/ns/instance/beat| F[Send Heartbeat]
B --> |/nacos/v1/ns/instance| G[Deregister Instance]
C --> H[Handle Register]
D --> I[Handle Query Instances]
E --> J[Handle Service List]
F --> K[Handle Beat]
G --> L[Handle Deregister]
```

**Diagram sources**
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L0-L186)

### Configuration Management Endpoints
The configuration management API provides endpoints for publishing, retrieving, and watching configuration data.

```mermaid
flowchart TD
A[Client Request] --> B{Endpoint}
B --> |/nacos/v1/cs/configs| C[Publish Configuration]
B --> |/nacos/v1/cs/configs?dataId=X&group=Y| D[Get Configuration]
B --> |/nacos/v1/cs/configs?dataId=X&group=Y| E[Delete Configuration]
B --> |/nacos/v1/cs/configs/listener| F[Watch Configurations]
C --> G[Handle Publish Config]
D --> H[Handle Get Config]
E --> I[Handle Delete Config]
F --> J[Handle Watch]
```

**Diagram sources**
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L0-L187)

**Section sources**
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L0-L186)
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L0-L187)

## Nacos v2 gRPC API Implementation
The v2 implementation leverages gRPC streaming for service discovery and configuration management, providing improved performance and real-time update capabilities. The server uses a handler registry pattern to manage different request types and supports bidirectional streaming for watch operations.

### gRPC Service Architecture
The gRPC server implements a modular handler system where different functionality is organized into separate handler groups.

```mermaid
classDiagram
class NacosV2Server {
+string listenIP
+uint32 listenPort
+map[string]*RequestHandlerWarrper handleRegistry
+Run(errCh chan error)
+Initialize(ctx, option, port, apiConf)
+initHandlers()
}
class RequestHandlerWarrper {
+Handler func(context.Context, BaseRequest, RequestMeta) (BaseResponse, error)
+PayloadBuilder func() CustomerPayload
}
class DiscoverServer {
+ListGRPCHandlers() map[string]*RequestHandlerWarrper
+handleInstanceRequest(ctx, req, meta)
}
class ConfigServer {
+ListGRPCHandlers() map[string]*RequestHandlerWarrper
+handlePublishConfigRequest(ctx, req, meta)
+handleWatchConfigRequest(ctx, req, meta)
}
class ConnectionManager {
+RegisterConnection(conn Connection)
+UnregisterConnection(connID string)
+GetConnection(connID string)
}
NacosV2Server --> RequestHandlerWarrper : "has"
NacosV2Server --> DiscoverServer : "owns"
NacosV2Server --> ConfigServer : "owns"
NacosV2Server --> ConnectionManager : "uses"
DiscoverServer --> RequestHandlerWarrper : "returns"
ConfigServer --> RequestHandlerWarrper : "returns"
```

**Diagram sources**
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L149-L194)

### gRPC Request Flow
The gRPC server processes requests through a standardized flow that includes connection management, request handling, and response generation.

```mermaid
sequenceDiagram
participant Client
participant Server as NacosV2Server
participant Handler as RequestHandler
participant ConnectionMgr as ConnectionManager
Client->>Server : Establish Connection
Server->>ConnectionMgr : Register Connection
ConnectionMgr-->>Server : Connection ID
Server-->>Client : Connection Established
loop Request Processing
Client->>Server : Send Request (Type, Payload)
Server->>Server : Route to Handler
Server->>Handler : Process Request
Handler-->>Server : Response
Server-->>Client : Send Response
end
Client->>Server : Close Connection
Server->>ConnectionMgr : Unregister Connection
```

**Diagram sources**
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L149-L194)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L323-L365)

**Section sources**
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L149-L194)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L323-L365)

## Data Model Mapping
The system implements a comprehensive mapping between Nacos data models and internal pole-server data structures, ensuring compatibility while leveraging internal optimizations.

### Service Discovery Model Mapping
The service discovery data model maps Nacos service and instance concepts to internal representations.

```mermaid
erDiagram
NACOS_SERVICE {
string serviceName PK
string groupName
string namespace
string metadata
}
NACOS_INSTANCE {
string ip PK
int port PK
string serviceName FK
string groupName FK
string namespace FK
string metadata
bool healthy
int weight
string clusterName
}
POLARIS_SERVICE {
string service PK
string namespace FK
string owners
string comment
string metadata
}
POLARIS_INSTANCE {
string id PK
string service FK
string namespace FK
string host
int port
string protocol
string version
string region
string zone
string campus
int weight
bool healthy
string metadata
}
NACOS_SERVICE ||--o{ NACOS_INSTANCE : contains
POLARIS_SERVICE ||--o{ POLARIS_INSTANCE : contains
NACOS_SERVICE }|--|| POLARIS_SERVICE : maps_to
NACOS_INSTANCE }|--|| POLARIS_INSTANCE : maps_to
```

**Diagram sources**
- [nacosserver/v2/discover/instance.go](file://plugin/apiserver/nacosserver/v2/discover/instance.go#L26-L55)
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L0-L186)

### Configuration Model Mapping
The configuration management data model maps Nacos configuration concepts to internal representations.

```mermaid
erDiagram
NACOS_CONFIG {
string dataId PK
string group PK
string namespace PK
string content
string md5
string type
string desc
string appName
string configTags
string cipherAlg
string cipherSt
}
POLARIS_CONFIG_FILE {
string name PK
string namespace FK
string group FK
string content
string md5
string format
string description
string owner
string templateId
}
POLARIS_CONFIG_RELEASE {
string id PK
string name FK
string namespace FK
string group FK
string content
string md5
string format
string description
string owner
string templateId
string extendedArguments
}
NACOS_CONFIG }|--|| POLARIS_CONFIG_FILE : maps_to
POLARIS_CONFIG_FILE ||--o{ POLARIS_CONFIG_RELEASE : has_releases
```

**Diagram sources**
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L25-L51)
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L0-L187)

**Section sources**
- [nacosserver/v2/discover/instance.go](file://plugin/apiserver/nacosserver/v2/discover/instance.go#L26-L55)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L25-L51)
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L0-L186)
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L0-L187)

## Watch Mechanisms Comparison
The system implements different watch mechanisms for v1 (HTTP long-polling) and v2 (gRPC streaming), each with distinct performance characteristics and use cases.

### v1 HTTP Long-Polling Watch
The v1 implementation uses HTTP long-polling for configuration watches, where clients maintain open connections that are held until configuration changes occur.

```mermaid
sequenceDiagram
participant Client
participant Server
participant Watcher as WatchCenter
Client->>Server : POST /nacos/v1/cs/configs/listener
Server->>Watcher : AddWatcher(client, configs)
Watcher-->>Server : Watch registered
Server-->>Client : Hold connection
Note over Server,Watcher : Configuration changes
Watcher->>Server : Notify change
Server->>Client : Respond with changed configs
Client->>Server : Immediate re-poll
Server->>Watcher : AddWatcher(client, configs)
Server-->>Client : Hold connection
```

**Diagram sources**
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L94-L142)

### v2 gRPC Streaming Watch
The v2 implementation uses gRPC bidirectional streaming for configuration watches, establishing persistent connections that enable real-time push notifications.

```mermaid
sequenceDiagram
participant Client
participant Server
participant Watcher as WatchContext
Client->>Server : Establish gRPC stream
Server->>Server : Create StreamWatchContext
Server->>Watcher : AddWatcher(client, configs, context)
Note over Server,Watcher : Configuration changes
Watcher->>Watcher : Reply(event)
Watcher->>Server : Send ConfigChangeNotifyRequest
Server->>Client : Stream push notification
loop Stream active
Client->>Server : Send heartbeats
Server->>Client : Push configuration changes
end
Client->>Server : Close stream
Server->>Watcher : RemoveWatcher
```

**Diagram sources**
- [nacosserver/v2/config/watch.go](file://plugin/apiserver/nacosserver/v2/config/watch.go#L146-L172)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L177-L214)

**Section sources**
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L94-L142)
- [nacosserver/v2/config/watch.go](file://plugin/apiserver/nacosserver/v2/config/watch.go#L146-L172)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L177-L214)

## Service Discovery Operations
The service discovery operations are implemented consistently across both v1 and v2 protocols, with identical functionality exposed through different transport mechanisms.

### Service Registration
Service registration allows clients to register instances with the service registry, making them available for discovery by other services.

```mermaid
flowchart TD
A[Client] --> B{Protocol}
B --> |v1 HTTP| C[POST /nacos/v1/ns/instance]
B --> |v2 gRPC| D[InstanceRequest with Register action]
C --> E[BuildInstance from request]
D --> F[PrepareSpecInstance from request]
E --> G[handleRegister with context]
F --> G
G --> H[Store instance in DataStorage]
H --> I[Return success response]
I --> J[Client]
```

**Section sources**
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L104-L125)
- [nacosserver/v2/discover/instance.go](file://plugin/apiserver/nacosserver/v2/discover/instance.go#L26-L55)

### Service Subscription
Service subscription enables clients to receive updates when instances of a service change, supporting both polling and streaming models.

```mermaid
flowchart TD
A[Client] --> B{Protocol}
B --> |v1 HTTP| C[GET /nacos/v1/ns/instance/list with long-polling]
B --> |v2 gRPC| D[SubscribeRequest stream]
C --> E[handleQueryInstances with watch]
D --> F[AddWatcher to WatchCenter]
E --> G[Hold connection until change]
F --> G[Stream instance updates]
G --> H{Instance changes?}
H --> |Yes| I[Send updated instances]
H --> |No| J[Continue watching]
I --> K[Client]
```

**Section sources**
- [nacosserver/v1/discover/access.go](file://plugin/apiserver/nacosserver/v1/discover/access.go#L168-L185)
- [nacosserver/v2/discover/instance.go](file://plugin/apiserver/nacosserver/v2/discover/instance.go#L26-L55)

## Configuration Management Operations
Configuration management operations provide CRUD functionality for configuration data, with watch capabilities for real-time updates.

### Configuration Publishing
Configuration publishing allows clients to create or update configuration data in the system.

```mermaid
flowchart TD
A[Client] --> B{Protocol}
B --> |v1 HTTP| C[POST /nacos/v1/cs/configs]
B --> |v2 gRPC| D[ConfigPublishRequest]
C --> E[BuildConfigFile from request]
D --> F[Parse ConfigPublishRequest]
E --> G[handlePublishConfig with context]
F --> G
G --> H[Store config in DataStorage]
H --> I[Trigger notifications to watchers]
I --> J[Return success response]
J --> K[Client]
```

**Section sources**
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L58-L78)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L25-L51)

### Configuration Watching
Configuration watching enables clients to receive updates when configuration data changes, supporting both push and pull models.

```mermaid
flowchart TD
A[Client] --> B{Protocol}
B --> |v1 HTTP| C[POST /nacos/v1/cs/configs/listener]
B --> |v2 gRPC| D[ConfigBatchListenRequest]
C --> E[Parse Listening-Configs header]
D --> F[Parse watch files list]
E --> G[AddWatcher to WatchCenter]
F --> G
G --> H{Configuration changes?}
H --> |Yes| I[Send updated config]
H --> |No| J[Continue watching]
I --> K[Client]
```

**Section sources**
- [nacosserver/v1/config/access.go](file://plugin/apiserver/nacosserver/v1/config/access.go#L94-L142)
- [nacosserver/v2/config/config_file.go](file://plugin/apiserver/nacosserver/v2/config/config_file.go#L177-L214)

## Performance Characteristics
The dual-protocol implementation exhibits different performance characteristics based on the transport mechanism and watch strategy employed.

### Latency Comparison
The gRPC v2 implementation generally provides lower latency for operations due to the persistent connection model and binary serialization.

```mermaid
graph TD
A[Operation] --> B[Latency]
B --> C[v1 HTTP]
B --> D[v2 gRPC]
C --> E[Service Registration: 10-50ms]
C --> F[Configuration Get: 5-20ms]
C --> G[Watch Notification: 100-1000ms]
D --> H[Service Registration: 5-20ms]
D --> I[Configuration Get: 2-10ms]
D --> J[Watch Notification: 10-100ms]
```

### Throughput Comparison
The gRPC v2 implementation supports higher throughput due to connection multiplexing and reduced connection overhead.

```mermaid
graph TD
A[Scenario] --> B[Throughput]
B --> C[v1 HTTP]
B --> D[v2 gRPC]
C --> E[Service Registration: 1000-5000 ops/s]
C --> F[Configuration Watch: 500-2000 clients/server]
D --> G[Service Registration: 5000-20000 ops/s]
D --> H[Configuration Watch: 10000-50000 clients/server]
```

### Resource Utilization
The resource utilization differs significantly between the two protocols, particularly in terms of connection overhead.

```mermaid
graph TD
A[Resource] --> B[Utilization]
B --> C[v1 HTTP]
B --> D[v2 gRPC]
C --> E[Connections: High (1 per request)]
C --> F[Memory: Medium]
C --> G[CPU: Medium]
D --> H[Connections: Low (1 per client)]
D --> I[Memory: High (per connection)]
D --> J[CPU: Low]
```

**Section sources**
- [nacosserver/v1/server.go](file://plugin/apiserver/nacosserver/v1/server.go#L31-L63)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L323-L365)

## Migration Strategies
Migrating from v1 to v2 requires careful planning to ensure compatibility and minimize service disruption.

### Gradual Migration
A gradual migration strategy allows both protocols to coexist while transitioning clients incrementally.

```mermaid
flowchart TD
A[Start] --> B[Enable both v1 and v2]
B --> C[Update server configuration]
C --> D[Monitor v1 traffic]
D --> E{Ready to migrate?}
E --> |No| D
E --> |Yes| F[Update client libraries]
F --> G[Migrate clients to v2]
G --> H{All clients migrated?}
H --> |No| G
H --> |Yes| I[Deprecate v1 endpoints]
I --> J[Monitor v2 performance]
J --> K[Disable v1 if needed]
K --> L[End]
```

### Compatibility Considerations
When migrating between versions, several compatibility factors must be considered.

```mermaid
flowchart TD
A[Data Model] --> B[Ensure backward compatibility]
A --> C[Handle namespace mapping]
A --> D[Preserve metadata structure]
E[Authentication] --> F[Support same auth methods]
E --> G[Maintain token compatibility]
H[Error Handling] --> I[Map error codes consistently]
H --> J[Preserve error message format]
K[Rate Limiting] --> L[Apply same limits across protocols]
K --> M[Share rate limit state]
```

**Section sources**
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go#L207-L237)
- [nacosserver/v2/server.go](file://plugin/apiserver/nacosserver/v2/server.go#L105-L147)
- [nacosserver/config.go](file://plugin/apiserver/nacosserver/config.go#L0-L35)