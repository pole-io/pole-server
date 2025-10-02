# XDS Server v3 Plugin

<cite>
**Referenced Files in This Document**   
- [server.go](file://plugin/apiserver/xdsserverv3/server.go)
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go)
- [lds.go](file://plugin/apiserver/xdsserverv3/lds.go)
- [rds.go](file://plugin/apiserver/xdsserverv3/rds.go)
- [cache/cache.go](file://plugin/apiserver/xdsserverv3/cache/cache.go)
- [generate.go](file://plugin/apiserver/xdsserverv3/generate.go)
- [cds.go](file://plugin/apiserver/xdsserverv3/cds.go)
- [eds.go](file://plugin/apiserver/xdsserverv3/eds.go)
- [vhds.go](file://plugin/apiserver/xdsserverv3/vhds.go)
- [hds.go](file://plugin/apiserver/xdsserverv3/hds.go)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [Project Structure](#project-structure)
3. [Core Components](#core-components)
4. [Architecture Overview](#architecture-overview)
5. [Detailed Component Analysis](#detailed-component-analysis)
6. [Dependency Analysis](#dependency-analysis)
7. [Performance Considerations](#performance-considerations)
8. [Troubleshooting Guide](#troubleshooting-guide)
9. [Conclusion](#conclusion)

## Introduction
The XDS Server v3 Plugin implements the Envoy xDS v3 APIs to support service mesh sidecar configuration. This document details the implementation of CDS, RDS, LDS, EDS, VHDS, and HDS services, the resource model for service routing, the streaming xDS protocol handling, and cache mechanisms for client subscriptions. The plugin integrates with the Polaris service mesh to deliver dynamic configuration updates for traffic management, security policies, and health discovery.

## Project Structure
The XDS v3 server plugin is located in the `plugin/apiserver/xdsserverv3` directory and implements the full suite of Envoy xDS APIs. The structure follows a modular design with separate files for each xDS resource type and core functionality.

```mermaid
graph TB
subgraph "XDSv3 Plugin"
server[server.go<br/>Main Server]
resource[resource/model.go<br/>Resource Model]
cache[cache/cache.go<br/>Resource Cache]
generate[generate.go<br/>Resource Generator]
lds[lds.go<br/>Listener Config]
rds[rds.go<br/>Route Config]
cds[cds.go<br/>Cluster Config]
eds[eds.go<br/>Endpoint Config]
vhds[vhds.go<br/>Virtual Host]
hds[hds.go<br/>Health Discovery]
debug[debug.go<br/>Debug Handlers]
end
server --> cache
server --> generate
generate --> lds
generate --> rds
generate --> cds
generate --> eds
generate --> vhds
server --> hds
server --> debug
resource --> server
resource --> generate
resource --> lds
resource --> rds
```

**Diagram sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L1-L50)
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go#L1-L50)

**Section sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L1-L50)
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go#L1-L50)

## Core Components
The XDS v3 plugin implements the core Envoy xDS APIs for service mesh configuration. The system is built around a resource model that transforms service routing rules into Envoy-compatible configurations. The streaming xDS protocol handles DiscoveryRequests with versioning and nonce management through the go-control-plane library. The cache mechanism tracks client subscriptions and manages incremental updates for efficient configuration distribution.

**Section sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L41-L76)
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go#L77-L150)

## Architecture Overview
The XDS v3 server architecture follows the Envoy control plane pattern, implementing the Aggregated Discovery Service (ADS) and individual xDS services. The server maintains a resource cache and generates dynamic configurations based on service registry data and traffic management policies.

```mermaid
graph TD
subgraph "Control Plane"
XDSServer[XDSServer]
ResourceCache[ResourceCache]
XdsResourceGenerator[XdsResourceGenerator]
XDSNodeManager[XDSNodeManager]
end
subgraph "Data Sources"
NamingServer[Naming Server<br/>Service Registry]
RuleServer[Rule Server<br/>Traffic Policies]
HealthSvr[Health Server]
end
subgraph "Envoy Proxies"
Envoy1[Envoy Sidecar 1]
Envoy2[Envoy Sidecar 2]
Envoy3[Envoy Gateway]
end
XDSServer --> ResourceCache
XDSServer --> XdsResourceGenerator
XDSServer --> XDSNodeManager
XdsResourceGenerator --> NamingServer
XdsResourceGenerator --> RuleServer
XdsResourceGenerator --> HealthSvr
ResourceCache --> XDSNodeManager
Envoy1 --> XDSServer
Envoy2 --> XDSServer
Envoy3 --> XDSServer
style XDSServer fill:#4CAF50,stroke:#388E3C
style ResourceCache fill:#2196F3,stroke:#1976D2
style XdsResourceGenerator fill:#FF9800,stroke:#F57C00
```

**Diagram sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L41-L76)
- [cache/cache.go](file://plugin/apiserver/xdsserverv3/cache/cache.go#L1-L20)

## Detailed Component Analysis

### Resource Model Implementation
The resource model in `resource/model.go` defines the ServiceInfo structure that represents service instances and their associated policies. This model serves as the foundation for generating Envoy configurations from service mesh data.

```mermaid
classDiagram
class ServiceInfo {
+string ID
+string Name
+string Namespace
+ServiceKey ServiceKey
+*Instance Instances
+string SvcInsRevision
+Routing Routing
+string SvcRoutingRevision
+*ServicePort Ports
+RateLimit RateLimit
+string SvcRateLimitRevision
+CircuitBreaker CircuitBreaker
+string CircuitBreakerRevision
+FaultDetector FaultDetect
+string FaultDetectRevision
+AliasFor *Service
+Equal(ServiceInfo) bool
+MatchService(string, string) bool
}
class XDSType {
+LDS
+RDS
+EDS
+CDS
+RLS
+SDS
+VHDS
+UnknownXDS
+FromSimpleXDS(string) XDSType
+FormatTypeUrl(string) XDSType
+ResourceType() Type
+String() string
}
class TLSMode {
+TLSModeNone
+TLSModeStrict
+TLSModePermissive
+EnableTLS(TLSMode) bool
}
ServiceInfo "1" -- "0..*" Instance : contains
ServiceInfo "1" -- "0..1" Routing : has
ServiceInfo "1" -- "0..1" RateLimit : has
ServiceInfo "1" -- "0..1" CircuitBreaker : has
ServiceInfo "1" -- "0..1" FaultDetector : has
```

**Diagram sources**
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go#L77-L150)

**Section sources**
- [resource/model.go](file://plugin/apiserver/xdsserverv3/resource/model.go#L77-L150)

### Streaming xDS Protocol Implementation
The xDS protocol implementation in `server.go` handles DiscoveryRequests with versioning and nonce management through the go-control-plane library. The server implements the full set of xDS services and manages client connections with proper lifecycle handling.

```mermaid
sequenceDiagram
participant Envoy as "Envoy Proxy"
participant XDSServer as "XDSServer"
participant Cache as "ResourceCache"
participant Generator as "XdsResourceGenerator"
Envoy->>XDSServer : Stream Request (type_url)
XDSServer->>Cache : Register Node
XDSServer->>Cache : Create Watch
loop Periodic Updates
Generator->>Cache : Generate Snapshot
Cache->>XDSServer : Notify Resources
XDSServer->>Envoy : Send Response (version_info, nonce)
Envoy->>XDSServer : ACK (version_info, nonce)
end
Envoy->>XDSServer : Stream Close
XDSServer->>Cache : Cancel Watch
XDSServer->>Cache : Remove Node
Note over XDSServer,Cache : Versioning and nonce management<br/>ensures consistent updates
```

**Diagram sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L166-L207)

**Section sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L17-L41)

### Listener and Route Configuration Generation
The listener and route configuration generation in `lds.go` and `rds.go` transforms service routing rules into Envoy-compatible configurations. The implementation handles both inbound and outbound traffic with appropriate route configuration names.

```mermaid
flowchart TD
Start([Generate Listener]) --> CreateListener["Create Listener Resource"]
CreateListener --> SetName["Set Listener Name"]
SetName --> |Inbound| InboundName["inbound|<port>"]
SetName --> |Outbound| OutboundName["outbound|<port>"]
InboundName --> ConfigureFilterChain["Configure Filter Chain"]
OutboundName --> ConfigureFilterChain
ConfigureFilterChain --> AddHttpConnectionManager["Add HTTP Connection Manager"]
AddHttpConnectionManager --> SetRouteConfig["Set Route Configuration"]
SetRouteConfig --> |Inbound| InboundRoute["polaris-inbound-cluster"]
SetRouteConfig --> |Outbound| OutboundRoute["polaris-outbound-router"]
InboundRoute --> End([Listener Complete])
OutboundRoute --> End
style Start fill:#4CAF50,stroke:#388E3C
style End fill:#4CAF50,stroke:#388E3C
```

**Diagram sources**
- [lds.go](file://plugin/apiserver/xdsserverv3/lds.go#L1-L20)
- [rds.go](file://plugin/apiserver/xdsserverv3/rds.go#L1-L20)

**Section sources**
- [lds.go](file://plugin/apiserver/xdsserverv3/lds.go#L1-L50)
- [rds.go](file://plugin/apiserver/xdsserverv3/rds.go#L1-L50)

### Cache Mechanism for Client Subscriptions
The cache mechanism in `cache/cache.go` tracks client subscriptions and manages incremental updates. The ResourceCache maintains snapshots of configuration state and handles watch registration for efficient change propagation.

```mermaid
classDiagram
class ResourceCache {
+*XDSServer server
+*XDSNodeManager nodeManager
+GenerateSnapshot(ServiceInfos) *Snapshot
+Fetch(context.Context, *Request) (*Response, error)
+StreamHandler(*StreamRequest, Stream)
+DeltaStreamHandler(*DeltaStreamRequest, DeltaStream)
}
class XDSNodeManager {
+map[string]*XDSNode nodes
+RegisterNode(*core.Node) *XDSNode
+GetNode(string) *XDSNode
+RemoveNode(string)
+ListNodes() []*XDSNode
}
class XDSNode {
+string nodeID
+string nodeType
+*core.Node metadata
+map[XDSType]*NodeWatch watches
+AddWatch(XDSType, *NodeWatch)
+RemoveWatch(XDSType)
}
class NodeWatch {
+XDSType type
+string version
+string nonce
+chan *Response responseChan
+Cancel()
}
ResourceCache --> XDSNodeManager : uses
XDSNodeManager --> XDSNode : contains
XDSNode --> NodeWatch : contains
```

**Diagram sources**
- [cache/cache.go](file://plugin/apiserver/xdsserverv3/cache/cache.go#L1-L20)

**Section sources**
- [cache/cache.go](file://plugin/apiserver/xdsserverv3/cache/cache.go#L1-L50)

## Dependency Analysis
The XDS v3 plugin has dependencies on several core components of the Polaris service mesh, including the naming server for service registry data, the rule server for traffic management policies, and the health server for instance health status.

```mermaid
graph TD
XDSServer --> NamingServer
XDSServer --> RuleServer
XDSServer --> HealthSvr
XDSServer --> go-control-plane
NamingServer --> MySQL
RuleServer --> MySQL
HealthSvr --> Redis
style XDSServer fill:#4CAF50,stroke:#388E3C
style NamingServer fill:#2196F3,stroke:#1976D2
style RuleServer fill:#2196F3,stroke:#1976D2
style HealthSvr fill:#2196F3,stroke:#1976D2
style go-control-plane fill:#FF9800,stroke:#F57C00
click XDSServer "https://github.com/pole-io/pole-server" "XDS Server v3 Plugin"
click NamingServer "https://github.com/pole-io/pole-server" "Naming Server"
click RuleServer "https://github.com/pole-io/pole-server" "Rule Server"
click HealthSvr "https://github.com/pole-io/pole-server" "Health Server"
click go-control-plane "https://github.com/envoyproxy/go-control-plane" "Envoy Go Control Plane"
```

**Diagram sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L41-L76)

**Section sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L41-L76)

## Performance Considerations
The XDS v3 implementation includes several performance optimizations:
- Connection limiting via connlimit package to prevent resource exhaustion
- Single-flight pattern to avoid duplicate service discovery calls
- Efficient cache synchronization with 5-second ticker for configuration updates
- Atomic operations for thread-safe version number management
- Pre-computed route configuration names for inbound and outbound traffic

The system is designed to handle high volumes of Envoy proxies with minimal latency in configuration updates. The use of the go-control-plane library ensures compatibility with Envoy's streaming xDS protocol requirements.

## Troubleshooting Guide
The XDS v3 plugin provides debug endpoints to assist with troubleshooting configuration issues:

```mermaid
flowchart TD
A[Debug Endpoint] --> B{Path}
B --> |/debug/apiserver/xds/envoy_nodes| C["Query Envoy Nodes<br/>Parameters: type=[sidecar,gateway]"]
B --> |/debug/apiserver/xds/resources| D["Query Resources<br/>Parameters: type=[eds,cds,rds,vhds,lds]<br/>nodeId"]
C --> E[Return Node List]
D --> F[Return Resource List]
style A fill:#4CAF50,stroke:#388E3C
style E fill:#4CAF50,stroke:#388E3C
style F fill:#4CAF50,stroke:#388E3C
```

**Diagram sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L470-L483)

**Section sources**
- [server.go](file://plugin/apiserver/xdsserverv3/server.go#L470-L483)

## Conclusion
The XDS v3 Server Plugin provides a comprehensive implementation of the Envoy xDS APIs for service mesh configuration. By leveraging the go-control-plane library and integrating with Polaris service registry and policy systems, it delivers dynamic configuration updates for sidecar proxies. The resource model in `resource/model.go` effectively transforms service routing rules into Envoy-compatible configurations, while the cache mechanism ensures efficient distribution of incremental updates. The implementation supports all major xDS resources (CDS, RDS, LDS, EDS, VHDS, HDS) and includes debugging capabilities for operational visibility.