package search

import (
	"math"
	"sort"
)

// LinearSearch performs linear search on a slice
func LinearSearch(arr []int, target int) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return -1
}

// BinarySearch performs binary search on a sorted slice
func BinarySearch(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// BinarySearchRecursive performs recursive binary search
func BinarySearchRecursive(arr []int, target int) int {
	return binarySearchRecursiveHelper(arr, target, 0, len(arr)-1)
}

func binarySearchRecursiveHelper(arr []int, target, left, right int) int {
	if left > right {
		return -1
	}

	mid := left + (right-left)/2
	if arr[mid] == target {
		return mid
	} else if arr[mid] < target {
		return binarySearchRecursiveHelper(arr, target, mid+1, right)
	} else {
		return binarySearchRecursiveHelper(arr, target, left, mid-1)
	}
}

// TernarySearch performs ternary search on a sorted slice
func TernarySearch(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid1 := left + (right-left)/3
		mid2 := right - (right-left)/3

		if arr[mid1] == target {
			return mid1
		}
		if arr[mid2] == target {
			return mid2
		}

		if target < arr[mid1] {
			right = mid1 - 1
		} else if target > arr[mid2] {
			left = mid2 + 1
		} else {
			left = mid1 + 1
			right = mid2 - 1
		}
	}
	return -1
}

// JumpSearch performs jump search on a sorted slice
func JumpSearch(arr []int, target int) int {
	n := len(arr)
	if n == 0 {
		return -1
	}

	step := int(math.Sqrt(float64(n)))
	prev := 0

	for arr[min(step, n)-1] < target {
		prev = step
		step += int(math.Sqrt(float64(n)))
		if prev >= n {
			return -1
		}
	}

	for arr[prev] < target {
		prev++
		if prev == min(step, n) {
			return -1
		}
	}

	if arr[prev] == target {
		return prev
	}
	return -1
}

// InterpolationSearch performs interpolation search on a sorted slice
func InterpolationSearch(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right && target >= arr[left] && target <= arr[right] {
		if left == right {
			if arr[left] == target {
				return left
			}
			return -1
		}

		pos := left + ((target-arr[left])*(right-left))/(arr[right]-arr[left])

		if arr[pos] == target {
			return pos
		} else if arr[pos] < target {
			left = pos + 1
		} else {
			right = pos - 1
		}
	}
	return -1
}

// ExponentialSearch performs exponential search on a sorted slice
func ExponentialSearch(arr []int, target int) int {
	n := len(arr)
	if n == 0 {
		return -1
	}

	if arr[0] == target {
		return 0
	}

	i := 1
	for i < n && arr[i] <= target {
		i *= 2
	}

	return binarySearchRange(arr, target, i/2, min(i, n-1))
}

func binarySearchRange(arr []int, target, left, right int) int {
	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// FibonacciSearch performs fibonacci search on a sorted slice
func FibonacciSearch(arr []int, target int) int {
	n := len(arr)
	if n == 0 {
		return -1
	}

	fibMMm2 := 0
	fibMMm1 := 1
	fibM := fibMMm2 + fibMMm1

	for fibM < n {
		fibMMm2 = fibMMm1
		fibMMm1 = fibM
		fibM = fibMMm2 + fibMMm1
	}

	offset := -1

	for fibM > 1 {
		i := min(offset+fibMMm2, n-1)

		if arr[i] < target {
			fibM = fibMMm1
			fibMMm1 = fibMMm2
			fibMMm2 = fibM - fibMMm1
			offset = i
		} else if arr[i] > target {
			fibM = fibMMm2
			fibMMm1 = fibMMm1 - fibMMm2
			fibMMm2 = fibM - fibMMm1
		} else {
			return i
		}
	}

	if fibMMm1 == 1 && arr[offset+1] == target {
		return offset + 1
	}

	return -1
}

// FindFirstOccurrence finds the first occurrence of target in a sorted slice
func FindFirstOccurrence(arr []int, target int) int {
	left, right := 0, len(arr)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			result = mid
			right = mid - 1
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return result
}

// FindLastOccurrence finds the last occurrence of target in a sorted slice
func FindLastOccurrence(arr []int, target int) int {
	left, right := 0, len(arr)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			result = mid
			left = mid + 1
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return result
}

// FindCount finds the count of target in a sorted slice
func FindCount(arr []int, target int) int {
	first := FindFirstOccurrence(arr, target)
	if first == -1 {
		return 0
	}
	last := FindLastOccurrence(arr, target)
	return last - first + 1
}

// FindFloor finds the largest element smaller than or equal to target
func FindFloor(arr []int, target int) int {
	left, right := 0, len(arr)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] <= target {
			result = mid
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return result
}

// FindCeiling finds the smallest element greater than or equal to target
func FindCeiling(arr []int, target int) int {
	left, right := 0, len(arr)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] >= target {
			result = mid
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return result
}

// SearchInRotatedArray searches in a rotated sorted array
func SearchInRotatedArray(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid := left + (right-left)/2
		if arr[mid] == target {
			return mid
		}

		if arr[left] <= arr[mid] {
			if target >= arr[left] && target < arr[mid] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		} else {
			if target > arr[mid] && target <= arr[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
	}
	return -1
}

// FindPeakElement finds a peak element in an array
func FindPeakElement(arr []int) int {
	n := len(arr)
	if n == 0 {
		return -1
	}
	if n == 1 {
		return 0
	}

	left, right := 0, n-1

	for left < right {
		mid := left + (right-left)/2
		if arr[mid] > arr[mid+1] {
			right = mid
		} else {
			left = mid + 1
		}
	}
	return left
}

// SearchIn2DMatrix searches in a 2D matrix where each row and column is sorted
func SearchIn2DMatrix(matrix [][]int, target int) bool {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return false
	}

	row, col := 0, len(matrix[0])-1

	for row < len(matrix) && col >= 0 {
		if matrix[row][col] == target {
			return true
		} else if matrix[row][col] > target {
			col--
		} else {
			row++
		}
	}
	return false
}

// IsValidSearchArray checks if an array is valid for binary search
func IsValidSearchArray(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}

// GenerateSortedArray generates a sorted array for testing
func GenerateSortedArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = i
	}
	return arr
}

// GenerateRandomSortedArray generates a random sorted array
func GenerateRandomSortedArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = i*2 + (i%3)*5
	}
	sort.Ints(arr)
	return arr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

