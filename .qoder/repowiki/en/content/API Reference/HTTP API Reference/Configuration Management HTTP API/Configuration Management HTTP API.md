# Configuration Management HTTP API

<cite>
**Referenced Files in This Document**   
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_group.go](file://pkg/config/config_file_group.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go)
- [config_file_template.go](file://pkg/config/config_file_template.go)
- [config_file_api.go](file://apis/store/config_file_api.go)
- [types.go](file://apis/pkg/types/config/types.go)
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go)
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
This document provides comprehensive API documentation for the Configuration Management HTTP endpoints in pole-server. It covers all operations under the `/config` path, including configuration file lifecycle management, group management, release processes, version history, and real-time update mechanisms. The system supports robust configuration management capabilities for distributed environments with features like long-polling, release rollback, and configuration templates.

## Project Structure
The configuration management system is organized across multiple packages with clear separation of concerns. The core logic resides in the `pkg/config` directory, with supporting types in `apis/pkg/types/config`, storage operations in `apis/store`, and HTTP server implementations in `plugin/apiserver/httpserver`.

```mermaid
graph TB
subgraph "Core Logic"
CF[config_file.go]
CG[config_file_group.go]
CR[config_file_release.go]
CH[config_file_release_history.go]
CT[config_file_template.go]
end
subgraph "Types & Storage"
T[types.go]
S[config_file_api.go]
end
subgraph "HTTP Interface"
H[config_server_apidoc.go]
end
CF --> S
CG --> S
CR --> S
CH --> S
CT --> S
H --> CF
H --> CG
H --> CR
H --> CH
H --> CT
```

**Diagram sources**
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_group.go](file://pkg/config/config_file_group.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go)
- [config_file_template.go](file://pkg/config/config_file_template.go)
- [config_file_api.go](file://apis/store/config_file_api.go)
- [types.go](file://apis/pkg/types/config/types.go)
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go)

**Section sources**
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_group.go](file://pkg/config/config_file_group.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)

## Core Components
The configuration management system consists of five core components: configuration files, configuration groups, configuration releases, release history, and configuration templates. These components work together to provide a complete configuration management solution with version control, rollback capabilities, and template-based configuration.

**Section sources**
- [config_file.go](file://pkg/config/config_file.go#L1-L565)
- [config_file_group.go](file://pkg/config/config_file_group.go#L1-L320)
- [config_file_release.go](file://pkg/config/config_file_release.go#L1-L789)

## Architecture Overview
The configuration management system follows a service-oriented architecture with clear separation between business logic, data access, and HTTP interface layers. The system uses transactional operations to ensure data consistency during configuration changes and provides comprehensive history tracking for audit purposes.

```mermaid
graph TD
Client --> |HTTP Requests| HTTPServer
HTTPServer --> ConfigService
ConfigService --> Storage
Storage --> Database[(Database)]
ConfigService --> Cache
ConfigService --> History
History --> AuditLog[(Audit Log)]
style HTTPServer fill:#f9f,stroke:#333
style ConfigService fill:#bbf,stroke:#333
style Storage fill:#f96,stroke:#333
style Cache fill:#9f9,stroke:#333
style History fill:#6cf,stroke:#333
```

**Diagram sources**
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go)

## Detailed Component Analysis

### Configuration File Management
The configuration file component handles the creation, retrieval, update, and deletion of configuration files. Each configuration file belongs to a namespace and group, and contains content, format, metadata, and optional encryption settings.

#### Configuration File Operations
```mermaid
classDiagram
class ConfigFile {
+string namespace
+string group
+string name
+string content
+string format
+map~string,string~ metadata
+string comment
+bool encrypt
+string encryptAlgo
+string createBy
+string modifyBy
+string createTime
+string modifyTime
}
class ConfigResponse {
+uint32 code
+string info
+ConfigFile configFile
}
class ConfigBatchWriteResponse {
+uint32 code
+string info
+uint32 total
+uint32 successCount
}
ConfigResponse --> ConfigFile : "contains"
ConfigBatchWriteResponse --> ConfigFile : "contains multiple"
```

**Diagram sources**
- [config_file.go](file://pkg/config/config_file.go#L1-L565)
- [types.go](file://apis/pkg/types/config/types.go)

**Section sources**
- [config_file.go](file://pkg/config/config_file.go#L1-L565)
- [config_file_api.go](file://apis/store/config_file_api.go)

### Configuration Group Management
Configuration groups provide a way to organize related configuration files. Groups can be created, updated, and deleted, and each group can contain multiple configuration files. The system automatically creates groups when creating configuration files if they don't exist.

#### Configuration Group Operations
```mermaid
classDiagram
class ConfigFileGroup {
+uint64 id
+string namespace
+string name
+string comment
+string business
+string department
+map~string,string~ metadata
+string createBy
+string modifyBy
+string createTime
+string modifyTime
+uint64 fileCount
+bool editable
}
class ConfigBatchQueryResponse {
+uint32 code
+string info
+uint32 total
+ConfigFileGroup[] configFileGroups
}
ConfigBatchQueryResponse --> ConfigFileGroup : "contains multiple"
```

**Diagram sources**
- [config_file_group.go](file://pkg/config/config_file_group.go#L1-L320)
- [types.go](file://apis/pkg/types/config/types.go)

**Section sources**
- [config_file_group.go](file://pkg/config/config_file_group.go#L1-L320)

### Configuration Release Management
The configuration release component manages the publishing of configuration changes to clients. Releases can be normal (full) releases or gray (canary) releases with specific labels. The system tracks active releases and provides rollback capabilities.

#### Configuration Release Workflow
```mermaid
sequenceDiagram
participant Client
participant ConfigService
participant Storage
participant Cache
Client->>ConfigService : PublishConfigFile(request)
ConfigService->>Storage : StartTx()
ConfigService->>Storage : GetConfigFile(fileKey)
alt File not found
Storage-->>ConfigService : nil
ConfigService-->>Client : NotFoundResource
else File found
Storage-->>ConfigService : configFile
ConfigService->>Storage : CreateConfigFileRelease(release)
ConfigService->>Storage : SaveGrayRule(if gray release)
ConfigService->>Storage : CommitTx()
ConfigService->>Cache : Update()
ConfigService->>ConfigService : RecordHistory()
ConfigService-->>Client : ExecuteSuccess
end
```

**Diagram sources**
- [config_file_release.go](file://pkg/config/config_file_release.go#L1-L789)
- [config_file.go](file://pkg/config/config_file.go#L1-L565)

**Section sources**
- [config_file_release.go](file://pkg/config/config_file_release.go#L1-L789)

### Configuration Release History
The release history component maintains an audit trail of all configuration changes. This includes creation, updates, deletions, and releases of configuration files. The history is used for rollback operations and auditing purposes.

#### Release History Data Model
```mermaid
erDiagram
CONFIG_FILE_RELEASE_HISTORY {
string name PK
string namespace PK
string group PK
string file_name PK
string content
string format
map metadata
string comment
string md5
uint64 version
string type
string status
string reason
string create_by
string modify_by
string release_description
timestamp create_time
timestamp modify_time
}
```

**Diagram sources**
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go#L1-L99)
- [config_file_release.go](file://pkg/config/config_file_release.go#L1-L789)

**Section sources**
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go#L1-L99)

### Configuration Templates
Configuration templates provide a way to define reusable configuration patterns. Templates can be created, updated, and retrieved by name, and are used to standardize configuration across different services and environments.

#### Template Management Flow
```mermaid
flowchart TD
Start([Create Template]) --> ValidateName["Validate Template Name"]
ValidateName --> NameExists{"Name Exists?"}
NameExists --> |Yes| ReturnError["Return ExistedResource"]
NameExists --> |No| SaveTemplate["Save Template to Storage"]
SaveTemplate --> UpdateMetadata["Set CreateBy/ModifyBy"]
UpdateMetadata --> ReturnSuccess["Return ExecuteSuccess"]
style Start fill:#f9f,stroke:#333
style ReturnError fill:#f96,stroke:#333
style ReturnSuccess fill:#9f9,stroke:#333
```

**Diagram sources**
- [config_file_template.go](file://pkg/config/config_file_template.go#L1-L148)
- [types.go](file://apis/pkg/types/config/types.go)

**Section sources**
- [config_file_template.go](file://pkg/config/config_file_template.go#L1-L148)

## Dependency Analysis
The configuration management system has well-defined dependencies between components. The HTTP interface depends on the configuration service, which in turn depends on storage and cache components. The release functionality depends on both the file and group components, creating a hierarchical dependency structure.

```mermaid
graph TD
HTTPInterface --> ConfigService
ConfigService --> ConfigFile
ConfigService --> ConfigGroup
ConfigService --> ConfigRelease
ConfigService --> ConfigHistory
ConfigService --> ConfigTemplate
ConfigFile --> Storage
ConfigGroup --> Storage
ConfigRelease --> Storage
ConfigHistory --> Storage
ConfigTemplate --> Storage
ConfigService --> Cache
style HTTPInterface fill:#f9f,stroke:#333
style ConfigService fill:#bbf,stroke:#333
style Storage fill:#f96,stroke:#333
style Cache fill:#9f9,stroke:#333
```

**Diagram sources**
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_group.go](file://pkg/config/config_file_group.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go)
- [config_file_template.go](file://pkg/config/config_file_template.go)
- [config_file_api.go](file://apis/store/config_file_api.go)

**Section sources**
- [config_file.go](file://pkg/config/config_file.go)
- [config_file_group.go](file://pkg/config/config_file_group.go)
- [config_file_release.go](file://pkg/config/config_file_release.go)

## Performance Considerations
The configuration management system is designed with performance in mind. It uses caching to reduce database load for read operations, batch processing for bulk operations, and pagination for large result sets. Write operations are transactional to ensure data consistency but may have higher latency due to the atomicity requirements.

## Troubleshooting Guide
Common issues in the configuration management system include configuration not being received by clients, watch timeouts, and large configuration file handling. Ensure that clients are properly connected to the configuration server and that watch timeouts are set appropriately for the network conditions. For large configuration files, consider breaking them into smaller, more manageable pieces.

**Section sources**
- [config_file.go](file://pkg/config/config_file.go#L1-L565)
- [config_file_release.go](file://pkg/config/config_file_release.go#L1-L789)
- [config_file_release_history.go](file://pkg/config/config_file_release_history.go#L1-L99)

## Conclusion
The Configuration Management HTTP API in pole-server provides a comprehensive solution for managing configuration in distributed systems. With support for configuration files, groups, releases, history, and templates, the system offers robust capabilities for configuration lifecycle management. The well-structured architecture ensures reliability and scalability, while the comprehensive API enables integration with various client applications.