package sorting

import (
	"math/rand"
	"time"
)

// BubbleSort implements bubble sort algorithm
func BubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		swapped := false
		for j := 0; j < n-i-1; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				swapped = true
			}
		}
		if !swapped {
			break
		}
	}
}

// SelectionSort implements selection sort algorithm
func SelectionSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		minIdx := i
		for j := i + 1; j < n; j++ {
			if arr[j] < arr[minIdx] {
				minIdx = j
			}
		}
		arr[i], arr[minIdx] = arr[minIdx], arr[i]
	}
}

// InsertionSort implements insertion sort algorithm
func InsertionSort(arr []int) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// MergeSort implements merge sort algorithm
func MergeSort(arr []int) {
	if len(arr) <= 1 {
		return
	}
	mergeSortHelper(arr, 0, len(arr)-1)
}

func mergeSortHelper(arr []int, left, right int) {
	if left < right {
		mid := left + (right-left)/2
		mergeSortHelper(arr, left, mid)
		mergeSortHelper(arr, mid+1, right)
		merge(arr, left, mid, right)
	}
}

func merge(arr []int, left, mid, right int) {
	n1 := mid - left + 1
	n2 := right - mid

	leftArr := make([]int, n1)
	rightArr := make([]int, n2)

	copy(leftArr, arr[left:left+n1])
	copy(rightArr, arr[mid+1:mid+1+n2])

	i, j, k := 0, 0, left

	for i < n1 && j < n2 {
		if leftArr[i] <= rightArr[j] {
			arr[k] = leftArr[i]
			i++
		} else {
			arr[k] = rightArr[j]
			j++
		}
		k++
	}

	for i < n1 {
		arr[k] = leftArr[i]
		i++
		k++
	}

	for j < n2 {
		arr[k] = rightArr[j]
		j++
		k++
	}
}

// QuickSort implements quick sort algorithm
func QuickSort(arr []int) {
	if len(arr) <= 1 {
		return
	}
	quickSortHelper(arr, 0, len(arr)-1)
}

func quickSortHelper(arr []int, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSortHelper(arr, low, pi-1)
		quickSortHelper(arr, pi+1, high)
	}
}

func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

// HeapSort implements heap sort algorithm
func HeapSort(arr []int) {
	n := len(arr)

	// Build heap
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	// Extract elements from heap one by one
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heapify(arr, i, 0)
	}
}

func heapify(arr []int, n, i int) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && arr[left] > arr[largest] {
		largest = left
	}

	if right < n && arr[right] > arr[largest] {
		largest = right
	}

	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		heapify(arr, n, largest)
	}
}

// RadixSort implements radix sort algorithm
func RadixSort(arr []int) {
	if len(arr) == 0 {
		return
	}

	max := getMax(arr)
	for exp := 1; max/exp > 0; exp *= 10 {
		countSort(arr, exp)
	}
}

func getMax(arr []int) int {
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}

func countSort(arr []int, exp int) {
	n := len(arr)
	output := make([]int, n)
	count := make([]int, 10)

	for i := 0; i < n; i++ {
		count[(arr[i]/exp)%10]++
	}

	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}

	for i := n - 1; i >= 0; i-- {
		output[count[(arr[i]/exp)%10]-1] = arr[i]
		count[(arr[i]/exp)%10]--
	}

	copy(arr, output)
}

// CountingSort implements counting sort algorithm
func CountingSort(arr []int) {
	if len(arr) == 0 {
		return
	}

	max := getMax(arr)
	count := make([]int, max+1)
	output := make([]int, len(arr))

	for _, v := range arr {
		count[v]++
	}

	for i := 1; i <= max; i++ {
		count[i] += count[i-1]
	}

	for i := len(arr) - 1; i >= 0; i-- {
		output[count[arr[i]]-1] = arr[i]
		count[arr[i]]--
	}

	copy(arr, output)
}

// BucketSort implements bucket sort algorithm
func BucketSort(arr []int) {
	if len(arr) == 0 {
		return
	}

	max := getMax(arr)
	bucketCount := len(arr)
	buckets := make([][]int, bucketCount)

	for i := range buckets {
		buckets[i] = make([]int, 0)
	}

	for _, v := range arr {
		bucketIndex := (v * bucketCount) / (max + 1)
		buckets[bucketIndex] = append(buckets[bucketIndex], v)
	}

	index := 0
	for _, bucket := range buckets {
		InsertionSort(bucket)
		for _, v := range bucket {
			arr[index] = v
			index++
		}
	}
}

// ShellSort implements shell sort algorithm
func ShellSort(arr []int) {
	n := len(arr)
	gap := n / 2

	for gap > 0 {
		for i := gap; i < n; i++ {
			temp := arr[i]
			j := i
			for j >= gap && arr[j-gap] > temp {
				arr[j] = arr[j-gap]
				j -= gap
			}
			arr[j] = temp
		}
		gap /= 2
	}
}

// TimSort implements tim sort algorithm (simplified version)
func TimSort(arr []int) {
	const RUN = 32
	n := len(arr)

	for i := 0; i < n; i += RUN {
		insertionSortRange(arr, i, min(i+RUN-1, n-1))
	}

	for size := RUN; size < n; size = 2 * size {
		for left := 0; left < n; left += 2 * size {
			mid := left + size - 1
			right := min(left+2*size-1, n-1)
			if mid < right {
				merge(arr, left, mid, right)
			}
		}
	}
}

func insertionSortRange(arr []int, left, right int) {
	for i := left + 1; i <= right; i++ {
		key := arr[i]
		j := i - 1
		for j >= left && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsSorted checks if an array is sorted
func IsSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}

// GenerateRandomArray generates a random array of given size
func GenerateRandomArray(size int) []int {
	rand.Seed(time.Now().UnixNano())
	arr := make([]int, size)
	for i := range arr {
		arr[i] = rand.Intn(1000)
	}
	return arr
}

// GenerateSortedArray generates a sorted array of given size
func GenerateSortedArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = i
	}
	return arr
}

// GenerateReverseSortedArray generates a reverse sorted array of given size
func GenerateReverseSortedArray(size int) []int {
	arr := make([]int, size)
	for i := range arr {
		arr[i] = size - i - 1
	}
	return arr
}

