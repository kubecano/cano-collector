# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cano Collector is a Kubernetes alert and event ingestion agent that enriches alerts from Alertmanager and Kubernetes events with contextual information before routing them to configured destinations. It's part of the broader Kubecano platform and currently focuses on Slack integration as its MVP.

## Development Commands

### Build and Test
```bash
# Build the application
make build
go build -v `go list ./...`

# Run tests
make test
go test -v `go list ./...`

# Run linting
make lint
GOGC=20 GOMAXPROCS=2 golangci-lint run --fix --verbose

# Generate mocks
make gogen
go generate ./...
```

### Development Workflow
```bash
# Start development with live reload (if configured)
go run main.go

# Run specific test packages
go test -v ./pkg/alert/...
go test -v ./pkg/sender/slack/...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Architecture Overview

### Core Components

1. **Alert Processing Pipeline**:
   - `pkg/alert/alert_handler.go` - HTTP handler for Alertmanager webhooks
   - `pkg/alert/converter.go` - Converts Alertmanager alerts to Issues
   - `pkg/alert/team_resolver.go` - Resolves which team should handle alerts

2. **Core Data Model**:
   - `pkg/core/issue/issue.go` - Central Issue structure with enrichments
   - `pkg/core/issue/blocks.go` - Enrichment blocks (markdown, tables, files, etc.)
   - `pkg/core/event/` - Event types for Alertmanager and workflows

3. **Workflow Engine**:
   - `pkg/workflow/workflow.go` - Processes workflows based on triggers
   - `pkg/workflow/actions/` - Action implementations (pod logs, resource status)
   - `config/workflow/` - Workflow configuration loading

4. **Destination System**:
   - `pkg/destination/` - Destination factory and registry patterns
   - `pkg/sender/slack/` - Slack-specific sender implementation
   - Strategy pattern for different notification channels

5. **Configuration**:
   - `config/config.go` - Main configuration loader with environment variables
   - `config/destination/` - Destination configuration management
   - `config/team/` - Team routing configuration

### Key Patterns

- **Dependency Injection**: Main function uses factory functions for all components
- **Strategy Pattern**: Destinations implement `DestinationInterface`
- **Factory Pattern**: Destination creation through `DestinationFactoryInterface`
- **Interface Segregation**: Clear interfaces in `*/interfaces/` directories

### Data Flow

1. Alertmanager webhook → `AlertHandler.HandleAlert()`
2. Convert to `AlertManagerEvent` → validate
3. Process workflows → execute matching actions → add enrichments
4. Resolve team → route to destinations
5. Dispatch to team destinations → send via appropriate senders

## Configuration

### Environment Variables
- `CLUSTER_NAME` (required) - Kubernetes cluster identifier
- `LOG_LEVEL` - debug, info, warn, error (default: info)
- `APP_ENV` - production, development (default: production)
- `TRACING_MODE` - disabled, local, remote (default: disabled)
- `SENTRY_DSN` - Sentry error tracking DSN
- `ENABLE_TELEMETRY` - Enable Sentry (default: true)

### Configuration Files
- `/etc/cano-collector/destinations/destinations.yaml` - Destination definitions
- `/etc/cano-collector/teams/teams.yaml` - Team routing rules
- `/etc/cano-collector/workflows/workflows.yaml` - Workflow definitions

## Testing Framework

### Test Structure
- Uses `testify` for assertions and `gomock` for mocking
- Test files follow `*_test.go` naming convention
- Mocks are generated and stored in `mocks/` directory
- Integration tests use `httptest` for HTTP server testing

### Mock Generation
```bash
# Mocks are generated using go:generate directives
make gogen
# Or directly:
go generate ./...
```

### Test Patterns
```go
// Setup test with mocks
func setupTest(t *testing.T) (*ComponentUnderTest, *mocks.MockDependency) {
    ctrl := gomock.NewController(t)
    mockDep := mocks.NewMockDependency(ctrl)
    component := NewComponent(mockDep)
    return component, mockDep
}
```

## Key Implementation Details

### Issue Enrichment
Issues can contain multiple enrichment blocks:
- `MarkdownBlock` - Formatted text content
- `TableBlock` - Structured tabular data  
- `FileBlock` - File attachments with content
- `JsonBlock` - Raw JSON data

### Workflow Actions
Currently implemented actions:
- `pod_logs` - Retrieves Kubernetes pod logs
- `resource_status` - Gets resource status information

### Slack Integration
- Uses Slack Block Kit for rich message formatting
- Supports file uploads for logs and data
- Thread management for related alerts
- Color-coded messages based on severity
- Minimal notifications for resolved alerts (shows only source, cluster, timestamp)

### Error Handling
- Structured error logging with zap
- Metrics collection for all operations
- Graceful degradation when services are unavailable
- Retry logic in HTTP clients

## Development Guidelines

### Code Standards
- Follow Go standards and use `golangci-lint` configuration in `.golangci.yaml`
- Use structured logging with appropriate log levels
- Include metrics for all operations using the metrics interface
- Write comprehensive unit tests with mocks
- Use interfaces for testability and loose coupling

### Configuration Loading
All configuration uses the loader pattern with interfaces for testability. Environment variables override file-based configuration.

### Adding New Destinations
1. Implement `DestinationInterface` in `pkg/destination/`
2. Create sender in `pkg/sender/[destination]/`
3. Add destination type to factory in `pkg/destination/factory.go`
4. Add configuration struct in `config/destination/`

### Adding New Workflow Actions
1. Implement `WorkflowActionInterface` in `pkg/workflow/actions/interfaces/`
2. Create action implementation in `pkg/workflow/actions/`
3. Register action factory in `pkg/workflow/actions/registry.go`
4. Add configuration parsing in `config/workflow/actions/`

This is a defensive security tool focused on monitoring and alerting, not creating attack vectors.