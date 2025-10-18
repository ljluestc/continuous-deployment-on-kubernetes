# Testing Guide for gceme Application

## Table of Contents
1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Test Structure](#test-structure)
4. [Running Tests](#running-tests)
5. [Coverage Reports](#coverage-reports)
6. [Writing Tests](#writing-tests)
7. [Troubleshooting](#troubleshooting)

## Overview

The gceme application has comprehensive test coverage including:
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test component interactions and end-to-end flows
- **Benchmark Tests**: Performance testing and optimization
- **Coverage Target**: 100% code coverage

### Test Statistics
- Total Test Files: 5
- Total Tests: 100+
- Coverage Target: 100%
- Test Execution Time: <5 seconds

## Quick Start

### Run All Tests
```bash
cd sample-app
go test -v ./...
```

### Run Tests with Coverage
```bash
cd sample-app
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Run with Race Detector
```bash
cd sample-app
go test -race ./...
```

### Run Using Python Orchestrator
```bash
python3 test_comprehensive.py --all
```

## Test Structure

### File Organization
```
sample-app/
├── main.go                 # Application code
├── html.go                 # HTML template
├── main_test.go            # Main function tests (existing)
├── backend_test.go         # Backend mode tests (new)
├── frontend_test.go        # Frontend mode tests (new)
├── instance_test.go        # Instance and metadata tests (new)
├── html_test.go            # Template tests (new)
└── integration_test.go     # Integration tests (new)
```

### Test Categories

#### 1. Backend Tests (`backend_test.go`)
Tests for backend mode functionality:
- HTTP endpoint handlers
- JSON marshaling/unmarshaling
- Health checks
- Version endpoints
- Concurrent request handling
- Large payload handling

#### 2. Frontend Tests (`frontend_test.go`)
Tests for frontend mode functionality:
- HTML template rendering
- Backend communication
- Error handling (backend unavailable, invalid JSON)
- Health check propagation
- XSS prevention
- Concurrent requests

#### 3. Instance Tests (`instance_test.go`)
Tests for instance metadata handling:
- GCE metadata retrieval
- Non-GCE environment handling
- Assigner helper functionality
- Error propagation
- Field population

#### 4. HTML Tests (`html_test.go`)
Tests for HTML template:
- Template parsing
- Template execution
- Data rendering
- HTML escaping
- Placeholder validation
- CSS/layout validation

#### 5. Integration Tests (`integration_test.go`)
End-to-end tests:
- Frontend-backend communication
- Multi-instance scenarios
- Load balancing simulation
- Failover handling
- Version consistency
- Stress testing

## Running Tests

### Basic Test Commands

#### Run All Tests
```bash
go test ./...
```

#### Run Specific Test File
```bash
go test -v -run TestBackendMode backend_test.go main.go html.go
```

#### Run Specific Test Function
```bash
go test -v -run TestBackendMode_RootEndpoint
```

#### Run Tests Matching Pattern
```bash
go test -v -run "Backend.*JSON"
```

### Coverage Commands

#### Generate Coverage Profile
```bash
go test -coverprofile=coverage.out ./...
```

#### View Coverage in Terminal
```bash
go tool cover -func=coverage.out
```

#### Generate HTML Coverage Report
```bash
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

#### Check Coverage by Package
```bash
go test -cover ./...
```

### Advanced Testing

#### Race Detection
```bash
go test -race ./...
```

#### Integration Tests Only
```bash
go test -v -tags=integration ./...
```

#### Short Mode (Skip Long Tests)
```bash
go test -short ./...
```

#### Verbose Output
```bash
go test -v ./...
```

#### Run Tests Multiple Times
```bash
go test -count=10 ./...
```

### Benchmark Tests

#### Run All Benchmarks
```bash
go test -bench=. ./...
```

#### Run Specific Benchmark
```bash
go test -bench=BenchmarkBackendMode_RootEndpoint
```

#### Benchmark with Memory Stats
```bash
go test -bench=. -benchmem ./...
```

#### Longer Benchmark Runs
```bash
go test -bench=. -benchtime=10s ./...
```

#### CPU Profiling
```bash
go test -bench=. -cpuprofile=cpu.out ./...
go tool pprof cpu.out
```

#### Memory Profiling
```bash
go test -bench=. -memprofile=mem.out ./...
go tool pprof mem.out
```

## Coverage Reports

### Understanding Coverage Output

```
cloud.google.com/go/compute/metadata/metadata.go:0.0%
sample-app/html.go:100.0%
sample-app/main.go:95.2%
total:(statements)92.5%
```

### Coverage Goals
- **main.go**: 95%+ (some metadata code not coverable locally)
- **html.go**: 100%
- **helpers**: 100%
- **handlers**: 100%
- **Overall**: 90%+ (excluding vendor and metadata API)

### Generating Reports

#### Text Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out > coverage.txt
```

#### HTML Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### JSON Report (for CI/CD)
```bash
go test -coverprofile=coverage.out -json ./... > test-results.json
```

## Writing Tests

### Test Naming Convention
```go
// Format: Test<Component>_<Scenario>_<ExpectedOutcome>
func TestBackendMode_RootEndpoint_ReturnsJSON(t *testing.T) {}
func TestFrontendMode_BackendUnavailable_ReturnsError(t *testing.T) {}
```

### Test Structure (AAA Pattern)
```go
func TestExample(t *testing.T) {
    // Arrange - Set up test data
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()

    // Act - Execute the code being tested
    handler.ServeHTTP(w, req)

    // Assert - Verify the results
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

### Using Table-Driven Tests
```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"scenario1", "input1", "output1"},
        {"scenario2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := function(tt.input)
            if result != tt.expected {
                t.Errorf("Expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

### Using Subtests
```go
func TestHTTPHandlers(t *testing.T) {
    t.Run("health", func(t *testing.T) {
        // Test health endpoint
    })

    t.Run("version", func(t *testing.T) {
        // Test version endpoint
    })
}
```

### Parallel Tests
```go
func TestParallel(t *testing.T) {
    t.Parallel() // Run in parallel with other parallel tests

    // Test code
}
```

### HTTP Testing with httptest
```go
func TestHTTPHandler(t *testing.T) {
    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(handler))
    defer server.Close()

    // Make request
    resp, err := http.Get(server.URL)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    // Verify response
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }
}
```

## Troubleshooting

### Common Issues

#### Tests Fail Due to Metadata API
**Problem**: Tests fail when not running on GCE
**Solution**: Tests already handle this with `metadata.OnGCE()` check

#### Race Condition Detected
**Problem**: `go test -race` reports data races
**Solution**: Use proper synchronization (mutexes, channels)

#### Coverage Not 100%
**Problem**: Some code is not covered
**Solution**:
- Run `go tool cover -html=coverage.out`
- Add tests for uncovered lines
- Some GCE metadata code cannot be covered locally (acceptable)

#### Tests Timeout
**Problem**: Tests hang or timeout
**Solution**:
- Increase timeout: `go test -timeout=30s`
- Check for deadlocks
- Use `t.Parallel()` carefully

#### Import Cycle
**Problem**: Import cycle when adding tests
**Solution**: Keep tests in same package, use `package main` not `package main_test`

### Debugging Tests

#### Run Single Test with Verbose Output
```bash
go test -v -run TestSpecificTest
```

#### Print Debug Information
```go
func TestDebug(t *testing.T) {
    t.Logf("Debug info: %v", someVariable)
    // Use t.Log or t.Logf instead of fmt.Print
}
```

#### Skip Tests Temporarily
```go
func TestSkipped(t *testing.T) {
    t.Skip("Skipping this test temporarily")
}
```

#### Conditional Test Execution
```go
func TestConditional(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
    // Long-running test
}
```

## Best Practices

### DO
✅ Write tests for all new code
✅ Use table-driven tests for multiple scenarios
✅ Test error cases and edge cases
✅ Use meaningful test names
✅ Keep tests independent
✅ Use `t.Parallel()` when appropriate
✅ Clean up resources (defer close)
✅ Use httptest for HTTP testing

### DON'T
❌ Skip tests without good reason
❌ Write flaky tests (timing-dependent)
❌ Test implementation details
❌ Use time.Sleep in tests
❌ Ignore test failures
❌ Write tests that depend on order
❌ Leave commented-out test code

## CI/CD Integration

Tests are automatically run in CI/CD pipeline:
1. On every push to main branches
2. On every pull request
3. Coverage must meet threshold (90%+)
4. All tests must pass before merge

See [CI_CD.md](CI_CD.md) for more details.

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)
- [httptest Package](https://golang.org/pkg/net/http/httptest/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

---
**Last Updated**: 2025-10-17
**Maintainer**: Engineering Team
