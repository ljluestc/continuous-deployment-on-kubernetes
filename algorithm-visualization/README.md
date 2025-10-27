# Algorithm Visualization Project

A comprehensive Go-based project for visualizing and testing fundamental algorithms including collision detection, union-find, sorting, and search algorithms.

## Project Structure

```
algorithm-visualization/
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── main.go                   # Main application entry point
├── algorithms/               # Core algorithm implementations
│   ├── collision/           # Collision detection algorithms
│   ├── unionfind/           # Union-Find data structure
│   ├── sorting/             # Sorting algorithms
│   └── search/              # Search algorithms
├── visualization/           # Visualization components
├── tests/                   # Comprehensive test suite
│   ├── unit/               # Unit tests
│   ├── integration/        # Integration tests
│   ├── performance/        # Performance tests
│   └── utils/              # Test utilities
├── scripts/                # Automation scripts
├── docs/                   # Documentation
└── coverage/               # Coverage reports
```

## Testing Framework

- **Unit Tests**: Comprehensive testing of individual algorithms
- **Integration Tests**: End-to-end testing of algorithm interactions
- **Performance Tests**: Benchmarking and performance analysis
- **Coverage**: 100% code coverage target
- **CI/CD**: Automated testing pipeline

## Algorithms Implemented

### Collision Detection
- AABB (Axis-Aligned Bounding Box) collision detection
- Circle collision detection
- Point-in-polygon algorithms
- Spatial partitioning (QuadTree, Grid)

### Union-Find
- Quick Find implementation
- Quick Union implementation
- Weighted Quick Union
- Path compression optimization

### Sorting Algorithms
- Bubble Sort
- Selection Sort
- Insertion Sort
- Merge Sort
- Quick Sort
- Heap Sort
- Radix Sort

### Search Algorithms
- Linear Search
- Binary Search
- Ternary Search
- Jump Search
- Interpolation Search
- Exponential Search

## Getting Started

1. **Setup Environment**:
   ```bash
   cd algorithm-visualization
   go mod tidy
   ```

2. **Run Tests**:
   ```bash
   go test ./...
   go test -cover ./...
   ```

3. **Run Performance Tests**:
   ```bash
   go test -bench=. ./...
   ```

4. **Generate Coverage Report**:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

## Testing Phases

- **Phase 1**: Infrastructure Setup (6 tasks)
- **Phase 2**: Unit Tests (4 tasks, 14 hours)
- **Phase 3**: Integration Tests
- **Phase 4**: Performance & Edge Cases
- **Phase 5**: Test Automation Scripts
- **Phase 6**: CI/CD Pipeline
- **Phase 7**: Documentation & Validation

## Coverage Goals

- **Target**: 100% code coverage
- **Minimum**: 90% code coverage
- **Current**: TBD (to be implemented)

## CI/CD Pipeline

- Automated testing on every commit
- Coverage reporting
- Performance benchmarking
- Documentation generation
- Release automation

---

**Status**: 🚧 In Development
**Version**: 0.1.0
**Go Version**: 1.21+

