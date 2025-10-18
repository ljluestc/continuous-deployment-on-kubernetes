# Product Requirements Document (PRD)
# Continuous Deployment Testing Infrastructure

## Executive Summary
This document outlines the comprehensive testing infrastructure for the gceme (Google Container Engine Metadata Example) application, a demonstration application for continuous deployment on Kubernetes with Jenkins.

## Project Overview
- **Project Name**: gceme - Continuous Deployment Demo Application
- **Technology Stack**: Go, Kubernetes, Jenkins, Docker
- **Current State**: Basic test coverage (~10%)
- **Target State**: 100% test coverage with comprehensive CI/CD pipeline

## Objectives
1. Achieve 100% unit test coverage for all Go code
2. Implement integration tests for all system components
3. Create performance benchmarks for critical paths
4. Establish automated CI/CD pipeline with pre-commit hooks
5. Generate comprehensive test reports and coverage metrics

## Application Architecture

### Components
1. **Backend Mode** (main.go:72-86)
   - Serves GCE instance metadata as JSON
   - Endpoints: `/`, `/healthz`, `/version`
   - Port: 8080 (default)

2. **Frontend Mode** (main.go:88-137)
   - Queries backend service
   - Renders HTML UI with instance metadata
   - Endpoints: `/`, `/healthz`
   - Port: 8080 (default)

3. **Instance Metadata Handler** (main.go:154-175)
   - Fetches GCE metadata
   - Handles non-GCE environments gracefully

### Key Functions to Test
- `main()` - Application entry point with flag parsing
- `backendMode(port int)` - Backend server initialization
- `frontendMode(port int, backendURL string)` - Frontend server initialization
- `newInstance() *Instance` - Instance metadata retrieval
- `(a *assigner) assign(getVal func() (string, error)) string` - Error handling helper

## Testing Strategy

### 1. Unit Tests (100% Coverage Target)
#### Test Files
- `sample-app/main_test.go` (existing - expand)
- `sample-app/backend_test.go` (new)
- `sample-app/frontend_test.go` (new)
- `sample-app/instance_test.go` (new)
- `sample-app/html_test.go` (new)

#### Test Scenarios
1. **Flag Parsing Tests**
   - Version flag display
   - Frontend/backend mode selection
   - Port configuration
   - Backend service URL configuration

2. **Backend Mode Tests**
   - HTTP server initialization
   - Root endpoint response (JSON format)
   - Health check endpoint
   - Version endpoint
   - Request metadata capture

3. **Frontend Mode Tests**
   - HTTP server initialization
   - Backend communication
   - Template rendering
   - Error handling (backend unavailable)
   - Health check with backend verification

4. **Instance Metadata Tests**
   - GCE environment detection
   - Metadata field retrieval
   - Error handling for metadata failures
   - Non-GCE environment behavior

5. **Assigner Helper Tests**
   - Successful value assignment
   - Error propagation
   - Multiple assignment scenarios

6. **HTML Template Tests**
   - Template parsing
   - Template rendering with data
   - HTML output validation

### 2. Integration Tests
#### Scenarios
1. **End-to-End Frontend/Backend Communication**
   - Start backend server
   - Start frontend server
   - Verify frontend queries backend
   - Verify data flow

2. **Load Balancer Simulation**
   - Multiple backend instances
   - Frontend routing
   - Health check validation

3. **Kubernetes Deployment Simulation**
   - Container environment testing
   - Service discovery
   - Network communication

### 3. Benchmark Tests
#### Performance Metrics
1. **Backend Performance**
   - Request processing time
   - JSON marshaling performance
   - Concurrent request handling

2. **Frontend Performance**
   - Template rendering time
   - Backend request latency
   - Concurrent user simulation

3. **Metadata Retrieval**
   - GCE metadata API call time
   - Caching effectiveness

### 4. Edge Cases and Boundary Conditions
1. Invalid port numbers (negative, >65535)
2. Malformed backend URLs
3. Network timeouts
4. Concurrent request handling
5. Large payload handling
6. Missing GCE metadata
7. Template rendering errors
8. JSON unmarshaling errors

## Test Infrastructure

### Tools and Frameworks
1. **Go Testing Framework** (`testing` package)
2. **HTTP Testing** (`net/http/httptest`)
3. **Coverage Tool** (`go test -cover`)
4. **Benchmark Tool** (`go test -bench`)
5. **Python Test Orchestration** (custom script)
6. **Pre-commit Hooks** (go fmt, go vet, go test)
7. **GitHub Actions** (CI/CD pipeline)

### Coverage Requirements
- **Minimum Coverage**: 100% for production code
- **Excluding**: Vendor dependencies
- **Format**: HTML report + JSON report + Terminal output

## CI/CD Pipeline

### Pre-commit Hooks
1. `go fmt` - Code formatting
2. `go vet` - Static analysis
3. `go test` - Unit tests
4. Coverage threshold check (100%)
5. Linting (golint/golangci-lint)

### GitHub Actions Workflow
1. **On Push/PR**
   - Checkout code
   - Set up Go environment
   - Run tests with coverage
   - Upload coverage reports
   - Run benchmarks
   - Build Docker image
   - Security scanning

2. **On Main Branch**
   - All above steps
   - Deploy to staging
   - Integration tests
   - Performance tests
   - Deploy to production (manual approval)

## Success Criteria
1. ✅ 100% unit test coverage achieved
2. ✅ All integration tests passing
3. ✅ Benchmark baselines established
4. ✅ Pre-commit hooks enforcing quality
5. ✅ CI/CD pipeline fully automated
6. ✅ Coverage reports generated automatically
7. ✅ All edge cases tested and handled

## Deliverables
1. Comprehensive test suite (unit + integration + benchmark)
2. Python test orchestration script
3. Pre-commit hook configuration
4. GitHub Actions workflow
5. Documentation:
   - TESTING.md - Test execution guide
   - CI_CD.md - Pipeline documentation
   - TASK_MASTER.md - Task tracking and scenarios
   - Coverage reports (HTML/JSON)

## Timeline
- **Phase 1**: Unit Tests - All core functionality (Current)
- **Phase 2**: Integration Tests - Component interaction
- **Phase 3**: CI/CD Setup - Automation and hooks
- **Phase 4**: Documentation and Reports
- **Phase 5**: Validation and Coverage Verification

## Risks and Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| GCE metadata unavailable in tests | High | Mock metadata service |
| Network dependencies in tests | Medium | Use httptest for HTTP mocking |
| Race conditions | Medium | Use race detector (`go test -race`) |
| Flaky tests | Low | Implement retry logic, avoid timing dependencies |

## Appendix

### Test Naming Convention
```
Test<FunctionName>_<Scenario>_<ExpectedOutcome>
```
Examples:
- `TestNewInstance_OnGCE_ReturnsMetadata`
- `TestBackendMode_HealthEndpoint_ReturnsOK`
- `TestFrontendMode_BackendUnavailable_ReturnsError`

### Coverage Report Format
```bash
go test -coverprofile=coverage.out ./sample-app
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out
```

### Benchmark Report Format
```bash
go test -bench=. -benchmem ./sample-app > benchmark.txt
```

---
**Document Version**: 1.0
**Last Updated**: 2025-10-17
**Owner**: Engineering Team
