# Getting Started

<cite>
**Referenced Files in This Document**   
- [pole-server.yaml](file://deploy/conf/pole-server.yaml)
- [start.sh](file://deploy/tools/start.sh)
- [Makefile](file://Makefile)
- [version.go](file://cmd/version.go)
- [client_access.go](file://plugin/apiserver/httpserver/discover/client_access.go)
</cite>

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Installation Methods](#installation-methods)
3. [Configuration](#configuration)
4. [Starting the Server](#starting-the-server)
5. [Verification and Health Check](#verification-and-health-check)
6. [Service Registration and Discovery](#service-registration-and-discovery)
7. [Troubleshooting Common Issues](#troubleshooting-common-issues)

## Prerequisites
Before installing pole-server, ensure your environment meets the following requirements:
- Go version 1.19 or higher
- MySQL database server accessible with proper credentials
- Network configuration allowing communication on required ports
- Sufficient file descriptor limits for high-concurrency scenarios

**Section sources**
- [pole-server.yaml](file://deploy/conf/pole-server.yaml#L1-L172)

## Installation Methods
pole-server can be installed through three primary methods: binary distribution, Docker container, or source compilation.

### Binary Installation
Download the pre-compiled binary package from the official release channel and extract it to your desired installation directory. The binary includes all necessary dependencies and configuration files.

### Docker Installation
Use the provided Dockerfile in the repository root to build a container image:
```bash
make build-docker
```
Alternatively, pull the official image from the container registry and run with appropriate environment variables for database connection.

### Source Compilation
Compile pole-server from source using the provided Makefile:
```bash
make build
```
This command compiles the binary for your current architecture. For cross-compilation, specify the ARCH parameter:
```bash
make build ARCH=arm64
```

**Section sources**
- [Makefile](file://Makefile#L0-L64)

## Configuration
The main configuration file `pole-server.yaml` controls server behavior and plugin settings. Key configuration sections include:

### Server Address and Logging
Configure the bootstrap section to set up logging and startup order:
```yaml
bootstrap:
  logger: ./conf/pole-log.yaml
  startInOrder:
    open: true
    key: sz
```

### Database Configuration
Set up MySQL connection in the store section:
```yaml
store:
  name: defaultStore
  option:
    master:
      dbType: mysql
      dns: "${MYSQL_USER}:${MYSQL_PWD}@tcp(${MYSQL_HOST})/pole_server?loc=Local"
      maxOpenConns: 300
      maxIdleConns: 50
      connMaxLifetime: 300
```

### Plugin Settings
Enable and configure plugins in the plugin section:
```yaml
plugin:
  crypto:
    entries:
      - name: AES
  cmdb:
    name: memory
  statis:
    entries:
      - name: local
        option:
          interval: 60
      - name: prometheus
```

**Section sources**
- [pole-server.yaml](file://deploy/conf/pole-server.yaml#L1-L172)

## Starting the Server
Start pole-server using either the startup script or direct binary execution.

### Using Startup Script
Execute the provided start.sh script:
```bash
./deploy/tools/start.sh
```
The script checks if the server is already running and starts it if not. Use the --enable-cron flag to enable periodic health checks.

### Direct Binary Execution
Run the compiled binary directly:
```bash
./pole-server
```
Ensure all environment variables (like MYSQL_USER, MYSQL_PWD, MYSQL_HOST) are set before execution.

**Section sources**
- [start.sh](file://deploy/tools/start.sh#L0-L52)

## Verification and Health Check
Verify pole-server operation through version command and health check endpoint.

### Version Verification
Check the installed version:
```bash
./pole-server version
```
This command outputs the current version information. The version command is implemented in the cmd package.

### Health Check Endpoint
Access the health check endpoint at `/Discover` to verify server status. The server responds with service instance information when operational.

**Section sources**
- [version.go](file://cmd/version.go#L0-L46)

## Service Registration and Discovery
Interact with pole-server using HTTP API for service registration and discovery.

### Registering a Service Instance
Use the RegisterInstance endpoint to register a service:
```bash
curl -X POST http://localhost:8080/RegisterInstance -d '{
  "service": "my-service",
  "namespace": "default",
  "host": "192.168.1.100",
  "port": 8081
}'
```
The registration request is processed by the HTTPServer's RegisterInstance method.

### Discovering Services
Use the Discover endpoint to find available services:
```bash
curl -X POST http://localhost:8080/Discover -d '{
  "type": "INSTANCE",
  "service": {
    "name": "my-service",
    "namespace": "default"
  }
}'
```

### Heartbeat Reporting
Maintain service registration with periodic heartbeat:
```bash
curl -X POST http://localhost:8080/Heartbeat -d '{
  "service": "my-service",
  "namespace": "default",
  "host": "192.168.1.100",
  "port": 8081
}'
```

**Section sources**
- [client_access.go](file://plugin/apiserver/httpserver/discover/client_access.go#L37-L186)

## Troubleshooting Common Issues
Address common startup and operation problems with these solutions.

### Port Conflicts
If the server fails to start due to port conflicts:
1. Check which process is using the port: `lsof -i :8080`
2. Either terminate the conflicting process or modify the server configuration to use a different port

### Database Connection Issues
For database connection failures:
1. Verify MySQL server accessibility from the pole-server host
2. Check database credentials in the dns configuration string
3. Ensure the pole_server database exists and has proper permissions
4. Validate network connectivity and firewall rules

### Configuration Validation
Always validate configuration changes before restarting:
1. Check YAML syntax with a validator
2. Ensure all environment variables are properly set
3. Verify file paths in the configuration are correct

**Section sources**
- [pole-server.yaml](file://deploy/conf/pole-server.yaml#L1-L172)
- [start.sh](file://deploy/tools/start.sh#L0-L52)