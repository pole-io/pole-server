# Performance Tuning

<cite>
**Referenced Files in This Document**   
- [config_file.go](file://pkg/cache/config/config_file.go)
- [config_group.go](file://pkg/cache/config/config_group.go)
- [base_db.go](file://plugin/store/mysql/base_db.go)
- [listener.go](file://pkg/common/conn/limit/listener.go)
- [keepalive.go](file://pkg/common/conn/keepalive/keepalive.go)
- [healthcheck/config.go](file://pkg/service/healthcheck/config.go)
- [tool.go](file://plugin/store/mysql/tool.go)
</cite>

## Table of Contents
1. [Cache Sizing and Eviction Strategies](#cache-sizing-and-eviction-strategies)
2. [Database Connection Pool Configuration](#database-connection-pool-configuration)
3. [Goroutine Management and Connection Throttling](#goroutine-management-and-connection-throttling)
4. [Health Check Tuning](#health-check-tuning)
5. [Benchmarking Methodologies](#benchmarking-methodologies)
6. [Network-Level Optimizations](#network-level-optimizations)

## Cache Sizing and Eviction Strategies

The pole-server implements a multi-layered caching system for both service and configuration data. Configuration file caching is managed through the `fileCache` structure which utilizes bbolt for persistent storage of configuration content to reduce memory footprint. The cache path can be configured via the `cachePath` option in the initialization parameters, with a default path of `./data/cache/config`. Configuration group caching uses in-memory synchronization maps with revision computation for namespace-level grouping.

TTL settings are not explicitly configured for caches as they rely on periodic update mechanisms driven by the cache manager's update cycle. The eviction policy is primarily based on the validity flag in stored records - when a configuration release or group is marked as invalid, it is removed from the active caches. The system automatically manages cache consistency by processing create, update, and delete operations during the cache update cycle, ensuring that only valid and active configurations remain in memory.

The cache implementation uses single-flight pattern to prevent thundering herd problems during updates, ensuring that only one update operation proceeds while others wait for its completion. This approach minimizes database load during cache refresh operations.

**Section sources**
- [config_file.go](file://pkg/cache/config/config_file.go#L1-L612)
- [config_group.go](file://pkg/cache/config/config_group.go#L1-L303)

## Database Connection Pool Configuration

Database connection pool settings are configured through the `dbConfig` structure in the MySQL store implementation. The following parameters can be tuned for optimal database performance:

- **maxOpenConns**: Maximum number of open connections to the database. When set to a positive value, it limits the total number of connections in the pool.
- **maxIdleConns**: Maximum number of idle connections in the pool. The system maintains up to this number of idle connections for reuse.
- **connMaxLifetime**: Maximum lifetime of a connection in seconds. Connections older than this duration are closed and removed from the pool.

The system sets a read committed isolation level by default and implements automatic retry logic for transient database errors such as deadlocks, bad connections, and invalid connections. The retry mechanism follows an exponential backoff pattern, with up to 20 retry attempts and increasing wait times between attempts.

Query optimization is achieved through the use of prepared statements and batch processing capabilities. The system implements query batching with a maximum batch size of 200 operations per batch to balance between efficiency and memory usage. Additionally, the system provides metrics reporting for all database operations, allowing for performance monitoring and bottleneck identification.

**Section sources**
- [base_db.go](file://plugin/store/mysql/base_db.go#L1-L276)

## Goroutine Management and Connection Throttling

The pole-server implements connection throttling through the `Listener` structure in the connection limit package. This mechanism provides two levels of connection control:

1. **Per-host connection limit**: Controlled by `maxConnPerHost`, limiting the number of concurrent connections from a single client IP address.
2. **Global connection limit**: Controlled by `maxConnLimit`, limiting the total number of concurrent connections to the server.

The system maintains active connection tracking using synchronized maps, allowing for efficient connection counting and cleanup. Connection cleanup is performed periodically with configurable purge intervals and expiration times. The default purge counter interval is one hour, with a purge counter expiration of 300 seconds.

Goroutine management is handled through the use of worker pools and controlled goroutine creation. The system avoids unbounded goroutine creation by using bounded worker pools for processing tasks such as health checks and configuration updates. The health check system uses a slot-based scheduler with a default of 30 slots, distributing health check operations across time to prevent resource spikes.

**Section sources**
- [listener.go](file://pkg/common/conn/limit/listener.go#L1-L269)

## Health Check Tuning

Health check intervals and concurrency levels are configured through the `Config` structure in the healthcheck package. Key configuration parameters include:

- **MinCheckInterval**: Minimum interval between health checks, defaulting to 1 second.
- **MaxCheckInterval**: Maximum interval between health checks, defaulting to 30 seconds.
- **ClientCheckInterval**: Interval at which clients report their health status, defaulting to 120 seconds.
- **ClientCheckTtl**: Time-to-live for client health reports, defaulting to 120 seconds.
- **SlotNum**: Number of time slots in the health check scheduler, defaulting to 30.

The health check system automatically adjusts check intervals based on service load and network conditions. For clusters with multiple nodes, the concurrency level should be tuned based on the total number of services and instances being monitored. The system uses a distributed health check approach where each server instance is responsible for checking a subset of services, reducing the load on any single node.

Health check operations are scheduled using a time-wheel algorithm that distributes checks evenly across time slots, preventing thundering herd scenarios where many checks would otherwise occur simultaneously.

**Section sources**
- [healthcheck/config.go](file://pkg/service/healthcheck/config.go#L1-L53)

## Benchmarking Methodologies

The pole-server includes benchmarking tools for performance evaluation, located in the test/benchmark directory. The primary benchmarking tool uses ghz for gRPC performance testing, allowing measurement of various service discovery operations including instance registration, heartbeat reporting, instance queries, and deregistration.

Benchmarking should be conducted using representative workloads that match expected production usage patterns. The test environment should closely mirror the production environment in terms of hardware specifications and network conditions. Multiple test runs should be conducted to ensure statistical significance, with results averaged across runs.

Performance metrics should include:
- Request latency (p50, p90, p99 percentiles)
- Requests per second (RPS)
- Error rates
- Memory and CPU utilization
- Connection establishment times

The benchmarking framework allows for testing different cluster sizes and configurations, enabling capacity planning and identification of scaling bottlenecks.

**Section sources**
- [discovery_benchmark-zh.md](file://test/benchmark/grpc/client/discovery_benchmark-zh.md#L1-L33)

## Network-Level Optimizations

Network performance is optimized through several mechanisms:

**Keep-alive Settings**: The system implements TCP keep-alive with a default period of 3 minutes, using the `TcpKeepAliveListener` to manage persistent connections. This setting helps maintain connection state through network intermediaries and reduces the overhead of establishing new connections for frequent operations.

**Connection Management**: The system uses connection pooling and reuse to minimize the overhead of connection establishment. The keep-alive listener automatically sets keep-alive options on accepted connections, including enabling keep-alive and setting the keep-alive period.

**Payload Compression**: While not explicitly configured in the provided code, the system supports payload compression through the underlying gRPC framework. Compression can be enabled for specific endpoints based on the content type and size.

**Time Synchronization**: The system includes mechanisms for time synchronization between server and clients, using the `GetUnixSecond` function to retrieve server time with configurable maximum wait times. This ensures consistent time references across distributed components, which is critical for operations such as TTL calculations and health check scheduling.

**Section sources**
- [keepalive.go](file://pkg/common/conn/keepalive/keepalive.go#L1-L60)
- [tool.go](file://plugin/store/mysql/tool.go#L1-L55)