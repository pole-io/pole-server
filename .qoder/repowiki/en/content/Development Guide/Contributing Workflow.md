# Contributing Workflow

<cite>
**Referenced Files in This Document**   
- [main.go](file://main.go)
- [cmd/root.go](file://cmd/root.go)
- [cmd/start.go](file://cmd/start.go)
- [bootstrap/server.go](file://bootstrap/server.go)
- [bootstrap/config/config.go](file://bootstrap/config/config.go)
- [plugin.go](file://plugin.go)
- [apis/plugin.go](file://apis/plugin.go)
- [test/suit/plugin.go](file://test/suit/plugin.go)
</cite>

## Table of Contents
1. [Introduction](#introduction)
2. [Development Environment Setup](#development-environment-setup)
3. [Forking and Branching Strategy](#forking-and-branching-strategy)
4. [Pull Request Structure and Commit Conventions](#pull-request-structure-and-commit-conventions)
5. [Plugin System Implementation](#plugin-system-implementation)
6. [Documentation and Testing Requirements](#documentation-and-testing-requirements)
7. [Backward Compatibility Guidelines](#backward-compatibility-guidelines)
8. [Issue Reporting and Feature Proposals](#issue-reporting-and-feature-proposals)
9. [Community Participation](#community-participation)
10. [Conclusion](#conclusion)

## Introduction
This document outlines the complete contribution workflow for the pole-server project, a cloud-native service governance platform with AI-native capabilities. The contribution process covers development environment setup, code contribution guidelines, plugin system implementation, and community engagement practices. The system supports various protocols including MCP, Apollo, Nacos, and XDS, with extensible plugin architecture for observability, access control, and storage components.

## Development Environment Setup

To set up a local development environment for pole-server, follow these steps:

1. **Prerequisites**: Ensure Go 1.19+ and make are installed on your system
2. **Clone the Repository**: After forking, clone your fork locally
3. **Build the Binary**: Use the Makefile to build the server
4. **Configuration**: Copy and modify the example configuration files from deploy/conf/
5. **Run the Server**: Start the server using the provided scripts

The server can be built and started using the following commands:
```bash
make build
./deploy/tools/start.sh
```

The bootstrap process loads configuration from YAML files, initializes components in a specific order to prevent database overload, and supports dynamic API server loading based on configuration.

**Section sources**
- [main.go](file://main.go#L1-L28)
- [cmd/root.go](file://cmd/root.go#L1-L43)
- [cmd/start.go](file://cmd/start.go#L1-L42)
- [bootstrap/server.go](file://bootstrap/server.go#L1-L716)
- [bootstrap/config/config.go](file://bootstrap/config/config.go#L1-L126)

## Forking and Branching Strategy

The recommended workflow for contributing to pole-server involves the following steps:

1. **Fork the Repository**: Create a personal fork of the pole-io/pole-server repository
2. **Create a Feature Branch**: Branch from the main branch using a descriptive name
3. **Implement Changes**: Make your modifications on the feature branch
4. **Sync with Upstream**: Regularly rebase your branch with upstream main
5. **Push and Create PR**: Push your branch and create a pull request

Feature branches should follow the naming convention: `feature/descriptive-name`, `bugfix/issue-description`, or `docs/update-topic`. The system uses a modular architecture where components like service discovery, configuration management, and governance rules are initialized separately during the bootstrap process, allowing for isolated development and testing of specific features.

**Section sources**
- [main.go](file://main.go#L1-L28)
- [bootstrap/server.go](file://bootstrap/server.go#L1-L716)

## Pull Request Structure and Commit Conventions

Pull requests to pole-server should follow a structured format:

- **Title**: Clear and descriptive, following conventional commits format
- **Description**: Detailed explanation of changes, motivation, and implementation approach
- **Related Issues**: Reference to any associated GitHub issues
- **Testing Information**: Description of test cases and verification steps
- **Documentation Updates**: List of documentation changes

Commit messages must follow the conventional commits format:
- `feat: add new feature`
- `fix: resolve issue`
- `docs: update documentation`
- `style: format code`
- `refactor: improve code structure`
- `test: add test cases`
- `chore: maintenance tasks`

Each PR should focus on a single concern and include appropriate tests. The codebase uses a component-based architecture where services like naming, configuration, and governance rules are initialized with specific options, and changes should maintain this modular structure.

**Section sources**
- [bootstrap/server.go](file://bootstrap/server.go#L1-L716)
- [bootstrap/config/config.go](file://bootstrap/config/config.go#L1-L126)

## Plugin System Implementation

The pole-server plugin system provides an extensible architecture for adding functionality. Plugins are defined in the `test/suit/plugin.go` file and registered through Go's init mechanism.

### Plugin Interface
The plugin system is defined by the Plugin interface in `apis/plugin.go`:
```go
type Plugin interface {
    Name() string
    Initialize(c *ConfigEntry) error
    Destroy() error
    Type() PluginType
}
```

### Adding New Plugins
To add a new plugin:

1. **Implement the Plugin Interface**: Create a new package that implements the Plugin interface
2. **Register the Plugin**: Add the plugin to the appropriate plugin.go file using import with underscore
3. **Configure the Plugin**: Define configuration in the YAML configuration file
4. **Initialize in Bootstrap**: Ensure the plugin is loaded during server startup

Plugins are categorized by type (PluginType) including:
- PluginTypeStatis: Statistics collection
- PluginTypeHistory: History recording
- PluginTypeDiscoverEvent: Discovery event handling
- PluginTypeRateLimit: Rate limiting
- PluginTypeWhitelist: IP whitelisting
- PluginTypeResourceAuth: Resource authentication
- PluginTypeCMDB: Configuration management database
- PluginTypeApiServer: API server implementations
- PluginTypeCrypto: Cryptographic operations
- PluginTypeStore: Storage backends

The plugin system uses a registration mechanism where plugins are registered via imports in `plugin.go`, and configuration is passed through the ConfigEntry structure with name and option fields.

```mermaid
classDiagram
class Plugin {
+Name() string
+Initialize(c *ConfigEntry) error
+Destroy() error
+Type() PluginType
}
class ConfigEntry {
+Name string
+Option map[string]interface{}
}
class PluginType {
+PluginTypeStatis
+PluginTypeHistory
+PluginTypeDiscoverEvent
+PluginTypeRateLimit
+PluginTypeWhitelist
+PluginTypeResourceAuth
+PluginTypeCMDB
+PluginTypeApiServer
+PluginTypeCrypto
+PluginTypeStore
+PluginTypeHealthCheck
}
class Config {
+CMDB ConfigEntry
+RateLimit ConfigEntry
+History PluginChanConfig
+Statis PluginChanConfig
+DiscoverStatis ConfigEntry
+ParsePassword ConfigEntry
+Whitelist ConfigEntry
+MeshResourceValidate ConfigEntry
+DiscoverEvent PluginChanConfig
+Crypto PluginChanConfig
}
Plugin <|-- RateLimitPlugin
Plugin <|-- AuthPlugin
Plugin <|-- StoragePlugin
ConfigEntry "1" -- "0..*" Plugin : uses
Plugin --> PluginType : has type
Config "1" -- "0..*" ConfigEntry : contains
```

**Diagram sources**
- [apis/plugin.go](file://apis/plugin.go#L1-L112)
- [plugin.go](file://plugin.go#L1-L54)
- [test/suit/plugin.go](file://test/suit/plugin.go#L1-L34)

**Section sources**
- [apis/plugin.go](file://apis/plugin.go#L1-L112)
- [plugin.go](file://plugin.go#L1-L54)
- [test/suit/plugin.go](file://test/suit/plugin.go#L1-L34)

## Documentation and Testing Requirements

All contributions must include appropriate documentation and testing:

### Documentation Requirements
- **Code Comments**: All exported functions and types must have Go doc comments
- **Configuration Documentation**: New configuration options must be documented
- **API Documentation**: Changes to APIs must be reflected in documentation
- **Plugin Documentation**: New plugins require usage documentation

### Testing Requirements
The project requires comprehensive test coverage:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Verify component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Benchmark Tests**: Measure performance characteristics

The test suite is organized in the `test/` directory with subdirectories for different test types:
- `integrate/`: Integration and end-to-end tests
- `benchmark/`: Performance benchmarks
- `suit/`: Test suite utilities

Plugins must include both unit tests and integration tests to ensure they work correctly within the system. The bootstrap process includes a test mode that can be used to verify plugin functionality during development.

**Section sources**
- [test/suit/plugin.go](file://test/suit/plugin.go#L1-L34)
- [bootstrap/server.go](file://bootstrap/server.go#L1-L716)

## Backward Compatibility Guidelines

Maintaining backward compatibility is critical for pole-server as it serves as a production-ready service governance platform. The following guidelines apply:

### API Compatibility
- **Breaking Changes**: Must be avoided in patch releases
- **Deprecation Process**: Mark features as deprecated before removal
- **Versioning**: Follow semantic versioning principles
- **Migration Paths**: Provide clear upgrade instructions

### Configuration Compatibility
- **New Options**: Can be added without breaking changes
- **Removed Options**: Must go through deprecation cycle
- **Default Values**: Should not change behavior unexpectedly
- **File Format**: YAML structure changes require versioning

### Plugin Compatibility
- **Interface Stability**: Plugin interfaces should remain stable
- **Configuration Migration**: Provide tools for configuration updates
- **Graceful Degradation**: System should function with missing plugins

The system uses a modular initialization process where components are initialized with specific options, allowing for gradual updates and compatibility layers. The bootstrap process includes version detection and can adapt behavior based on configuration version.

**Section sources**
- [bootstrap/server.go](file://bootstrap/server.go#L1-L716)
- [bootstrap/config/config.go](file://bootstrap/config/config.go#L1-L126)

## Issue Reporting and Feature Proposals

### Issue Reporting
When reporting issues, include:
- **Environment Information**: OS, Go version, pole-server version
- **Reproduction Steps**: Clear steps to reproduce the issue
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Logs and Errors**: Relevant log output and error messages
- **Configuration**: Relevant configuration snippets (with sensitive data removed)

### Feature Proposals
To propose new features:
1. **Create an Issue**: Describe the problem and proposed solution
2. **Discuss Design**: Engage with maintainers on implementation approach
3. **Provide Use Cases**: Explain the scenarios the feature enables
4. **Consider Alternatives**: Discuss other possible solutions
5. **Implementation Plan**: Outline the proposed changes

The project roadmap includes AI-native features like MCP protocol support, enhanced service discovery, and improved configuration management. Proposals should align with these strategic directions when possible.

**Section sources**
- [README.md](file://README.md#L1-L31)

## Community Participation

The pole-server community welcomes contributions through various channels:

### Contribution Areas
- **Code Contributions**: New features, bug fixes, performance improvements
- **Documentation**: Tutorials, examples, API references
- **Testing**: Bug reports, test case development
- **Community Support**: Helping other users, answering questions
- **Tooling**: Development tools, deployment scripts

### Communication Channels
- **GitHub Issues**: For bug reports and feature requests
- **Pull Requests**: For code contributions
- **Discussions**: For broader topics and questions

### Code Review Process
The code review process includes:
1. **Automated Checks**: CI/CD pipeline validation
2. **Technical Review**: Architecture and implementation assessment
3. **Style Review**: Code formatting and conventions
4. **Documentation Review**: Adequacy of documentation
5. **Test Review**: Sufficient test coverage

Maintainers will provide timely feedback on contributions, and contributors are expected to respond to review comments promptly. The project values clear communication and collaborative problem-solving.

**Section sources**
- [README.md](file://README.md#L1-L31)
- [Makefile](file://Makefile#L1-L64)

## Conclusion
Contributing to pole-server involves following a structured workflow that ensures code quality, maintainability, and backward compatibility. The plugin-based architecture allows for extensible functionality while maintaining a clean core system. By following the guidelines outlined in this document, contributors can effectively participate in the development of this cloud-native service governance platform. The combination of clear contribution processes, comprehensive testing requirements, and active community engagement ensures the continued success and evolution of the project.