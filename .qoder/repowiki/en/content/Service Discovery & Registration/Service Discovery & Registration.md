# Service Discovery & Registration

<cite>
**Referenced Files in This Document**   
- [instance.go](file://apis/pkg/types/service/instance.go)
- [service_contract.go](file://pkg/service/service_contract.go)
- [service_contract_test.go](file://pkg/service/service_contract_test.go)
- [service_contract.go](file://pkg/cache/service/service_contract.go)
- [service_contract.go](file://plugin/store/mysql/service_contract.go)
- [server.go](file://pkg/service/healthcheck/server.go)
- [check.go](file://pkg/service/healthcheck/check.go)
- [beat_checker.go](file://plugin/service/healthchecker/heartbeat/beat_checker.go)
- [beat_checker_test.go](file://plugin/service/healthchecker/heartbeat/beat_checker_test.go)
- [heartbeat_access.go](file://plugin/apiserver/grpcserver/discover/v1/heartbeat_access.go)
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go)
- [nacosserver/v2/discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go)
- [eurekaserver/model.go](file://plugin/apiserver/eurekaserver/model.go)
- [xdsserverv3/server.go](file://plugin/apiserver/xdsserverv3/server.go)
- [apolloserver/server.go](file://plugin/apiserver/apolloserver/server.go)
- [service_alias.go](file://pkg/service/service_alias.go)
</cite>

## Table of Contents
1. [Data Model for Services and Instances](#data-model-for-services-and-instances)
2. [Heartbeat Mechanism and TTL-Based Expiration](#heartbeat-mechanism-and-ttl-based-expiration)
3. [Registration, Deregistration, and Querying Workflows](#registration-deregistration-and-querying-workflows)
4. [Health Checking Implementation](#health-checking-implementation)
5. [Multi-Protocol Support](#multi-protocol-support)
6. [Service Aliasing and Contract Management](#service-aliasing-and-contract-management)
7. [Consistency Models and Caching Strategies](#consistency-models-and-caching-strategies)
8. [Troubleshooting Common Issues](#troubleshooting-common-issues)

## Data Model for Services and Instances

The service discovery system implements a comprehensive data model for services and instances with rich metadata support. Each service instance is represented by the `Instance` struct defined in [instance.go](file://apis/pkg/types/service/instance.go), which contains core attributes such as host, port, protocol, and metadata. The metadata field supports key-value pairs with validation constraints on both the number of entries and character limits, as demonstrated in the test cases in [instance_test.go](file://pkg/service/instance_test.go).

Service instances can be queried using metadata filters, allowing clients to discover services based on specific metadata key-value pairs. The system supports complex queries with multiple metadata conditions, returning instances that match all specified criteria. When metadata queries include non-existent keys or values, the system returns empty results, ensuring predictable behavior for client applications.

Service contracts are managed through the service contract subsystem, which stores interface definitions, protocols, versions, and associated metadata. The contract data model includes fields for ID, namespace, service name, protocol, version, revision, content, and metadata. Contracts are identified by a unique SHA-1 hash generated from the namespace, service, name, protocol, and version fields, ensuring global uniqueness across the system.

**Section sources**
- [instance.go](file://apis/pkg/types/service/instance.go#L152-L230)
- [instance_test.go](file://pkg/service/instance_test.go#L939-L969)

## Heartbeat Mechanism and TTL-Based Expiration

The system implements a robust heartbeat mechanism for service instance health monitoring. Instances with heartbeat-based health checking enabled must periodically send heartbeat signals to maintain their healthy status. The default TTL (Time To Live) value is configured in the system, with instances required to heartbeat within this interval to remain active.

Heartbeat operations are processed through the health check server, which receives heartbeat reports via the `Report` and `Reports` methods in [server.go](file://pkg/service/healthcheck/server.go). The system supports both individual and batch heartbeat reporting, allowing clients to optimize network usage. Heartbeat records can be queried and deleted in bulk through dedicated endpoints in [heartbeat_access.go](file://plugin/apiserver/grpcserver/discover/v1/heartbeat_access.go).

When an instance fails to heartbeat within the configured TTL period, the system automatically marks it as unhealthy. The expiration process is managed by the check scheduler, which uses a time wheel implementation to efficiently schedule health checks. The system implements a three-phase expiration model where instances must miss multiple consecutive heartbeats before being considered expired, providing resilience against transient network issues.

The last heartbeat timestamp is stored in instance metadata under the key `last_heartbeat_time`, allowing clients to query the most recent heartbeat time for any instance. This metadata is automatically updated on successful heartbeat processing and removed when an instance transitions from unhealthy to healthy state.

**Section sources**
- [server.go](file://pkg/service/healthcheck/server.go#L240-L267)
- [heartbeat_access.go](file://plugin/apiserver/grpcserver/discover/v1/heartbeat_access.go#L38-L96)
- [check.go](file://pkg/service/healthcheck/check.go#L400-L450)

## Registration, Deregistration, and Querying Workflows

Service registration follows a standardized workflow where clients submit instance information including host, port, service name, namespace, and optional metadata. The system validates the registration request, generates a unique instance ID if not provided, and stores the instance in the backend storage. Registration is idempotent, allowing clients to safely retry registration without creating duplicate instances.

Deregistration can occur through explicit client requests or automatically when instances fail health checks. The system supports both graceful shutdown notifications and automatic cleanup of stale instances. When an instance is deregistered, the system publishes instance event notifications to inform subscribers of the change.

Instance querying supports multiple filtering criteria including service name, namespace, metadata, and health status. Clients can retrieve all instances for a service or filter by specific metadata key-value pairs. The query interface returns comprehensive instance information including all metadata, health status, and registration details. The system implements efficient indexing on commonly queried fields to ensure fast response times even with large numbers of registered instances.

The batch operations subsystem optimizes registration and deregistration workflows by aggregating multiple operations into single transactions, improving throughput and reducing database load. This is particularly beneficial for services that register multiple instances simultaneously.

**Section sources**
- [instance.go](file://apis/pkg/types/service/instance.go#L588-L656)
- [service.go](file://pkg/service/service.go#L100-L150)
- [instance_test.go](file://pkg/service/instance_test.go#L155-L180)

## Health Checking Implementation

The health checking system implements both active and passive monitoring modes. Passive health checking is based on the heartbeat mechanism described earlier, where instances must actively report their status at regular intervals. Active health checking involves the server periodically probing instances using configured health check methods.

The health check server manages multiple checker implementations, with heartbeat checking being the primary method. Checkers are registered with the server and can be queried by type. The system uses a dispatcher to route health check operations to the appropriate checker implementation based on the instance's health check configuration.

Health check results trigger state transitions that are published as events to the event hub. When an instance changes from healthy to unhealthy or vice versa, the system generates an instance event that can be consumed by monitoring systems or used to update service routing tables. The health check scheduler uses a time wheel to efficiently manage the scheduling of health checks across potentially thousands of instances.

The system implements a distributed health checking model where multiple server instances coordinate health check responsibilities. Each server instance is responsible for checking a subset of instances based on consistent hashing, ensuring balanced load distribution and fault tolerance. If a server instance fails, its health checking responsibilities are automatically redistributed to remaining instances.

**Section sources**
- [server.go](file://pkg/service/healthcheck/server.go#L0-L268)
- [check.go](file://pkg/service/healthcheck/check.go#L0-L630)
- [beat_checker.go](file://plugin/service/healthchecker/heartbeat/beat_checker.go#L0-L200)

## Multi-Protocol Support

The service discovery system provides comprehensive multi-protocol support through dedicated server implementations for each protocol. The system supports Eureka's REST API, Nacos's HTTP/gRPC interfaces, Apollo's JSON format, and XDS v3, allowing seamless integration with diverse service mesh and microservices architectures.

The Nacos server implementation in [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go) and [nacosserver/v2/discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go) provides both HTTP and gRPC endpoints for service registration and discovery. The gRPC implementation uses a handler registry to map request types to appropriate processing functions, supporting operations like instance registration, service querying, and subscription management.

The Eureka server implementation exposes the standard Eureka REST API endpoints for application and instance management. The XDS v3 server provides Envoy-compatible discovery service endpoints for CDS, EDS, LDS, RDS, and HDS, enabling integration with Istio and other Istio-compatible service meshes.

The Apollo server implementation supports Apollo's configuration and service discovery endpoints, allowing Apollo clients to register services and discover dependencies. All protocol implementations translate between the external protocol format and the internal service model, ensuring consistent behavior regardless of the access protocol used.

**Section sources**
- [nacosserver/server.go](file://plugin/apiserver/nacosserver/server.go#L37-L91)
- [nacosserver/v2/discover/server.go](file://plugin/apiserver/nacosserver/v2/discover/server.go#L83-L125)
- [eurekaserver/model.go](file://plugin/apiserver/eurekaserver/model.go#L10-L50)
- [xdsserverv3/server.go](file://plugin/apiserver/xdsserverv3/server.go#L10-L50)
- [apolloserver/server.go](file://plugin/apiserver/apolloserver/server.go#L10-L50)

## Service Aliasing and Contract Management

Service aliasing allows multiple logical service names to refer to the same physical service, enabling flexible service routing and migration scenarios. The aliasing system is implemented in [service_alias.go](file://pkg/service/service_alias.go), providing operations to create, update, and delete service aliases. Aliases include metadata that can be used for routing decisions and support TTL-based expiration.

Service contract management provides a structured way to define and manage service interfaces. Contracts are stored in the database with fields for type, namespace, service, protocol, version, revision, content, and metadata. The system supports creating, updating, and deleting contracts through dedicated API endpoints. Contract interfaces can be added manually or reported by clients, with source tracking to distinguish between manual and client-reported interfaces.

Contract operations include validation of required fields and generation of unique contract IDs based on namespace, service, name, protocol, and version. The system maintains revision history for contracts, allowing rollbacks to previous versions. Contract queries support filtering by all major fields, enabling clients to discover contracts based on specific criteria.

The contract system integrates with the service discovery mechanism, allowing clients to discover not just service instances but also the interfaces they expose. This enables advanced routing and validation scenarios where clients can verify compatibility before establishing connections.

**Section sources**
- [service_alias.go](file://pkg/service/service_alias.go#L10-L100)
- [service_contract.go](file://pkg/service/service_contract.go#L41-L96)
- [service_contract.go](file://pkg/cache/service/service_contract.go#L131-L176)
- [service_contract.go](file://plugin/store/mysql/service_contract.go#L36-L72)

## Consistency Models and Caching Strategies

The system implements a multi-layer caching strategy to ensure high performance and scalability. The primary cache layer is managed by the cache manager in [pkg/cache](file://pkg/cache), which maintains in-memory representations of services, instances, and other configuration data. Cache updates are triggered by database changes, ensuring consistency between the persistent store and cached data.

Service contract caching is implemented in [service_contract.go](file://pkg/cache/service/service_contract.go), using BoltDB as a persistent cache backend. The cache stores contract data in buckets keyed by contract ID, enabling efficient lookups. The cache is updated transactionally to maintain data consistency and supports value serialization to JSON for storage.

The system employs a write-through caching pattern where updates are applied to both the cache and database atomically. For read operations, the system first checks the in-memory cache before falling back to database queries, significantly reducing database load. Cache invalidation is handled through event-driven updates, where changes to service data trigger cache refresh operations.

Consistency is maintained through revision tracking, where each service and contract has a revision field that is updated on every modification. Clients can use revision numbers to implement optimistic locking and detect concurrent modifications. The system also supports long-polling for cache updates, allowing clients to receive near-real-time notifications of changes.

**Section sources**
- [service_contract.go](file://pkg/cache/service/service_contract.go#L0-L42)
- [cache.go](file://pkg/cache/cache.go#L10-L100)
- [service.go](file://pkg/cache/service/service.go#L50-L100)

## Troubleshooting Common Issues

### Missed Heartbeats
When instances experience missed heartbeats, first verify network connectivity between the client and server. Check firewall rules and ensure the heartbeat port is accessible. Review client-side logs for heartbeat transmission errors and server-side logs for heartbeat processing failures. If using batch heartbeats, ensure the batch size is within acceptable limits to prevent timeouts.

### Stale Instances
Stale instances that remain in the registry after shutdown may indicate issues with graceful deregistration. Ensure clients send deregistration requests before terminating. Verify that the health check TTL is appropriately configured for your environment - too long a TTL may keep failed instances active too long, while too short a TTL may cause false positives during network congestion.

### Metadata Propagation Issues
When metadata changes are not reflected in queries, verify that the metadata update operation completed successfully. Check that metadata key and value limits are not exceeded, as this will cause update operations to fail. Ensure that cache updates are propagating by forcing a cache refresh or waiting for the next scheduled update cycle.

### High Memory Usage
If the server experiences high memory usage, review the number of registered services and instances. Consider implementing service cleanup policies for unused services. Monitor cache hit rates and adjust cache sizes if necessary. For environments with high churn, tune the health check scheduler parameters to balance responsiveness with resource usage.

### Protocol Compatibility Issues
When integrating with specific protocols, verify that the request format matches the expected schema. For Nacos gRPC, ensure the payload type matches the registered handler. For Eureka REST API, validate that JSON/XML formatting conforms to the expected structure. Use protocol-specific debugging endpoints where available to inspect request processing.

**Section sources**
- [server.go](file://pkg/service/healthcheck/server.go#L240-L267)
- [check.go](file://pkg/service/healthcheck/check.go#L500-L600)
- [instance_test.go](file://pkg/service/instance_test.go#L2828-L2895)
- [beat_checker_test.go](file://plugin/service/healthchecker/heartbeat/beat_checker_test.go#L55-L113)