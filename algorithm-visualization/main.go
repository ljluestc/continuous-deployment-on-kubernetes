package main

import (
	"fmt"
	"os"

	"algorithm-visualization/algorithms/collision"
	"algorithm-visualization/algorithms/unionfind"
	"algorithm-visualization/algorithms/sorting"
	"algorithm-visualization/algorithms/search"
)

// Version represents the application version
const Version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("Algorithm Visualization v%s\n", Version)
		return
	}

	fmt.Println("🚀 Algorithm Visualization Project")
	fmt.Printf("Version: %s\n", Version)
	fmt.Println("=====================================")

	// Demonstrate algorithms
	demonstrateCollisionDetection()
	demonstrateUnionFind()
	demonstrateSorting()
	demonstrateSearch()

	fmt.Println("\n✅ All algorithms demonstrated successfully!")
}

func demonstrateCollisionDetection() {
	fmt.Println("\n📦 Collision Detection Algorithms:")
	
	// AABB collision detection
	box1 := collision.NewAABB(0, 0, 10, 10)
	box2 := collision.NewAABB(5, 5, 15, 15)
	
	if collision.CheckAABBCollision(box1, box2) {
		fmt.Println("  ✅ AABB collision detected")
	} else {
		fmt.Println("  ❌ No AABB collision")
	}

	// Circle collision detection
	circle1 := collision.NewCircle(0, 0, 5)
	circle2 := collision.NewCircle(3, 3, 4)
	
	if collision.CheckCircleCollision(circle1, circle2) {
		fmt.Println("  ✅ Circle collision detected")
	} else {
		fmt.Println("  ❌ No circle collision")
	}
}

func demonstrateUnionFind() {
	fmt.Println("\n🔗 Union-Find Algorithms:")
	
	uf := unionfind.NewQuickUnion(10)
	
	// Perform some unions
	uf.Union(0, 1)
	uf.Union(2, 3)
	uf.Union(4, 5)
	uf.Union(0, 2)
	
	// Check connections
	if uf.Connected(0, 3) {
		fmt.Println("  ✅ Elements 0 and 3 are connected")
	}
	
	if uf.Connected(1, 4) {
		fmt.Println("  ✅ Elements 1 and 4 are connected")
	} else {
		fmt.Println("  ❌ Elements 1 and 4 are not connected")
	}
}

func demonstrateSorting() {
	fmt.Println("\n📊 Sorting Algorithms:")
	
	arr := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Printf("  Original array: %v\n", arr)
	
	// Quick Sort
	quickArr := make([]int, len(arr))
	copy(quickArr, arr)
	sorting.QuickSort(quickArr)
	fmt.Printf("  Quick Sort: %v\n", quickArr)
	
	// Merge Sort
	mergeArr := make([]int, len(arr))
	copy(mergeArr, arr)
	sorting.MergeSort(mergeArr)
	fmt.Printf("  Merge Sort: %v\n", mergeArr)
}

func demonstrateSearch() {
	fmt.Println("\n🔍 Search Algorithms:")
	
	arr := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	target := 7
	
	// Linear Search
	linearIndex := search.LinearSearch(arr, target)
	fmt.Printf("  Linear Search found %d at index: %d\n", target, linearIndex)
	
	// Binary Search
	binaryIndex := search.BinarySearch(arr, target)
	fmt.Printf("  Binary Search found %d at index: %d\n", target, binaryIndex)
}