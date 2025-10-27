package main

import (
	"fmt"
	"testing"

	"algorithm-visualization/algorithms/collision"
	"algorithm-visualization/algorithms/unionfind"
	"algorithm-visualization/algorithms/sorting"
	"algorithm-visualization/algorithms/search"
)

func TestBasicFunctionality(t *testing.T) {
	t.Run("collision detection", func(t *testing.T) {
		box1 := collision.NewAABB(0, 0, 10, 10)
		box2 := collision.NewAABB(5, 5, 15, 15)
		
		if !collision.CheckAABBCollision(box1, box2) {
			t.Error("Expected AABB collision")
		}
	})
	
	t.Run("union find", func(t *testing.T) {
		uf := unionfind.NewQuickUnion(10)
		uf.Union(0, 1)
		
		if !uf.Connected(0, 1) {
			t.Error("Expected elements to be connected")
		}
	})
	
	t.Run("sorting", func(t *testing.T) {
		arr := []int{3, 1, 4, 1, 5}
		sorting.QuickSort(arr)
		
		if !sorting.IsSorted(arr) {
			t.Error("Expected array to be sorted")
		}
	})
	
	t.Run("search", func(t *testing.T) {
		arr := []int{1, 2, 3, 4, 5}
		index := search.BinarySearch(arr, 3)
		
		if index != 2 {
			t.Errorf("Expected index 2, got %d", index)
		}
	})
}

func TestMain(m *testing.M) {
	fmt.Println("🧪 Running basic functionality tests...")
	m.Run()
}

