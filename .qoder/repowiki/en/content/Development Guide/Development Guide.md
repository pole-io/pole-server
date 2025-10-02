# Development Guide

<cite>
**Referenced Files in This Document**   
- [Makefile](file://Makefile)
- [import-format.sh](file://import-format.sh)
- [client_grpc_test.go](file://test/integrate/client_grpc_test.go)
- [discover_test.go](file://test/benchmark/grpc/client/discover_test.go)
</cite>

## Table of Contents
1. [Build Process](#build-process)
2. [Coding Standards and Code Formatting](#coding-standards-and-code-formatting)
3. [Testing Strategies](#testing-strategies)
4. [Writing New Features and Tests](#writing-new-features-and-tests)
5. [Debugging and Performance Testing](#debugging-and-performance-testing)
6. [Contribution Workflow](#contribution-workflow)

## Build Process

The build process for pole-server is managed through the Makefile, which provides several targets for compilation, cleaning, and Docker image building. The default target `all` invokes the `build` target, which executes the release script to generate binaries and tarballs.

Key Makefile targets include:
- `build`: Compiles the binary and creates a tarball using the version from the version file and specified architecture
- `build-docker`: Builds Docker images with the specified image tag
- `clean`: Removes generated build artifacts and release directories
- `help`: Displays all available targets with descriptions

The build system supports version overrides through environment variables or command-line arguments, allowing for flexible version management during development and release cycles.

**Section sources**
- [Makefile](file://Makefile#L1-L64)

## Coding Standards and Code Formatting

Code formatting and import organization are standardized using the `import-format.sh` script. This script ensures consistent code style across the codebase by:

1. Running `go mod tidy -compat=1.17` to clean and organize module dependencies
2. Executing `go fmt ./...` to format all Go files according to standard conventions
3. Using `goimports-reviser` to organize imports with specific project and company prefixes

The script automatically downloads the appropriate version of `goimports-reviser` based on the operating system and architecture. It processes all `.go` files except generated protobuf files (`.pb.go`), test tools, and the main plugin file.

Import organization follows these rules:
- Removes unused imports
- Applies company prefixes for external packages (`github.com/pole-io/specification`)
- Uses project name for internal package organization (`github.com/pole-io/pole-server`)

**Section sources**
- [import-format.sh](file://import-format.sh#L1-L48)

## Testing Strategies

The pole-server project implements a comprehensive testing strategy with multiple testing layers:

### Unit Tests
Unit tests are implemented throughout the codebase using Go's testing framework. They focus on testing individual functions and methods in isolation, ensuring that each component works correctly on its own.

### Integration Tests
Integration tests are located in the `test/integrate` directory and verify the interaction between multiple components. These tests use both HTTP and gRPC clients to test the server's API endpoints. The integration test suite includes:
- Service discovery tests
- Configuration management tests
- Instance registration and heartbeat tests
- Namespace and service management tests

Example integration test structure:
```go
func TestClientGRPC_DiscoverInstance(t *testing.T) {
    // Test setup
    DiscoveryRunAndInitResource(t, func(t *testing.T, clientHttp *http.Client, namespaces []*apimodel.Namespace, services []*apiservice.Service) {
        // Test cases
        t.Run("GRPC——上报SDK客户端信息", func(t *testing.T) {
            // Test implementation
        })
    })
}
```

### Benchmarking
Performance benchmarks are located in the `test/benchmark` directory, specifically focusing on gRPC operations. The benchmark suite includes:
- Service discovery performance testing
- Load testing for various API operations

The benchmarking framework allows for performance measurement under different conditions and can be configured via environment variables to target specific server instances.

**Section sources**
- [client_grpc_test.go](file://test/integrate/client_grpc_test.go#L1-L224)
- [discover_test.go](file://test/benchmark/grpc/client/discover_test.go#L1-L113)

## Writing New Features and Tests

When adding new features to pole-server, follow these guidelines:

1. **Feature Implementation**:
   - Place new code in the appropriate package based on functionality
   - Follow existing code patterns and design principles
   - Ensure proper error handling and logging
   - Document public APIs and complex logic

2. **Test Creation**:
   - Write unit tests for all new functions and methods
   - Create integration tests for API endpoints and cross-component interactions
   - Add benchmarks for performance-critical operations
   - Use table-driven tests for comprehensive coverage

3. **Code Organization**:
   - Follow the existing directory structure
   - Use descriptive function and variable names
   - Maintain consistent formatting using the provided scripts

4. **Documentation**:
   - Update relevant documentation
   - Add comments for complex algorithms or business logic
   - Ensure API documentation is complete and accurate

## Debugging and Performance Testing

### Debugging Techniques
- Use Go's built-in debugging tools and logging
- Leverage the integration test framework to reproduce issues
- Utilize the benchmark suite to identify performance bottlenecks
- Enable detailed logging for specific components when needed

### Profiling Methods
The project supports various profiling methods:
- CPU profiling to identify computational bottlenecks
- Memory profiling to detect leaks and optimize allocations
- Block profiling to analyze goroutine contention
- Mutex profiling to identify locking issues

### Performance Testing Procedures
1. Set up the benchmark environment using environment variables:
   - `BENCHMARK_SERVER_ADDRESS`: Target server address for gRPC benchmarks
   - `BENCHMARK_SERVER_HTTP_ADDRESS`: Target server address for HTTP operations

2. Run benchmarks using the standard Go benchmark framework:
   ```bash
   go test -bench=Benchmark_DiscoverServicesWithoutRevision -run=^$ ./test/benchmark/grpc/client/
   ```

3. Analyze results and compare against baseline performance metrics

4. Optimize code based on profiling data and re-run benchmarks to verify improvements

**Section sources**
- [discover_test.go](file://test/benchmark/grpc/client/discover_test.go#L1-L113)

## Contribution Workflow

### Development Setup
1. Clone the repository
2. Install required dependencies
3. Set up the development environment

### Code Review Expectations
- All code must pass formatting checks using `import-format.sh`
- Comprehensive test coverage is required for new features
- Documentation must be updated for API changes
- Code should follow existing patterns and conventions
- Performance implications should be considered and documented

### Pull Request Process
1. Create a feature branch from the main branch
2. Implement the feature with appropriate tests
3. Run `import-format.sh` to ensure code style consistency
4. Verify all tests pass
5. Submit a pull request with a clear description of changes
6. Address any feedback from code reviewers
7. Merge after approval

### Release Process
The release process is automated through the Makefile and release scripts:
1. Update the version file
2. Run `make build` to create the release artifacts
3. Run `make build-docker` to create Docker images
4. Verify the build artifacts
5. Tag the release in the repository
6. Publish the release artifacts

**Section sources**
- [Makefile](file://Makefile#L1-L64)
- [import-format.sh](file://import-format.sh#L1-L48)