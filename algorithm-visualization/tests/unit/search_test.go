package unit_test

import (
	"sort"
	"testing"

	"algorithm-visualization/algorithms/search"
	"algorithm-visualization/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data generators
func generateSearchTestCases() []struct {
	name     string
	arr      []int
	target   int
	expected int
} {
	return []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"empty array", []int{}, 5, -1},
		{"single element found", []int{42}, 42, 0},
		{"single element not found", []int{42}, 5, -1},
		{"two elements found first", []int{1, 2}, 1, 0},
		{"two elements found second", []int{1, 2}, 2, 1},
		{"two elements not found", []int{1, 2}, 3, -1},
		{"three elements found middle", []int{1, 2, 3}, 2, 1},
		{"three elements found last", []int{1, 2, 3}, 3, 2},
		{"three elements not found", []int{1, 2, 3}, 4, -1},
		{"duplicates found first", []int{1, 1, 2, 2}, 1, 0},
		{"duplicates found second", []int{1, 1, 2, 2}, 2, 2},
		{"large array found", []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}, 7, 3},
		{"large array not found", []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}, 6, -1},
		{"negative numbers", []int{-5, -3, -1, 0, 2, 4}, -1, 2},
		{"all same elements", []int{5, 5, 5, 5}, 5, 0},
	}
}

func TestLinearSearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := search.LinearSearch(tc.arr, tc.target)
			assert.Equal(t, tc.expected, result, "LinearSearch should return correct index")
		})
	}
}

func TestBinarySearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Binary search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.BinarySearch(arr, tc.target)
			assert.Equal(t, expected, result, "BinarySearch should return correct index")
		})
	}
}

func TestBinarySearchRecursive(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Binary search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.BinarySearchRecursive(arr, tc.target)
			assert.Equal(t, expected, result, "BinarySearchRecursive should return correct index")
		})
	}
}

func TestTernarySearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ternary search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.TernarySearch(arr, tc.target)
			assert.Equal(t, expected, result, "TernarySearch should return correct index")
		})
	}
}

func TestJumpSearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Jump search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.JumpSearch(arr, tc.target)
			assert.Equal(t, expected, result, "JumpSearch should return correct index")
		})
	}
}

func TestInterpolationSearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Interpolation search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.InterpolationSearch(arr, tc.target)
			assert.Equal(t, expected, result, "InterpolationSearch should return correct index")
		})
	}
}

func TestExponentialSearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Exponential search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.ExponentialSearch(arr, tc.target)
			assert.Equal(t, expected, result, "ExponentialSearch should return correct index")
		})
	}
}

func TestFibonacciSearch(t *testing.T) {
	testCases := generateSearchTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Fibonacci search requires sorted array
			arr := make([]int, len(tc.arr))
			copy(arr, tc.arr)
			sort.Ints(arr)
			
			// Find expected index in sorted array
			expected := -1
			if tc.expected != -1 {
				expected = sort.SearchInts(arr, tc.target)
				if expected >= len(arr) || arr[expected] != tc.target {
					expected = -1
				}
			}
			
			result := search.FibonacciSearch(arr, tc.target)
			assert.Equal(t, expected, result, "FibonacciSearch should return correct index")
		})
	}
}

func TestFindFirstOccurrence(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"no duplicates", []int{1, 2, 3, 4, 5}, 3, 2},
		{"duplicates first occurrence", []int{1, 2, 2, 2, 3}, 2, 1},
		{"duplicates last occurrence", []int{1, 2, 2, 2, 3}, 2, 1},
		{"not found", []int{1, 2, 3, 4, 5}, 6, -1},
		{"all same elements", []int{5, 5, 5, 5}, 5, 0},
		{"empty array", []int{}, 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindFirstOccurrence(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "FindFirstOccurrence should return correct index")
		})
	}
}

func TestFindLastOccurrence(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"no duplicates", []int{1, 2, 3, 4, 5}, 3, 2},
		{"duplicates last occurrence", []int{1, 2, 2, 2, 3}, 2, 3},
		{"not found", []int{1, 2, 3, 4, 5}, 6, -1},
		{"all same elements", []int{5, 5, 5, 5}, 5, 3},
		{"empty array", []int{}, 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindLastOccurrence(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "FindLastOccurrence should return correct index")
		})
	}
}

func TestFindCount(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"no duplicates", []int{1, 2, 3, 4, 5}, 3, 1},
		{"multiple duplicates", []int{1, 2, 2, 2, 3}, 2, 3},
		{"not found", []int{1, 2, 3, 4, 5}, 6, 0},
		{"all same elements", []int{5, 5, 5, 5}, 5, 4},
		{"empty array", []int{}, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindCount(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "FindCount should return correct count")
		})
	}
}

func TestFindFloor(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"exact match", []int{1, 2, 3, 4, 5}, 3, 2},
		{"floor value", []int{1, 2, 3, 4, 5}, 3, 2},
		{"smaller than all", []int{1, 2, 3, 4, 5}, 0, -1},
		{"larger than all", []int{1, 2, 3, 4, 5}, 6, 4},
		{"duplicates", []int{1, 2, 2, 2, 3}, 2, 3},
		{"empty array", []int{}, 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindFloor(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "FindFloor should return correct index")
		})
	}
}

func TestFindCeiling(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"exact match", []int{1, 2, 3, 4, 5}, 3, 2},
		{"ceiling value", []int{1, 2, 3, 4, 5}, 2, 2},
		{"smaller than all", []int{1, 2, 3, 4, 5}, 0, 0},
		{"larger than all", []int{1, 2, 3, 4, 5}, 6, -1},
		{"duplicates", []int{1, 2, 2, 2, 3}, 2, 1},
		{"empty array", []int{}, 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindCeiling(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "FindCeiling should return correct index")
		})
	}
}

func TestSearchInRotatedArray(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{"not rotated", []int{1, 2, 3, 4, 5}, 3, 2},
		{"rotated once", []int{5, 1, 2, 3, 4}, 3, 3},
		{"rotated multiple times", []int{3, 4, 5, 1, 2}, 3, 0},
		{"not found", []int{3, 4, 5, 1, 2}, 6, -1},
		{"single element found", []int{1}, 1, 0},
		{"single element not found", []int{1}, 2, -1},
		{"empty array", []int{}, 5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.SearchInRotatedArray(tt.arr, tt.target)
			assert.Equal(t, tt.expected, result, "SearchInRotatedArray should return correct index")
		})
	}
}

func TestFindPeakElement(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		expected int
	}{
		{"single element", []int{1}, 0},
		{"two elements ascending", []int{1, 2}, 1},
		{"two elements descending", []int{2, 1}, 0},
		{"peak in middle", []int{1, 3, 2}, 1},
		{"peak at end", []int{1, 2, 3}, 2},
		{"peak at start", []int{3, 2, 1}, 0},
		{"multiple peaks", []int{1, 3, 2, 4, 1}, 1}, // Should return any peak
		{"empty array", []int{}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.FindPeakElement(tt.arr)
			if tt.expected == -1 {
				assert.Equal(t, tt.expected, result, "FindPeakElement should return -1 for empty array")
			} else {
				assert.True(t, result >= 0 && result < len(tt.arr), "FindPeakElement should return valid index")
				// Verify it's actually a peak
				if result > 0 {
					assert.True(t, tt.arr[result] >= tt.arr[result-1], "Element should be >= left neighbor")
				}
				if result < len(tt.arr)-1 {
					assert.True(t, tt.arr[result] >= tt.arr[result+1], "Element should be >= right neighbor")
				}
			}
		})
	}
}

func TestSearchIn2DMatrix(t *testing.T) {
	tests := []struct {
		name     string
		matrix   [][]int
		target   int
		expected bool
	}{
		{
			name: "found in matrix",
			matrix: [][]int{
				{1, 4, 7, 11},
				{2, 5, 8, 12},
				{3, 6, 9, 16},
			},
			target:   5,
			expected: true,
		},
		{
			name: "not found in matrix",
			matrix: [][]int{
				{1, 4, 7, 11},
				{2, 5, 8, 12},
				{3, 6, 9, 16},
			},
			target:   13,
			expected: false,
		},
		{
			name: "empty matrix",
			matrix: [][]int{},
			target:   5,
			expected: false,
		},
		{
			name: "single element found",
			matrix: [][]int{{5}},
			target:   5,
			expected: true,
		},
		{
			name: "single element not found",
			matrix: [][]int{{5}},
			target:   3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.SearchIn2DMatrix(tt.matrix, tt.target)
			assert.Equal(t, tt.expected, result, "SearchIn2DMatrix should return correct result")
		})
	}
}

func TestIsValidSearchArray(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		expected bool
	}{
		{"sorted array", []int{1, 2, 3, 4, 5}, true},
		{"unsorted array", []int{1, 3, 2, 4, 5}, false},
		{"empty array", []int{}, true},
		{"single element", []int{42}, true},
		{"duplicates sorted", []int{1, 1, 2, 2, 3}, true},
		{"reverse sorted", []int{5, 4, 3, 2, 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search.IsValidSearchArray(tt.arr)
			assert.Equal(t, tt.expected, result, "IsValidSearchArray should return correct result")
		})
	}
}

func TestGenerateSortedArray_Search(t *testing.T) {
	t.Run("generate sorted arrays", func(t *testing.T) {
		sizes := []int{0, 1, 10, 100}
		
		for _, size := range sizes {
			arr := search.GenerateSortedArray(size)
			assert.Equal(t, size, len(arr), "Array should have correct size")
			assert.True(t, search.IsValidSearchArray(arr), "Generated array should be sorted")
			
			// Check that elements are consecutive
			for i := 0; i < size; i++ {
				assert.Equal(t, i, arr[i], "Element at index %d should be %d", i, i)
			}
		}
	})
}

func TestGenerateRandomSortedArray(t *testing.T) {
	t.Run("generate random sorted arrays", func(t *testing.T) {
		sizes := []int{0, 1, 10, 100}
		
		for _, size := range sizes {
			arr := search.GenerateRandomSortedArray(size)
			assert.Equal(t, size, len(arr), "Array should have correct size")
			assert.True(t, search.IsValidSearchArray(arr), "Generated array should be sorted")
		}
	})
}

// Edge cases and stress tests
func TestSearchAlgorithms_EdgeCases(t *testing.T) {
	t.Run("very large array", func(t *testing.T) {
		arr := search.GenerateSortedArray(10000)
		target := 5000
		
		algorithms := []struct {
			name string
			fn   func([]int, int) int
		}{
			{"BinarySearch", search.BinarySearch},
			{"BinarySearchRecursive", search.BinarySearchRecursive},
			{"TernarySearch", search.TernarySearch},
			{"JumpSearch", search.JumpSearch},
			{"InterpolationSearch", search.InterpolationSearch},
			{"ExponentialSearch", search.ExponentialSearch},
			{"FibonacciSearch", search.FibonacciSearch},
		}
		
		for _, alg := range algorithms {
			t.Run(alg.name, func(t *testing.T) {
				result := alg.fn(arr, target)
				assert.Equal(t, target, result, "%s should find correct index in large array", alg.name)
			})
		}
	})
	
	t.Run("array with many duplicates", func(t *testing.T) {
		arr := make([]int, 1000)
		for i := range arr {
			arr[i] = i % 10 // Only 10 unique values
		}
		sort.Ints(arr)
		target := 5
		
		algorithms := []struct {
			name string
			fn   func([]int, int) int
		}{
			{"BinarySearch", search.BinarySearch},
			{"FindFirstOccurrence", search.FindFirstOccurrence},
			{"FindLastOccurrence", search.FindLastOccurrence},
		}
		
		for _, alg := range algorithms {
			t.Run(alg.name, func(t *testing.T) {
				result := alg.fn(arr, target)
				if result != -1 {
					assert.Equal(t, target, arr[result], "%s should find correct element", alg.name)
				}
			})
		}
	})
}

// Benchmark tests
func BenchmarkLinearSearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.LinearSearch(arr, target)
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.BinarySearch(arr, target)
	}
}

func BenchmarkBinarySearchRecursive(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.BinarySearchRecursive(arr, target)
	}
}

func BenchmarkTernarySearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.TernarySearch(arr, target)
	}
}

func BenchmarkJumpSearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.JumpSearch(arr, target)
	}
}

func BenchmarkInterpolationSearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.InterpolationSearch(arr, target)
	}
}

func BenchmarkExponentialSearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.ExponentialSearch(arr, target)
	}
}

func BenchmarkFibonacciSearch(b *testing.B) {
	arr := search.GenerateSortedArray(1000)
	target := 500
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		search.FibonacciSearch(arr, target)
	}
}

// Performance comparison for different array sizes
func BenchmarkSearchAlgorithms_DifferentSizes(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}
	
	algorithms := []struct {
		name string
		fn   func([]int, int) int
	}{
		{"LinearSearch", search.LinearSearch},
		{"BinarySearch", search.BinarySearch},
		{"TernarySearch", search.TernarySearch},
		{"JumpSearch", search.JumpSearch},
		{"InterpolationSearch", search.InterpolationSearch},
	}
	
	for _, size := range sizes {
		arr := search.GenerateSortedArray(size)
		target := size / 2
		
		for _, algorithm := range algorithms {
			b.Run(algorithm.name+"_Size"+string(rune(size)), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					algorithm.fn(arr, target)
				}
			})
		}
	}
}

// Property-based tests
func TestSearchAlgorithms_Properties(t *testing.T) {
	tdg := utils.NewTestDataGenerator()
	
	t.Run("search correctness", func(t *testing.T) {
		algorithms := []struct {
			name string
			fn   func([]int, int) int
		}{
			{"BinarySearch", search.BinarySearch},
			{"BinarySearchRecursive", search.BinarySearchRecursive},
			{"TernarySearch", search.TernarySearch},
			{"JumpSearch", search.JumpSearch},
			{"InterpolationSearch", search.InterpolationSearch},
			{"ExponentialSearch", search.ExponentialSearch},
			{"FibonacciSearch", search.FibonacciSearch},
		}
		
		for _, alg := range algorithms {
			t.Run(alg.name, func(t *testing.T) {
				for i := 0; i < 50; i++ {
					arr := tdg.GenerateRandomIntArray(100, 1000)
					sort.Ints(arr)
					target := tdg.GenerateRandomIntArray(1, 1000)[0]
					
					result := alg.fn(arr, target)
					
					if result != -1 {
						assert.Equal(t, target, arr[result], "%s should return correct element", alg.name)
					} else {
						// Verify element is not in array
						found := false
						for _, v := range arr {
							if v == target {
								found = true
								break
							}
						}
						assert.False(t, found, "%s should correctly identify missing element", alg.name)
					}
				}
			})
		}
	})
	
	t.Run("first occurrence property", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			arr := tdg.GenerateRandomIntArray(100, 10) // Small range for duplicates
			sort.Ints(arr)
			target := tdg.GenerateRandomIntArray(1, 10)[0]
			
			result := search.FindFirstOccurrence(arr, target)
			
			if result != -1 {
				assert.Equal(t, target, arr[result], "FindFirstOccurrence should return correct element")
				
				// Verify it's the first occurrence
				if result > 0 {
					assert.NotEqual(t, target, arr[result-1], "Should be the first occurrence")
				}
			}
		}
	})
	
	t.Run("last occurrence property", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			arr := tdg.GenerateRandomIntArray(100, 10) // Small range for duplicates
			sort.Ints(arr)
			target := tdg.GenerateRandomIntArray(1, 10)[0]
			
			result := search.FindLastOccurrence(arr, target)
			
			if result != -1 {
				assert.Equal(t, target, arr[result], "FindLastOccurrence should return correct element")
				
				// Verify it's the last occurrence
				if result < len(arr)-1 {
					assert.NotEqual(t, target, arr[result+1], "Should be the last occurrence")
				}
			}
		}
	})
}
