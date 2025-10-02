# Configuration Management

<cite>
**Referenced Files in This Document**   
- [config.go](file://pkg/admin/config.go)
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go)
- [pole-server.yaml](file://deploy/conf/pole-server.yaml)
- [pole-apiserver.yaml](file://deploy/conf/pole-apiserver.yaml)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [Configuration Reload Endpoint](#configuration-reload-endpoint)
3. [Reload Mechanism](#reload-mechanism)
4. [Impact on Running Services](#impact-on-running-services)
5. [Usage Examples](#usage-examples)
6. [Common Issues and Troubleshooting](#common-issues-and-troubleshooting)
7. [Best Practices](#best-practices)
8. [Relationship with Maintenance Mode](#relationship-with-maintenance-mode)

## Introduction
This document provides comprehensive API documentation for the configuration management functionality in the Administration HTTP API. It focuses on the runtime configuration reload capability, which allows administrators to apply configuration changes without restarting the server. The system supports dynamic reloading of key configuration files, enabling seamless updates to server behavior while maintaining service availability.

## Configuration Reload Endpoint
The POST `/admin/reload` endpoint triggers the runtime reloading of configuration files. This operation allows administrators to apply configuration changes without requiring a server restart, minimizing downtime and service disruption.

The endpoint accepts an empty request body and returns a success confirmation upon completion of the reload process. Authentication is required to invoke this endpoint, ensuring only authorized users can modify system configuration.

```mermaid
sequenceDiagram
participant Client
participant AdminAPI
participant ConfigManager
Client->>AdminAPI : POST /admin/reload
AdminAPI->>ConfigManager : Trigger configuration reload
ConfigManager->>ConfigManager : Load pole-server.yaml
ConfigManager->>ConfigManager : Load pole-apiserver.yaml
ConfigManager-->>AdminAPI : Reload complete
AdminAPI-->>Client : 200 OK
```

**Diagram sources**
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)
- [config.go](file://pkg/admin/config.go#L1-L47)

**Section sources**
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)

## Reload Mechanism
The configuration reload mechanism is implemented in the `pkg/admin/config.go` file, which defines the configuration structure and default values for administrative operations. When the reload endpoint is invoked, the system re-reads the primary configuration files from the filesystem and applies the new values to the running instance.

The reload process specifically targets two configuration files:
- `pole-server.yaml`: Contains core server configuration including service discovery settings, storage configuration, and global parameters
- `pole-apiserver.yaml`: Contains API server-specific configuration including HTTP server settings, security policies, and API rate limiting rules

During reload, the system performs validation on the new configuration values before applying them. If validation fails, the reload process is aborted and the previous configuration remains active to prevent service disruption.

```mermaid
flowchart TD
Start([Reload Request]) --> ReadConfig["Read pole-server.yaml"]
ReadConfig --> ValidateCore["Validate Core Configuration"]
ValidateCore --> ReadAPIConfig["Read pole-apiserver.yaml"]
ReadAPIConfig --> ValidateAPI["Validate API Configuration"]
ValidateAPI --> ApplyConfig["Apply New Configuration"]
ApplyConfig --> NotifyServices["Notify Dependent Services"]
NotifyServices --> End([Reload Complete])
ValidateCore --> |Validation Failed| Rollback["Maintain Current Configuration"]
ValidateAPI --> |Validation Failed| Rollback
Rollback --> ErrorResponse["Return Error to Client"]
ErrorResponse --> End
```

**Diagram sources**
- [config.go](file://pkg/admin/config.go#L1-L47)
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)

**Section sources**
- [config.go](file://pkg/admin/config.go#L1-L47)

## Impact on Running Services
Configuration reloading can have transient effects on running services as components adapt to new settings. The impact varies depending on the specific configuration parameters that change:

- **Logging configuration changes**: Immediately affect log output levels and destinations
- **Rate limiting policy updates**: Propagated to all connected clients through the configuration distribution mechanism
- **Security policy modifications**: Applied to new connections; existing connections may continue under previous policies until re-established
- **Service discovery settings**: Affect how new service instances are registered and discovered

During the reload process, there may be a brief period of inconsistency as different components receive and apply the new configuration at slightly different times. The system is designed to maintain availability during this transition period, with most operations continuing normally.

The reload operation is atomic at the file level - either the entire configuration file is successfully reloaded, or the reload is aborted and the previous configuration remains in effect.

## Usage Examples
To trigger a configuration reload after updating the configuration files, use the following curl command:

```bash
curl -X POST http://<server-address>/admin/reload \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json"
```

Example with specific server address and token:
```bash
curl -X POST http://localhost:8080/admin/reload \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xJ..." \
  -H "Content-Type: application/json"
```

The response will be:
```json
{
  "code": 0,
  "message": "ok"
}
```

Before triggering a reload in production, update the configuration files in the `deploy/conf/` directory:

```bash
# Update configuration files
cp ./new-config/pole-server.yaml ./deploy/conf/pole-server.yaml
cp ./new-config/pole-apiserver.yaml ./deploy/conf/pole-apiserver.yaml

# Trigger reload
curl -X POST http://production-server/admin/reload -H "Authorization: Bearer $TOKEN"
```

**Section sources**
- [pole-server.yaml](file://deploy/conf/pole-server.yaml)
- [pole-apiserver.yaml](file://deploy/conf/pole-apiserver.yaml)
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)

## Common Issues and Troubleshooting
Several issues may occur during configuration reload operations:

**Partial Configuration Reload Failures**: If one configuration file is valid but another contains errors, the entire reload operation fails and no changes are applied. Check server logs for validation error messages.

**Validation Errors in Updated Configs**: The system validates configuration files against defined schemas. Common validation issues include:
- Invalid data types for configuration parameters
- Out-of-range numeric values
- Malformed YAML syntax
- Required fields missing

**Inconsistent Behavior with Dependent Services**: After reload, some services may temporarily operate with mixed configurations until all components have fully propagated the changes. Monitor service behavior and logs for any anomalies.

**Authentication Issues**: Ensure the admin token has sufficient privileges to perform configuration reload operations. Insufficient permissions result in 403 Forbidden responses.

**File Permission Problems**: The server process must have read permissions on the configuration files. Permission denied errors prevent successful reload.

To troubleshoot reload issues:
1. Check server logs for detailed error messages
2. Validate configuration files using YAML validators
3. Verify file permissions and ownership
4. Test configuration changes in a staging environment first
5. Use incremental changes rather than large configuration updates

**Section sources**
- [config.go](file://pkg/admin/config.go#L1-L47)
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)

## Best Practices
Follow these best practices when managing configuration changes:

**Test in Staging**: Always test configuration changes in a staging environment that mirrors production before applying them to live systems.

**Incremental Changes**: Make small, incremental configuration changes rather than large updates. This makes it easier to identify issues if problems occur.

**Backup Configurations**: Maintain backups of previous working configurations to enable quick rollback if needed.

**Rollback Strategy**: Have a documented rollback procedure. If a configuration change causes issues, immediately revert to the previous configuration and trigger another reload.

**Change Window**: Perform configuration updates during maintenance windows or periods of low traffic to minimize potential impact on users.

**Monitoring**: Monitor system metrics and logs closely after applying configuration changes to detect any adverse effects.

**Validation**: Validate configuration files with appropriate tools before deploying them to ensure they are syntactically correct and conform to the expected schema.

**Documentation**: Maintain up-to-date documentation of configuration parameters and their effects on system behavior.

## Relationship with Maintenance Mode
The configuration reload operation is closely related to maintenance mode operations. While both are administrative functions, they serve different purposes:

- **Configuration reload**: Updates the server's operational parameters without interrupting service
- **Maintenance mode**: Temporarily takes the server out of service for major updates or repairs

The reload endpoint can be used while the server is in normal operation, whereas maintenance mode operations require coordination with load balancers and service discovery systems to prevent new connections.

When planning major configuration changes, consider using maintenance mode to ensure a consistent state across the entire system. For minor parameter adjustments, the reload endpoint provides a more efficient approach with minimal disruption.

The two operations are independent - configuration reload can be performed regardless of maintenance mode status, and entering maintenance mode does not automatically trigger a configuration reload.

**Section sources**
- [admin_access.go](file://plugin/apiserver/httpserver/admin_access.go#L40-L57)
- [config.go](file://pkg/admin/config.go#L1-L47)