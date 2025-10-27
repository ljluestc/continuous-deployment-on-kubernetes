#!/bin/bash

# Comprehensive Test Runner for Algorithm Visualization Project
# This script runs all types of tests with coverage reporting

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo "🚀 Algorithm Visualization - Comprehensive Test Suite"
echo "=================================================="
echo "Project Root: $PROJECT_ROOT"
echo "Timestamp: $TIMESTAMP"
echo ""

# Create coverage directory
mkdir -p "$COVERAGE_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run tests with coverage
run_tests_with_coverage() {
    local test_type=$1
    local test_path=$2
    local coverage_file="$COVERAGE_DIR/${test_type}_coverage_${TIMESTAMP}.out"
    local html_file="$COVERAGE_DIR/${test_type}_coverage_${TIMESTAMP}.html"
    
    print_status $BLUE "Running $test_type tests..."
    
    if go test -v -coverprofile="$coverage_file" -covermode=atomic "$test_path"; then
        print_status $GREEN "✅ $test_type tests passed"
        
        # Generate HTML coverage report
        if go tool cover -html="$coverage_file" -o "$html_file"; then
            print_status $GREEN "📊 Coverage report generated: $html_file"
        else
            print_status $YELLOW "⚠️  Could not generate HTML coverage report"
        fi
        
        # Show coverage summary
        if go tool cover -func="$coverage_file" | tail -1; then
            echo ""
        fi
        
        return 0
    else
        print_status $RED "❌ $test_type tests failed"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local benchmark_path=$1
    local benchmark_file="$COVERAGE_DIR/benchmarks_${TIMESTAMP}.txt"
    
    print_status $BLUE "Running benchmarks..."
    
    if go test -bench=. -benchmem -run=^$ "$benchmark_path" > "$benchmark_file" 2>&1; then
        print_status $GREEN "✅ Benchmarks completed"
        print_status $BLUE "📈 Benchmark results saved to: $benchmark_file"
        
        # Show top 5 benchmarks
        echo ""
        print_status $YELLOW "Top 5 Benchmarks:"
        head -20 "$benchmark_file" | grep "Benchmark" | head -5
        echo ""
        
        return 0
    else
        print_status $RED "❌ Benchmarks failed"
        return 1
    fi
}

# Function to run race detection
run_race_tests() {
    local test_path=$1
    
    print_status $BLUE "Running race detection tests..."
    
    if go test -race -v "$test_path"; then
        print_status $GREEN "✅ Race detection tests passed"
        return 0
    else
        print_status $RED "❌ Race detection tests failed"
        return 1
    fi
}

# Function to run static analysis
run_static_analysis() {
    print_status $BLUE "Running static analysis..."
    
    local static_analysis_file="$COVERAGE_DIR/static_analysis_${TIMESTAMP}.txt"
    
    # Run go vet
    echo "=== Go Vet Analysis ===" > "$static_analysis_file"
    if go vet ./... >> "$static_analysis_file" 2>&1; then
        print_status $GREEN "✅ Go vet passed"
    else
        print_status $YELLOW "⚠️  Go vet found issues"
    fi
    
    # Run go fmt check
    echo "" >> "$static_analysis_file"
    echo "=== Go Fmt Check ===" >> "$static_analysis_file"
    if [ "$(gofmt -l .)" ]; then
        print_status $YELLOW "⚠️  Code formatting issues found"
        gofmt -l . >> "$static_analysis_file"
    else
        print_status $GREEN "✅ Code formatting is correct"
    fi
    
    print_status $BLUE "📋 Static analysis results saved to: $static_analysis_file"
    return 0
}

# Main execution
main() {
    local exit_code=0
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_status $RED "❌ Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check Go version
    go_version=$(go version | cut -d' ' -f3)
    print_status $BLUE "Using Go version: $go_version"
    echo ""
    
    # Download dependencies
    print_status $BLUE "Downloading dependencies..."
    if go mod tidy; then
        print_status $GREEN "✅ Dependencies downloaded"
    else
        print_status $RED "❌ Failed to download dependencies"
        exit 1
    fi
    echo ""
    
    # Run unit tests
    if ! run_tests_with_coverage "unit" "./tests/unit/..."; then
        exit_code=1
    fi
    echo ""
    
    # Run integration tests
    if ! run_tests_with_coverage "integration" "./tests/integration/..."; then
        exit_code=1
    fi
    echo ""
    
    # Run performance tests
    if ! run_tests_with_coverage "performance" "./tests/performance/..."; then
        exit_code=1
    fi
    echo ""
    
    # Run benchmarks
    if ! run_benchmarks "./tests/performance/..."; then
        exit_code=1
    fi
    echo ""
    
    # Run race detection
    if ! run_race_tests "./..."; then
        exit_code=1
    fi
    echo ""
    
    # Run static analysis
    if ! run_static_analysis; then
        exit_code=1
    fi
    echo ""
    
    # Generate overall coverage report
    print_status $BLUE "Generating overall coverage report..."
    local overall_coverage="$COVERAGE_DIR/overall_coverage_${TIMESTAMP}.out"
    local overall_html="$COVERAGE_DIR/overall_coverage_${TIMESTAMP}.html"
    
    if go test -coverprofile="$overall_coverage" -covermode=atomic ./...; then
        if go tool cover -html="$overall_coverage" -o "$overall_html"; then
            print_status $GREEN "📊 Overall coverage report generated: $overall_html"
        fi
        
        # Show overall coverage
        print_status $YELLOW "Overall Coverage Summary:"
        go tool cover -func="$overall_coverage" | tail -1
        echo ""
    fi
    
    # Summary
    echo "=================================================="
    if [ $exit_code -eq 0 ]; then
        print_status $GREEN "🎉 All tests completed successfully!"
    else
        print_status $RED "❌ Some tests failed. Check the output above."
    fi
    
    print_status $BLUE "📁 Coverage reports saved in: $COVERAGE_DIR"
    print_status $BLUE "🕒 Timestamp: $TIMESTAMP"
    
    exit $exit_code
}

# Run main function
main "$@"

