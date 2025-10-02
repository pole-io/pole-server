# Service Discovery

<cite>
**Referenced Files in This Document**   
- [nacos_grpc_service.proto](file://plugin/apiserver/nacosserver/v2/pb/nacos_grpc_service.proto)
- [server.go](file://plugin/apiserver/nacosserver/v2/server.go)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go)
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go)
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [gRPC Service Definition](#grpc-service-definition)
3. [Bidirectional Streaming Model](#bidirectional-streaming-model)
4. [Discovery Stream Lifecycle](#discovery-stream-lifecycle)
5. [Subscription Management](#subscription-management)
6. [Push-Based Notification System](#push-based-notification-system)
7. [Client State and Inflight Management](#client-state-and-inflight-management)
8. [Code Examples](#code-examples)
9. [Performance Considerations](#performance-considerations)
10. [Troubleshooting Guide](#troubleshooting-guide)

## Introduction
The Nacos v2 gRPC Service Discovery implementation in pole-server provides a real-time, bidirectional streaming interface for service discovery operations. This system enables clients to subscribe to service instances and receive immediate push updates when changes occur, eliminating the need for polling. The implementation follows the Nacos v2 gRPC protocol specification and integrates with pole-server's internal service registry and event distribution system.

**Section sources**
- [server.go](file://plugin/apiserver/nacosserver/v2/server.go#L53-L108)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)

## gRPC Service Definition
The service discovery functionality is defined in the `nacos_grpc_service.proto` file, which specifies the bidirectional streaming interface for service discovery operations. The core service is `BiRequestStream` which enables full-duplex communication between clients and server.

```mermaid
classDiagram
class Payload {
+Metadata metadata
+google.protobuf.Any body
}
class Metadata {
+string type
+string clientIp
+map~string, string~ headers
}
class BiRequestStream {
+requestBiStream(stream Payload) returns (stream Payload)
}
Payload "1" *-- "1" Metadata
BiRequestStream ..> Payload : uses
```

**Diagram sources**
- [nacos_grpc_service.proto](file://plugin/apiserver/nacosserver/v2/pb/nacos_grpc_service.proto#L0-L42)

The service supports various request types identified by the `type` field in the Metadata, including:
- `SubscribeServiceRequest`: For subscribing to service instance changes
- `InstanceRequest`: For querying service instances
- `BatchInstanceRequest`: For batch querying of service instances
- `ConnectionSetupRequest`: For establishing the initial connection

**Section sources**
- [nacos_grpc_service.proto](file://plugin/apiserver/nacosserver/v2/pb/nacos_grpc_service.proto#L0-L42)

## Bidirectional Streaming Model
The bidirectional streaming model allows clients to establish a persistent connection with the server and receive real-time updates about service instance changes. This model provides several advantages over traditional request-response patterns:

- **Real-time updates**: Clients receive instance changes immediately via push notifications
- **Reduced latency**: Eliminates polling intervals and associated delays
- **Efficient resource usage**: Single connection can handle multiple subscriptions
- **Stateful communication**: Server maintains client context and subscription state

The streaming model is implemented using gRPC's bidirectional streaming capability, where both client and server can send multiple messages in any order over a single connection.

```mermaid
sequenceDiagram
participant Client
participant Server
Client->>Server : Establish Stream Connection
Server-->>Client : Connection Established
Client->>Server : Send SubscribeServiceRequest
Server->>Server : Validate Subscription
Server-->>Client : Initial Instance Sync
Server->>Client : Push Instance Updates (delta)
Client->>Server : Send Keepalive/Heartbeat
Server->>Client : Process ConnectionSetupRequest
Client->>Server : Close Stream
Server-->>Client : Stream Closed
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L145)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L145)

## Discovery Stream Lifecycle
The lifecycle of a discovery stream consists of several phases: connection establishment, subscription validation, initial synchronization, delta updates, and disconnection handling.

### Connection Establishment
When a client establishes a bidirectional stream, the server registers the connection in the ConnectionManager and sets up the stream reference for the client.

### Initial Synchronization
After a successful subscription request, the server performs an initial synchronization by sending the complete set of current instances for the requested service.

### Delta Updates via grpc_push
The server uses the grpc_push system to deliver incremental updates to subscribed clients whenever service instances change. This ensures clients receive only the changes (deltas) rather than the entire dataset.

### Disconnection Handling
The system properly handles both graceful and abrupt disconnections. When a stream is closed, the server cleans up the client's subscriptions and removes it from the active connections registry.

```mermaid
flowchart TD
A[Client Connects] --> B{Connection Valid?}
B --> |No| C[Reject Connection]
B --> |Yes| D[Register Connection]
D --> E[Client Sends Subscribe Request]
E --> F{Subscription Valid?}
F --> |No| G[Send Error Response]
F --> |Yes| H[Store Subscription]
H --> I[Send Initial Instance Data]
I --> J[Monitor for Instance Changes]
J --> K{Instance Changed?}
K --> |Yes| L[Push Delta Update]
L --> J
K --> |No| M{Stream Active?}
M --> |Yes| J
M --> |No| N[Clean Up Resources]
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)

## Subscription Management
The system validates subscriptions through a multi-step process that ensures clients have proper authorization and the requested services exist in the registry.

### Subscription Validation
When a client sends a SubscribeServiceRequest, the server validates:
- Client authentication and authorization
- Service namespace existence
- Service name validity
- Client IP and metadata

### Inflight Management
The ConnectionManager maintains client state through inflight tracking, which allows the server to:
- Track active subscriptions per client
- Manage request-response correlation
- Handle acknowledgments for push notifications
- Prevent duplicate processing of requests

```mermaid
classDiagram
class DiscoverServer {
+*ConnectionClientManager clientManager
+*ConnectionManager connMgr
+PushCenter pushCenter
+*NacosDataStorage store
+map~string, *RequestHandlerWarrper~ handleRegistry
+*Checker checker
}
class ConnectionManager {
+GetClient(connID string) (*Client, bool)
+RegisterConnection(ctx context.Context, payload *Payload, msg *ConnectionSetupRequest) error
+RefreshClient(ctx context.Context)
+InFlights() *InFlightManager
}
class ConnectionClientManager {
+AddClient(client *Client) error
+RemoveClient(client *Client)
+GetClient(connID string) (*Client, bool)
}
class Client {
+SetStreamRef(stream *SyncServerStream)
+GetStreamRef() *SyncServerStream
+Key string
}
class InFlightManager {
+NotifyInFlight(connID string, response BaseResponse)
}
DiscoverServer --> ConnectionManager : uses
DiscoverServer --> ConnectionClientManager : uses
ConnectionManager --> Client : manages
ConnectionManager --> InFlightManager : delegates
```

**Diagram sources**
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)

**Section sources**
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)

## Push-Based Notification System
The push-based notification system is the core mechanism for delivering real-time service instance updates to subscribed clients.

### grpc_push Implementation
The GrpcPushCenter implements the push center interface and manages subscribers for gRPC-based push notifications. It integrates with the eventhub system to receive service instance change events.

### Push Data Flow
When a service instance changes, the following sequence occurs:
1. The change is detected by the system
2. An event is published to the eventhub
3. The GrpcPushCenter receives the event
4. The push center identifies all subscribers to the affected service
5. Delta updates are sent to each subscribed client

```mermaid
sequenceDiagram
participant Service as Service Registry
participant EventHub as EventHub
participant PushCenter as GrpcPushCenter
participant Client as Subscriber Client
Service->>EventHub : Publish Instance Change Event
EventHub->>PushCenter : Deliver Event
PushCenter->>PushCenter : Identify Subscribers
loop For each subscriber
PushCenter->>Client : Push Instance Update
Client-->>PushCenter : Acknowledge Receipt
end
PushCenter->>PushCenter : Update Last Refresh Time
```

**Diagram sources**
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L145)

**Section sources**
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)

## Client State and Inflight Management
The system maintains detailed client state to ensure reliable message delivery and proper handling of network disruptions.

### Connection State Management
The ConnectionManager tracks each client's connection state, including:
- Connection ID and metadata
- Stream reference for push operations
- Last refresh time for keepalive detection
- Active subscriptions

### Inflight Request Tracking
The inflight system tracks pending requests and responses, allowing the server to:
- Correlate requests with their responses
- Handle acknowledgments for push notifications
- Manage timeout and retry logic
- Prevent message loss during network interruptions

```mermaid
classDiagram
class ConnectionManager {
+map~string, *Client~ clients
+*InFlightManager inflights
+RegisterConnection(ctx context.Context, payload *Payload, msg *ConnectionSetupRequest) error
+GetClient(connID string) (*Client, bool)
+RefreshClient(ctx context.Context)
}
class Client {
+string connID
+*SyncServerStream streamRef
+int64 lastRefreshTime
+map~string, *Subscription~ subscriptions
+SetStreamRef(stream *SyncServerStream)
+GetStreamRef() *SyncServerStream
}
class InFlightManager {
+map~string, chan BaseResponse~ pendingResponses
+NotifyInFlight(connID string, response BaseResponse)
+WaitForResponse(requestID string) BaseResponse
+RemoveRequest(requestID string)
}
class Subscription {
+string serviceKey
+string namespace
+int64 createTime
}
ConnectionManager "1" *-- "0..*" Client : manages
ConnectionManager --> InFlightManager : delegates
Client "1" *-- "0..*" Subscription : has
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)

## Code Examples
The following examples demonstrate how to interact with the Nacos v2 gRPC service discovery system.

### Subscribing to a Service
Clients can subscribe to service instance changes by sending a SubscribeServiceRequest over the bidirectional stream.

### Handling Instance Push Events
Clients receive instance updates through the stream and should implement proper handling for both full syncs and delta updates.

### Managing Reconnections
Clients should implement reconnection logic to handle network failures and maintain subscription continuity.

```mermaid
flowchart TD
A[Create gRPC Connection] --> B[Establish Bidirectional Stream]
B --> C[Send ConnectionSetupRequest]
C --> D[Send SubscribeServiceRequest]
D --> E{Receive Response?}
E --> |Yes| F[Process Initial Instance Data]
E --> |No| G[Handle Error]
F --> H[Listen for Push Updates]
H --> I{Update Received?}
I --> |Yes| J[Process Instance Changes]
J --> H
I --> |Connection Lost| K[Attempt Reconnection]
K --> L{Reconnection Successful?}
L --> |Yes| M[Resubscribe to Services]
L --> |No| N[Exponential Backoff]
N --> K
M --> H
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L145)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L145)

## Performance Considerations
The system incorporates several performance optimizations to handle high-volume service discovery traffic efficiently.

### Stream Multiplexing
A single bidirectional stream can carry multiple subscription requests and responses, reducing the overhead of maintaining multiple connections.

### Backpressure via Flow Control
The gRPC framework provides built-in flow control mechanisms that prevent overwhelming clients with too many push notifications.

### Efficient Delta Encoding
The system sends only the changes (deltas) to service instances rather than the complete dataset, minimizing network bandwidth usage.

### Connection Keepalive
Regular keepalive messages ensure connection health and allow for timely detection of failed connections.

```mermaid
flowchart LR
A[Stream Multiplexing] --> B[Single Connection<br/>Multiple Subscriptions]
C[Flow Control] --> D[Client-Server<br/>Window Management]
E[Delta Encoding] --> F[Send Only Changes<br/>Not Full Dataset]
G[Keepalive] --> H[Detect Failed<br/>Connections Early]
style A fill:#f9f,stroke:#333
style C fill:#f9f,stroke:#333
style E fill:#f9f,stroke:#333
style G fill:#f9f,stroke:#333
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L134-L205)
- [discover/grpc_push.go](file://plugin/apiserver/nacosserver/v2/discover/grpc_push.go#L0-L97)

## Troubleshooting Guide
This section addresses common issues encountered when using the Nacos v2 gRPC service discovery system.

### Missed Updates
If clients are missing instance updates, check:
- Client connection stability
- Network connectivity between client and server
- Server-side push center functionality
- Client processing logic for incoming messages

### Stream Timeouts
Stream timeouts can occur due to:
- Network interruptions
- Client processing delays
- Server resource constraints
- Firewall or proxy interference

### Authentication Failures
Authentication failures may result from:
- Invalid or expired tokens
- Incorrect client metadata
- Authorization policy changes
- Network proxy altering headers

```mermaid
flowchart TD
A[Issue: Missed Updates] --> B[Check Client Connection]
B --> C[Verify Network Stability]
C --> D[Inspect Server Logs]
D --> E[Review Client Processing Logic]
F[Issue: Stream Timeout] --> G[Check Network Connectivity]
G --> H[Monitor Client Processing Time]
H --> I[Review Server Resources]
I --> J[Inspect Firewall/Proxy Settings]
K[Issue: Authentication Failure] --> L[Validate Token Validity]
L --> M[Check Metadata Headers]
M --> N[Review Authorization Policies]
N --> O[Test Without Proxy]
```

**Diagram sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)

**Section sources**
- [access.go](file://plugin/apiserver/nacosserver/v2/access.go#L88-L133)
- [discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L42-L55)