# HTTP API Reference

<cite>
**Referenced Files in This Document**   
- [admin_apidoc.go](file://plugin/apiserver/httpserver/docs/admin_apidoc.go)
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go)
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go)
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go)
- [core_console_apidoc.go](file://plugin/apiserver/httpserver/docs/core_console_apidoc.go)
- [naming_console_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_console_access_apidoc.go)
- [auth.go](file://apis/access_control/auth/auth.go)
- [ratelimit.go](file://apis/access_control/ratelimit/ratelimit.go)
- [auth_response.go](file://pkg/common/api/v1/auth_response.go)
- [config_response.go](file://pkg/common/api/v1/config_response.go)
- [naming_response.go](file://pkg/common/api/v1/naming_response.go)
- [client_v1.go](file://pkg/service/client_v1.go)
- [config_file.go](file://pkg/config/config_file.go)
- [server.go](file://pkg/admin/server.go)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [API Endpoints Overview](#api-endpoints-overview)
3. [Naming Service Endpoints](#naming-service-endpoints)
4. [Configuration Management Endpoints](#configuration-management-endpoints)
5. [Authentication and Authorization Endpoints](#authentication-and-authorization-endpoints)
6. [Admin and Maintenance Endpoints](#admin-and-maintenance-endpoints)
7. [Console Endpoints](#console-endpoints)
8. [Error Handling and Status Codes](#error-handling-and-status-codes)
9. [Rate Limiting and Throttling](#rate-limiting-and-throttling)
10. [Versioning Strategy](#versioning-strategy)
11. [CORS and Content Negotiation](#cors-and-content-negotiation)
12. [Internationalization (i18n) Support](#internationalization-i18n-support)
13. [Client Examples](#client-examples)
14. [Security Considerations](#security-considerations)

## Introduction
The pole-server provides a comprehensive HTTP API for service discovery, configuration management, authentication, and administrative operations. This document details all available endpoints under the `/naming`, `/config`, `/auth`, `/admin`, and `/console` paths, including request/response schemas, authentication mechanisms, and usage examples. The API supports both client-facing operations for service instances and console/admin operations for management tasks.

## API Endpoints Overview
The pole-server HTTP API is organized into several logical groups based on functionality:

- **/naming**: Service registration, heartbeat, and discovery operations
- **/config**: Configuration file management and client configuration retrieval
- **/auth**: User authentication, authorization policies, and identity management
- **/admin**: Server maintenance, monitoring, and operational commands
- **/console**: Administrative console operations for configuration and service management

All endpoints follow RESTful principles and return JSON responses with a consistent structure defined in the common API response types.

**Section sources**
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go#L1-L70)
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L1-L317)
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L1-L394)

## Naming Service Endpoints
The naming service endpoints handle service registration, health checking, and service discovery operations.

### Service Registration
Registers a service instance with the registry.

**Endpoint**: `POST /naming/v1/instance`

**Request Schema**:
```json
{
  "instance": {
    "id": "string",
    "service": "string",
    "namespace": "string",
    "host": "string",
    "port": 80,
    "protocol": "string",
    "version": "string",
    "metadata": {
      "key": "value"
    }
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "instance": {
    "id": "string",
    "service": "string",
    "namespace": "string",
    "host": "string",
    "port": 80,
    "protocol": "string",
    "version": "string",
    "metadata": {
      "key": "value"
    }
  }
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/naming/v1/instance \
  -H "Content-Type: application/json" \
  -d '{
    "instance": {
      "service": "my-service",
      "namespace": "default",
      "host": "192.168.1.100",
      "port": 8080,
      "protocol": "http"
    }
  }'
```

**Section sources**
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go#L25-L30)
- [client_v1.go](file://pkg/service/client_v1.go#L15-L50)

### Instance Heartbeat
Reports health status of a registered service instance.

**Endpoint**: `POST /naming/v1/heartbeat`

**Request Schema**:
```json
{
  "instance": {
    "id": "string",
    "service": "string",
    "namespace": "string"
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/naming/v1/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "instance": {
      "id": "instance-123",
      "service": "my-service",
      "namespace": "default"
    }
  }'
```

**Section sources**
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go#L31-L35)
- [client_v1.go](file://pkg/service/client_v1.go#L51-L75)

### Service Discovery
Discovers available service instances.

**Endpoint**: `POST /naming/v1/discover`

**Request Schema**:
```json
{
  "type": "INSTANCE",
  "service": {
    "name": "string",
    "namespace": "string"
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "instances": [
    {
      "id": "string",
      "service": "string",
      "namespace": "string",
      "host": "string",
      "port": 80,
      "healthy": true,
      "isolate": false,
      "weight": 100,
      "metadata": {
        "key": "value"
      }
    }
  ],
  "routing": {
    "inbounds": [],
    "outbounds": []
  }
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/naming/v1/discover \
  -H "Content-Type: application/json" \
  -d '{
    "type": "INSTANCE",
    "service": {
      "name": "my-service",
      "namespace": "default"
    }
  }'
```

**Section sources**
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go#L36-L40)
- [client_v1.go](file://pkg/service/client_v1.go#L76-L120)

### Instance Deregistration
Removes a service instance from the registry.

**Endpoint**: `POST /naming/v1/deregister`

**Request Schema**:
```json
{
  "instance": {
    "id": "string",
    "service": "string",
    "namespace": "string"
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/naming/v1/deregister \
  -H "Content-Type: application/json" \
  -d '{
    "instance": {
      "id": "instance-123",
      "service": "my-service",
      "namespace": "default"
    }
  }'
```

**Section sources**
- [naming_client_access_apidoc.go](file://plugin/apiserver/httpserver/docs/naming_client_access_apidoc.go#L20-L24)
- [client_v1.go](file://pkg/service/client_v1.go#L121-L145)

## Configuration Management Endpoints
The configuration management endpoints provide CRUD operations for configuration files and client configuration retrieval.

### Create Configuration File
Creates a new configuration file.

**Endpoint**: `POST /config/v1/config_file`

**Request Schema**:
```json
{
  "configFile": {
    "namespace": "string",
    "group": "string",
    "name": "string",
    "content": "string",
    "format": "string",
    "comment": "string",
    "tags": {
      "key": "value"
    }
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/config/v1/config_file \
  -H "Content-Type: application/json" \
  -d '{
    "configFile": {
      "namespace": "default",
      "group": "app",
      "name": "application.yaml",
      "content": "server:\n  port: 8080\nlogging:\n  level: INFO",
      "format": "yaml"
    }
  }'
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L55-L60)
- [config_file.go](file://pkg/config/config_file.go#L25-L50)

### Get Configuration File
Retrieves a configuration file by namespace, group, and name.

**Endpoint**: `GET /config/v1/config_file`

**Query Parameters**:
- `namespace`: Configuration namespace (required)
- `group`: Configuration group (required)
- `name`: Configuration file name (required)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "configFile": {
    "namespace": "string",
    "group": "string",
    "name": "string",
    "content": "string",
    "format": "string",
    "create_time": "timestamp",
    "modify_time": "timestamp"
  }
}
```

**Example curl command**:
```bash
curl -X GET "http://localhost:8080/config/v1/config_file?namespace=default&group=app&name=application.yaml"
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L175-L185)
- [config_file.go](file://pkg/config/config_file.go#L51-L75)

### Update Configuration File
Updates an existing configuration file.

**Endpoint**: `PUT /config/v1/config_file`

**Request Schema**:
```json
{
  "configFile": {
    "namespace": "string",
    "group": "string",
    "name": "string",
    "content": "string",
    "comment": "string"
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X PUT http://localhost:8080/config/v1/config_file \
  -H "Content-Type: application/json" \
  -d '{
    "configFile": {
      "namespace": "default",
      "group": "app",
      "name": "application.yaml",
      "content": "server:\n  port: 9090\nlogging:\n  level: DEBUG"
    }
  }'
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L205-L210)
- [config_file.go](file://pkg/config/config_file.go#L76-L100)

### Delete Configuration File
Deletes a configuration file.

**Endpoint**: `DELETE /config/v1/config_file`

**Query Parameters**:
- `namespace`: Configuration namespace (required)
- `group`: Configuration group (required)
- `name`: Configuration file name (required)
- `deleteBy`: Operator identifier (optional)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X DELETE "http://localhost:8080/config/v1/config_file?namespace=default&group=app&name=application.yaml"
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L225-L235)
- [config_file.go](file://pkg/config/config_file.go#L101-L125)

### Client Configuration Retrieval
Retrieves configuration for client applications with version-based change detection.

**Endpoint**: `GET /config/v1/client_config`

**Query Parameters**:
- `namespace`: Configuration namespace (required)
- `group`: Configuration group (required)
- `fileName`: Configuration file name (required)
- `version`: Client-side version identifier (required)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "change": true,
  "configFile": {
    "namespace": "string",
    "group": "string",
    "name": "string",
    "content": "string",
    "format": "string",
    "version": 1
  }
}
```

**Example curl command**:
```bash
curl -X GET "http://localhost:8080/config/v1/client_config?namespace=default&group=app&fileName=application.yaml&version=0"
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L305-L315)
- [config_file.go](file://pkg/config/config_file.go#L126-L150)

### Configuration Watch
Long-polling endpoint for watching configuration changes.

**Endpoint**: `POST /config/v1/watch`

**Request Schema**:
```json
{
  "watchList": [
    {
      "namespace": "string",
      "group": "string",
      "fileName": "string",
      "version": 0
    }
  ]
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "change": true,
  "configFiles": [
    {
      "namespace": "string",
      "group": "string",
      "name": "string",
      "content": "string",
      "format": "string",
      "version": 1
    }
  ]
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/config/v1/watch \
  -H "Content-Type: application/json" \
  -d '{
    "watchList": [
      {
        "namespace": "default",
        "group": "app",
        "fileName": "application.yaml",
        "version": 0
      }
    ]
  }'
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L316-L325)
- [config_file.go](file://pkg/config/config_file.go#L151-L175)

## Authentication and Authorization Endpoints
The authentication endpoints handle user authentication, token management, and authorization policy operations.

### User Login
Authenticates a user and returns an authentication token.

**Endpoint**: `POST /auth/v1/login`

**Request Schema**:
```json
{
  "username": "string",
  "password": "string"
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "loginResponse": {
    "user": {
      "id": "string",
      "name": "string",
      "source": "string"
    },
    "token": "string",
    "ttl": 86400
  }
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/auth/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L100-L110)
- [auth.go](file://apis/access_control/auth/auth.go#L25-L50)

### Get Authentication Status
Retrieves the current authentication system status.

**Endpoint**: `GET /auth/v1/status`

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "optionSwitch": {
    "options": {
      "clientOen": true,
      "consoleOpen": true,
      "auth": true
    }
  }
}
```

**Example curl command**:
```bash
curl -X GET http://localhost:8080/auth/v1/status
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L25-L35)
- [auth.go](file://apis/access_control/auth/auth.go#L51-L75)

### Create Authentication Strategy
Creates a new authorization policy.

**Endpoint**: `POST /auth/v1/strategies`

**Request Schema**:
```json
{
  "id": "string",
  "name": "string",
  "action": "ALLOW|DENY",
  "permissions": [
    {
      "resource": {
        "type": "NAMESPACE|SERVICE|CONFIG_GROUP",
        "id": "string"
      },
      "allow": ["READ", "WRITE", "DELETE"]
    }
  ],
  "principals": [
    {
      "principalId": "string",
      "principalType": "USER|GROUP"
    }
  ]
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "authStrategy": {
    "id": "string",
    "name": "string",
    "action": "ALLOW|DENY",
    "permissions": [],
    "principals": [],
    "createTime": "timestamp",
    "modifyTime": "timestamp"
  }
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/auth/v1/strategies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "read-only-access",
    "action": "ALLOW",
    "permissions": [
      {
        "resource": {
          "type": "NAMESPACE",
          "id": "default"
        },
        "allow": ["READ"]
      }
    ],
    "principals": [
      {
        "principalId": "user-123",
        "principalType": "USER"
      }
    ]
  }'
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L40-L55)
- [auth.go](file://apis/access_control/auth/auth.go#L76-L100)

### Query Authentication Strategies
Retrieves a list of authentication strategies with filtering options.

**Endpoint**: `GET /auth/v1/strategies`

**Query Parameters**:
- `id`: Strategy ID (optional)
- `name`: Strategy name (optional, fuzzy search)
- `default`: "0" for custom strategies, "1" for default strategies (optional)
- `res_id`: Resource ID (optional)
- `res_type`: Resource type (optional: namespace, service, config_group)
- `principal_id`: Principal ID (optional)
- `principal_type`: Principal type (optional: user, group)
- `show_detail`: Whether to show detailed information (optional)
- `offset`: Query offset (optional, default: 0)
- `limit`: Number of results per page (optional, max: 100)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "amount": 1,
  "total": 1,
  "authStrategies": [
    {
      "id": "string",
      "name": "string",
      "action": "ALLOW|DENY",
      "permissions": [],
      "principals": [],
      "createTime": "timestamp",
      "modifyTime": "timestamp"
    }
  ]
}
```

**Example curl command**:
```bash
curl -X GET "http://localhost:8080/auth/v1/strategies?res_type=namespace&res_id=default&limit=10"
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L75-L100)
- [auth.go](file://apis/access_control/auth/auth.go#L101-L125)

## Admin and Maintenance Endpoints
Administrative endpoints for server maintenance, monitoring, and operational tasks.

### Get Server Connections
Retrieves server connection statistics.

**Endpoint**: `GET /admin/v1/connections`

**Query Parameters**:
- `protocol`: Protocol to filter by (required)
- `host`: Host to filter by (optional)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "count": 10,
  "connections": [
    {
      "id": "string",
      "host": "string",
      "port": 8080,
      "protocol": "string",
      "createTime": "timestamp"
    }
  ]
}
```

**Example curl command**:
```bash
curl -X GET "http://localhost:8080/admin/v1/connections?protocol=http"
```

**Section sources**
- [admin_apidoc.go](file://plugin/apiserver/httpserver/docs/admin_apidoc.go#L25-L35)
- [server.go](file://pkg/admin/server.go#L25-L50)

### Release Leader Election
Forcibly releases the leader role in a cluster.

**Endpoint**: `POST /admin/v1/leader/release`

**Request Schema**:
```json
{
  "electKey": "string"
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/admin/v1/leader/release \
  -H "Content-Type: application/json" \
  -d '{
    "electKey": "naming-service"
  }'
```

**Section sources**
- [admin_apidoc.go](file://plugin/apiserver/httpserver/docs/admin_apidoc.go#L125-L135)
- [server.go](file://pkg/admin/server.go#L51-L75)

### Clean Unhealthy Instances
Removes unhealthy service instances from the registry.

**Endpoint**: `POST /admin/v1/instances/clean`

**Request Schema**:
```json
{
  "flag": 1
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "cleaned": 5
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/admin/v1/instances/clean \
  -H "Content-Type: application/json" \
  -d '{
    "flag": 1
  }'
```

**Section sources**
- [admin_apidoc.go](file://plugin/apiserver/httpserver/docs/admin_apidoc.go#L65-L75)
- [server.go](file://pkg/admin/server.go#L76-L100)

### Set Log Output Level
Dynamically changes the logging level for debugging.

**Endpoint**: `POST /admin/v1/log/level`

**Request Schema**:
```json
{
  "scope": "string",
  "level": "DEBUG|INFO|WARN|ERROR"
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/admin/v1/log/level \
  -H "Content-Type: application/json" \
  -d '{
    "scope": "naming",
    "level": "DEBUG"
  }'
```

**Section sources**
- [admin_apidoc.go](file://plugin/apiserver/httpserver/docs/admin_apidoc.go#L110-L120)
- [server.go](file://pkg/admin/server.go#L101-L125)

## Console Endpoints
Console-specific endpoints for administrative operations through the web interface.

### Create User
Creates a new user account.

**Endpoint**: `POST /console/v1/users`

**Request Schema**:
```json
{
  "users": [
    {
      "name": "string",
      "password": "string",
      "source": "string",
      "owner": "string",
      "comment": "string"
    }
  ]
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "responses": [
    {
      "code": 200,
      "info": "string",
      "user": {
        "id": "string",
        "name": "string",
        "source": "string",
        "owner": "string",
        "createTime": "timestamp",
        "modifyTime": "timestamp"
      }
    }
  ]
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/console/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "users": [
      {
        "name": "new-user",
        "password": "password123",
        "source": "local"
      }
    ]
  }'
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L155-L170)
- [auth.go](file://apis/access_control/auth/auth.go#L126-L150)

### Query Users
Retrieves a list of users with filtering options.

**Endpoint**: `GET /console/v1/users`

**Query Parameters**:
- `id`: User ID (optional)
- `name`: User name (optional, fuzzy search)
- `source`: User source (optional)
- `group_id`: Group ID to filter users in a specific group (optional)
- `offset`: Query offset (optional, default: 0)
- `limit`: Number of results per page (optional, max: 100)

**Response Schema**:
```json
{
  "code": 200,
  "info": "string",
  "amount": 1,
  "total": 1,
  "users": [
    {
      "id": "string",
      "name": "string",
      "source": "string",
      "owner": "string",
      "createTime": "timestamp",
      "modifyTime": "timestamp"
    }
  ]
}
```

**Example curl command**:
```bash
curl -X GET "http://localhost:8080/console/v1/users?name=admin&limit=10"
```

**Section sources**
- [auth_apidoc.go](file://plugin/apiserver/httpserver/docs/auth_apidoc.go#L130-L150)
- [auth.go](file://apis/access_control/auth/auth.go#L151-L175)

### Create Configuration File Group
Creates a new configuration file group.

**Endpoint**: `POST /console/v1/config_file_groups`

**Request Schema**:
```json
{
  "configFileGroup": {
    "name": "string",
    "namespace": "string",
    "comment": "string"
  }
}
```

**Response Schema**:
```json
{
  "code": 200,
  "info": "string"
}
```

**Example curl command**:
```bash
curl -X POST http://localhost:8080/console/v1/config_file_groups \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "configFileGroup": {
      "name": "database-config",
      "namespace": "default"
    }
  }'
```

**Section sources**
- [config_server_apidoc.go](file://plugin/apiserver/httpserver/docs/config_server_apidoc.go#L25-L35)
- [config_file.go](file://pkg/config/config_file.go#L176-L200)

## Error Handling and Status Codes
The API uses standard HTTP status codes and a consistent error response format.

### Error Response Format
All error responses follow this structure:
```json
{
  "code": 400,
  "info": "Error description message"
}
```

### Common Error Codes
| Code | Type | Description |
|------|------|-------------|
| 200 | Success | Operation completed successfully |
| 400 | Bad Request | Invalid request parameters or body |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Insufficient permissions for the operation |
| 404 | Not Found | Requested resource does not exist |
| 409 | Conflict | Operation conflicts with current state |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Unexpected server error |
| 503 | Service Unavailable | Service temporarily unavailable |

### Specific Error Codes
The API also uses specific error codes in the response body for more granular error reporting:

- **200001**: Resource not found
- **200002**: Resource already exists
- **200003**: Invalid request parameter
- **200004**: Authentication failed
- **200005**: Permission denied
- **200006**: Rate limit exceeded
- **300001**: Storage operation failed
- **300002**: Transaction failed

**Section sources**
- [auth_response.go](file://pkg/common/api/v1/auth_response.go#L1-L50)
- [config_response.go](file://pkg/common/api/v1/config_response.go#L1-L50)
- [naming_response.go](file://pkg/common/api/v1/naming_response.go#L1-L50)

## Rate Limiting and Throttling
The API implements rate limiting to prevent abuse and ensure service stability.

### Rate Limiting Strategy
The system uses a token bucket algorithm for rate limiting, with configurable limits at multiple levels:

- **Global rate limits**: Apply to the entire API
- **Per-endpoint rate limits**: Different limits for different endpoints
- **Per-client rate limits**: Limits based on client identity
- **Per-resource rate limits**: Limits based on resource being accessed

### Rate Limit Headers
Rate-limited responses include the following headers:

- `X-RateLimit-Limit`: The maximum number of requests allowed
- `X-RateLimit-Remaining`: The number of requests remaining in the current window
- `X-RateLimit-Reset`: The time at which the rate limit will reset (Unix timestamp)

### Configurable Rate Limits
Rate limits can be configured through the ratelimit plugin configuration, with default values:

- **Client endpoints**: 100 requests per second per client
- **Console endpoints**: 10 requests per second per user
- **Admin endpoints**: 5 requests per second per administrator

**Section sources**
- [ratelimit.go](file://apis/access_control/ratelimit/ratelimit.go#L1-L50)
- [ratelimit_rule.go](file://pkg/goverrule/ratelimit_rule.go#L1-L50)

## Versioning Strategy
The API uses URL-based versioning to ensure backward compatibility.

### Version Format
All endpoints are versioned using the format `/v{major-version}` in the URL path:
- `/naming/v1/`
- `/config/v1/`
- `/auth/v1/`
- `/admin/v1/`
- `/console/v1/`

### Version Lifecycle
- **Active**: Current stable version, fully supported
- **Deprecated**: Version still functional but no longer recommended, may be removed in future
- **Removed**: Version no longer available

### Backward Compatibility
The API maintains backward compatibility within major versions:
- New endpoints and fields may be added
- Existing endpoints and fields will not be removed or changed
- Breaking changes require a new major version

**Section sources**
- [client_v1.go](file://pkg/service/client_v1.go#L1-L20)
- [config_file.go](file://pkg/config/config_file.go#L1-L20)

## CORS and Content Negotiation
The API supports cross-origin requests and content negotiation.

### CORS Policy
The server implements a configurable CORS policy:

- **Allowed Origins**: Configurable, defaults to same origin
- **Allowed Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Allowed Headers**: Content-Type, Authorization, X-Requested-With, X-Real-IP
- **Exposed Headers**: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
- **Credentials**: Allowed (with specific origin configuration)
- **Max Age**: 86400 seconds (24 hours)

### Content Negotiation
The API supports content negotiation through headers:

- **Accept**: Specifies preferred response format (application/json)
- **Content-Type**: Specifies request body format (application/json)

All responses are currently in JSON format, with potential for additional formats in future versions.

**Section sources**
- [server.go](file://pkg/admin/server.go#L1-L20)
- [httpserver.go](file://plugin/apiserver/httpserver/server.go#L1-L50)

## Internationalization (i18n) Support
The API supports multiple languages through request headers.

### Language Header
Clients can specify their preferred language using the `Accept-Language` header:
```
Accept-Language: en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7
```

### Supported Languages
- **en**: English
- **zh**: Chinese (Simplified)

### Response Localization
Error messages and informational responses are localized based on the `Accept-Language` header. The server selects the best match from available translations.

### Language Fallback
If a requested language is not available, the server falls back to:
1. Exact language match
2. Language with any region (e.g., en-US → en)
3. Default language (English)

**Section sources**
- [translate.go](file://plugin/apiserver/httpserver/i18n/translate.go#L1-L50)
- [en.toml](file://deploy/conf/i18n/en.toml#L1-L100)
- [zh.toml](file://deploy/conf/i18n/zh.toml#L1-L100)

## Client Examples
Example implementations for common operations.

### Go Client: Service Registration
```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type Instance struct {
    ID        string            `json:"id,omitempty"`
    Service   string            `json:"service"`
    Namespace string            `json:"namespace"`
    Host      string            `json:"host"`
    Port      int               `json:"port"`
    Protocol  string            `json:"protocol,omitempty"`
    Version   string            `json:"version,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

func registerService() error {
    instance := Instance{
        Service:   "my-service",
        Namespace: "default",
        Host:      "192.168.1.100",
        Port:      8080,
        Protocol:  "http",
    }
    
    jsonData, _ := json.Marshal(map[string]Instance{
        "instance": instance,
    })
    
    resp, err := http.Post(
        "http://localhost:8080/naming/v1/instance",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### Go Client: Configuration Retrieval
```go
package main

import (
    "fmt"
    "net/http"
    "net/url"
)

func getConfig() error {
    baseURL := "http://localhost:8080/config/v1/client_config"
    
    params := url.Values{}
    params.Add("namespace", "default")
    params.Add("group", "app")
    params.Add("fileName", "application.yaml")
    params.Add("version", "0")
    
    resp, err := http.Get(baseURL + "?" + params.Encode())
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 200 {
        fmt.Println("Configuration retrieved successfully")
    }
    
    return nil
}
```

### Go Client: Authentication
```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code         int    `json:"code"`
    Info         string `json:"info"`
    LoginResponse struct {
        User  map[string]string `json:"user"`
        Token string            `json:"token"`
        TTL   int               `json:"ttl"`
    } `json:"loginResponse"`
}

func authenticate() (string, error) {
    loginReq := LoginRequest{
        Username: "admin",
        Password: "admin",
    }
    
    jsonData, _ := json.Marshal(loginReq)
    
    resp, err := http.Post(
        "http://localhost:8080/auth/v1/login",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var loginResp LoginResponse
    // JSON decoding would be here in actual implementation
    
    return loginResp.LoginResponse.Token, nil
}
```

**Section sources**
- [client_v1.go](file://pkg/service/client_v1.go#L1-L200)
- [config_file.go](file://pkg/config/config_file.go#L1-L200)
- [auth.go](file://apis/access_control/auth/auth.go#L1-L200)

## Security Considerations
Security best practices for using the API.

### Authentication Mechanisms
- **Token-based authentication**: Most endpoints require a valid token
- **HTTP Basic Auth**: Supported for some legacy endpoints
- **API Keys**: Configurable for programmatic access

### Secure Communication
- **HTTPS**: Recommended for all production deployments
- **TLS**: Configurable with custom certificates
- **HSTS**: Supported for enhanced security

### Authorization Model
- **Role-Based Access Control (RBAC)**: Users assigned to roles with specific permissions
- **Attribute-Based Access Control (ABAC)**: Fine-grained policies based on attributes
- **Least Privilege**: Users granted minimum necessary permissions

### Security Headers
The API includes security headers in responses:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`

**Section sources**
- [auth.go](file://apis/access_control/auth/auth.go#L1-L200)
- [auth_response.go](file://pkg/common/api/v1/auth_response.go#L1-L100)
- [security.go](file://pkg/common/secure/config.go#L1-L50)