package testutils

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// TestDataGenerator provides utilities for generating test data
type TestDataGenerator struct {
	rand *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomIntArray generates a random integer array
func (tdg *TestDataGenerator) GenerateRandomIntArray(size, max int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = tdg.rand.Intn(max)
	}
	return arr
}

// GenerateSortedIntArray generates a sorted integer array
func (tdg *TestDataGenerator) GenerateSortedIntArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = i
	}
	return arr
}

// GenerateReverseSortedIntArray generates a reverse sorted integer array
func (tdg *TestDataGenerator) GenerateReverseSortedIntArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = size - i - 1
	}
	return arr
}

// GenerateDuplicateIntArray generates an array with duplicates
func (tdg *TestDataGenerator) GenerateDuplicateIntArray(size, uniqueCount int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = tdg.rand.Intn(uniqueCount)
	}
	return arr
}

// GenerateNearlySortedIntArray generates a nearly sorted array
func (tdg *TestDataGenerator) GenerateNearlySortedIntArray(size int) []int {
	arr := tdg.GenerateSortedIntArray(size)
	// Randomly swap some elements
	for i := 0; i < size/10; i++ {
		j := tdg.rand.Intn(size)
		k := tdg.rand.Intn(size)
		arr[j], arr[k] = arr[k], arr[j]
	}
	return arr
}

// TestCase represents a test case with input and expected output
type TestCase struct {
	Name     string
	Input    interface{}
	Expected interface{}
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	Name  string
	Cases []TestCase
}

// NewTestSuite creates a new test suite
func NewTestSuite(name string) *TestSuite {
	return &TestSuite{
		Name:  name,
		Cases: make([]TestCase, 0),
	}
}

// AddTestCase adds a test case to the suite
func (ts *TestSuite) AddTestCase(name string, input, expected interface{}) {
	ts.Cases = append(ts.Cases, TestCase{
		Name:     name,
		Input:    input,
		Expected: expected,
	})
}

// AssertEqual checks if two values are equal
func AssertEqual(actual, expected interface{}) error {
	if actual != expected {
		return fmt.Errorf("expected %v, got %v", expected, actual)
	}
	return nil
}

// AssertArrayEqual checks if two arrays are equal
func AssertArrayEqual(actual, expected []int) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("array length mismatch: expected %d, got %d", len(expected), len(actual))
	}
	for i, v := range actual {
		if v != expected[i] {
			return fmt.Errorf("array element mismatch at index %d: expected %v, got %v", i, expected[i], v)
		}
	}
	return nil
}

// AssertSorted checks if an array is sorted
func AssertSorted(arr []int) error {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return fmt.Errorf("array is not sorted: element at index %d (%d) is less than element at index %d (%d)", i, arr[i], i-1, arr[i-1])
		}
	}
	return nil
}

// BenchmarkResult represents the result of a benchmark
type BenchmarkResult struct {
	Name     string
	Duration time.Duration
	Memory   int64
	Ops      int64
}

// RunBenchmark runs a benchmark function multiple times
func RunBenchmark(name string, fn func(), iterations int) BenchmarkResult {
	var totalDuration time.Duration
	var totalOps int64
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		duration := time.Since(start)
		totalDuration += duration
		totalOps++
	}
	
	avgDuration := totalDuration / time.Duration(iterations)
	
	return BenchmarkResult{
		Name:     name,
		Duration: avgDuration,
		Ops:      totalOps,
	}
}

// PerformanceProfiler helps profile algorithm performance
type PerformanceProfiler struct {
	results map[string][]time.Duration
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		results: make(map[string][]time.Duration),
	}
}

// Profile records the execution time of a function
func (pp *PerformanceProfiler) Profile(name string, fn func()) {
	start := time.Now()
	fn()
	duration := time.Since(start)
	pp.results[name] = append(pp.results[name], duration)
}

// GetAverageTime returns the average execution time for a function
func (pp *PerformanceProfiler) GetAverageTime(name string) time.Duration {
	durations := pp.results[name]
	if len(durations) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

// GetResults returns all profiling results
func (pp *PerformanceProfiler) GetResults() map[string][]time.Duration {
	return pp.results
}

// ClearResults clears all profiling results
func (pp *PerformanceProfiler) ClearResults() {
	pp.results = make(map[string][]time.Duration)
}

// TestHelper provides common test helper functions
type TestHelper struct{}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// IsSorted checks if a slice is sorted in ascending order
func (th *TestHelper) IsSorted(arr []int) bool {
	return sort.IntsAreSorted(arr)
}

// IsReverseSorted checks if a slice is sorted in descending order
func (th *TestHelper) IsReverseSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] > arr[i-1] {
			return false
		}
	}
	return true
}

// Contains checks if a slice contains a value
func (th *TestHelper) Contains(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// CountOccurrences counts occurrences of a value in a slice
func (th *TestHelper) CountOccurrences(arr []int, value int) int {
	count := 0
	for _, v := range arr {
		if v == value {
			count++
		}
	}
	return count
}

// GenerateTestMatrix generates a test matrix for comprehensive testing
func (th *TestHelper) GenerateTestMatrix() map[string][]int {
	return map[string][]int{
		"empty":           {},
		"single":          {42},
		"two_elements":    {1, 2},
		"three_elements":  {3, 1, 2},
		"duplicates":      {1, 1, 1, 1},
		"reverse_sorted":  {5, 4, 3, 2, 1},
		"already_sorted":  {1, 2, 3, 4, 5},
		"negative":        {-3, -1, -2},
		"mixed_signs":     {-1, 2, -3, 4, -5},
		"large_numbers":   {1000, 999, 1001, 998},
	}
}

// ValidateSortResult validates that a sort operation produced correct results
func (th *TestHelper) ValidateSortResult(original, sorted []int) error {
	// Check length
	if len(original) != len(sorted) {
		return fmt.Errorf("length mismatch: original %d, sorted %d", len(original), len(sorted))
	}
	
	// Check if sorted
	if !th.IsSorted(sorted) {
		return fmt.Errorf("result is not sorted")
	}
	
	// Check if all original elements are present
	originalCounts := make(map[int]int)
	sortedCounts := make(map[int]int)
	
	for _, v := range original {
		originalCounts[v]++
	}
	for _, v := range sorted {
		sortedCounts[v]++
	}
	
	for value, count := range originalCounts {
		if sortedCounts[value] != count {
			return fmt.Errorf("element count mismatch for value %d: original %d, sorted %d", value, count, sortedCounts[value])
		}
	}
	
	return nil
}

