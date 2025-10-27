package unit_test

import (
	"testing"

	"algorithm-visualization/algorithms/unionfind"
	"algorithm-visualization/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuickFind_Creation(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"small size", 5, 5},
		{"medium size", 100, 100},
		{"large size", 1000, 1000},
		{"zero size", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qf := unionfind.NewQuickFind(tt.size)
			assert.Equal(t, tt.expected, qf.Count())
			
			// Verify all elements are initially separate
			for i := 0; i < tt.size; i++ {
				assert.Equal(t, i, qf.Find(i))
			}
		})
	}
}

func TestQuickFind_Union(t *testing.T) {
	qf := unionfind.NewQuickFind(10)
	
	t.Run("union two elements", func(t *testing.T) {
		qf.Union(0, 1)
		assert.True(t, qf.Connected(0, 1))
		assert.Equal(t, 9, qf.Count())
	})
	
	t.Run("union same element", func(t *testing.T) {
		initialCount := qf.Count()
		qf.Union(2, 2)
		assert.Equal(t, initialCount, qf.Count())
		assert.True(t, qf.Connected(2, 2))
	})
	
	t.Run("union multiple elements", func(t *testing.T) {
		qf.Union(2, 3)
		qf.Union(3, 4)
		assert.True(t, qf.Connected(2, 4))
		assert.Equal(t, 7, qf.Count())
	})
	
	t.Run("union already connected elements", func(t *testing.T) {
		initialCount := qf.Count()
		qf.Union(0, 1) // Already connected
		assert.Equal(t, initialCount, qf.Count())
	})
}

func TestQuickFind_Connected(t *testing.T) {
	qf := unionfind.NewQuickFind(5)
	
	t.Run("elements not connected initially", func(t *testing.T) {
		assert.False(t, qf.Connected(0, 1))
		assert.False(t, qf.Connected(2, 3))
	})
	
	t.Run("element connected to itself", func(t *testing.T) {
		assert.True(t, qf.Connected(0, 0))
		assert.True(t, qf.Connected(4, 4))
	})
	
	t.Run("elements connected after union", func(t *testing.T) {
		qf.Union(0, 1)
		assert.True(t, qf.Connected(0, 1))
		assert.True(t, qf.Connected(1, 0))
	})
}

func TestQuickUnion_Creation(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"small size", 5, 5},
		{"medium size", 100, 100},
		{"large size", 1000, 1000},
		{"zero size", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qu := unionfind.NewQuickUnion(tt.size)
			assert.Equal(t, tt.expected, qu.Count())
			
			// Verify all elements are initially separate
			for i := 0; i < tt.size; i++ {
				assert.Equal(t, i, qu.Find(i))
			}
		})
	}
}

func TestQuickUnion_Union(t *testing.T) {
	qu := unionfind.NewQuickUnion(10)
	
	t.Run("union two elements", func(t *testing.T) {
		qu.Union(0, 1)
		assert.True(t, qu.Connected(0, 1))
		assert.Equal(t, 9, qu.Count())
	})
	
	t.Run("union same element", func(t *testing.T) {
		initialCount := qu.Count()
		qu.Union(2, 2)
		assert.Equal(t, initialCount, qu.Count())
		assert.True(t, qu.Connected(2, 2))
	})
	
	t.Run("union multiple elements", func(t *testing.T) {
		qu.Union(2, 3)
		qu.Union(3, 4)
		assert.True(t, qu.Connected(2, 4))
		assert.Equal(t, 7, qu.Count())
	})
}

func TestWeightedQuickUnion_Creation(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"small size", 5, 5},
		{"medium size", 100, 100},
		{"large size", 1000, 1000},
		{"zero size", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wqu := unionfind.NewWeightedQuickUnion(tt.size)
			assert.Equal(t, tt.expected, wqu.Count())
			
			// Verify all elements are initially separate
			for i := 0; i < tt.size; i++ {
				assert.Equal(t, i, wqu.Find(i))
			}
		})
	}
}

func TestWeightedQuickUnion_Union(t *testing.T) {
	wqu := unionfind.NewWeightedQuickUnion(10)
	
	t.Run("union two elements", func(t *testing.T) {
		wqu.Union(0, 1)
		assert.True(t, wqu.Connected(0, 1))
		assert.Equal(t, 9, wqu.Count())
	})
	
	t.Run("union same element", func(t *testing.T) {
		initialCount := wqu.Count()
		wqu.Union(2, 2)
		assert.Equal(t, initialCount, wqu.Count())
		assert.True(t, wqu.Connected(2, 2))
	})
	
	t.Run("union multiple elements", func(t *testing.T) {
		wqu.Union(2, 3)
		wqu.Union(3, 4)
		assert.True(t, wqu.Connected(2, 4))
		assert.Equal(t, 7, wqu.Count())
	})
}

func TestWeightedQuickUnionWithPathCompression_Creation(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{"small size", 5, 5},
		{"medium size", 100, 100},
		{"large size", 1000, 1000},
		{"zero size", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(tt.size)
			assert.Equal(t, tt.expected, wqupc.Count())
			
			// Verify all elements are initially separate
			for i := 0; i < tt.size; i++ {
				assert.Equal(t, i, wqupc.Find(i))
			}
		})
	}
}

func TestWeightedQuickUnionWithPathCompression_Union(t *testing.T) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(10)
	
	t.Run("union two elements", func(t *testing.T) {
		wqupc.Union(0, 1)
		assert.True(t, wqupc.Connected(0, 1))
		assert.Equal(t, 9, wqupc.Count())
	})
	
	t.Run("union same element", func(t *testing.T) {
		initialCount := wqupc.Count()
		wqupc.Union(2, 2)
		assert.Equal(t, initialCount, wqupc.Count())
		assert.True(t, wqupc.Connected(2, 2))
	})
	
	t.Run("union multiple elements", func(t *testing.T) {
		wqupc.Union(2, 3)
		wqupc.Union(3, 4)
		assert.True(t, wqupc.Connected(2, 4))
		assert.Equal(t, 7, wqupc.Count())
	})
}

func TestWeightedQuickUnionWithPathCompression_GetComponentSize(t *testing.T) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(10)
	
	t.Run("single element component", func(t *testing.T) {
		assert.Equal(t, 1, wqupc.GetComponentSize(0))
	})
	
	t.Run("multiple element component", func(t *testing.T) {
		wqupc.Union(0, 1)
		wqupc.Union(1, 2)
		assert.Equal(t, 3, wqupc.GetComponentSize(0))
		assert.Equal(t, 3, wqupc.GetComponentSize(1))
		assert.Equal(t, 3, wqupc.GetComponentSize(2))
	})
}

func TestWeightedQuickUnionWithPathCompression_GetAllComponents(t *testing.T) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(5)
	
	t.Run("all separate initially", func(t *testing.T) {
		components := wqupc.GetAllComponents()
		assert.Equal(t, 5, len(components))
	})
	
	t.Run("some connected", func(t *testing.T) {
		wqupc.Union(0, 1)
		wqupc.Union(2, 3)
		
		components := wqupc.GetAllComponents()
		assert.Equal(t, 3, len(components))
		
		// Check that components contain correct elements
		found := make(map[int]bool)
		for _, component := range components {
			for _, element := range component {
				found[element] = true
			}
		}
		
		for i := 0; i < 5; i++ {
			assert.True(t, found[i], "Element %d should be in some component", i)
		}
	})
}

func TestWeightedQuickUnionWithPathCompression_IsValidIndex(t *testing.T) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(5)
	
	t.Run("valid indices", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			assert.True(t, wqupc.IsValidIndex(i))
		}
	})
	
	t.Run("invalid indices", func(t *testing.T) {
		assert.False(t, wqupc.IsValidIndex(-1))
		assert.False(t, wqupc.IsValidIndex(5))
		assert.False(t, wqupc.IsValidIndex(100))
	})
}

func TestWeightedQuickUnionWithPathCompression_Reset(t *testing.T) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(5)
	
	t.Run("reset after unions", func(t *testing.T) {
		wqupc.Union(0, 1)
		wqupc.Union(2, 3)
		assert.Equal(t, 3, wqupc.Count())
		
		wqupc.Reset()
		assert.Equal(t, 5, wqupc.Count())
		
		// Verify all elements are separate again
		for i := 0; i < 5; i++ {
			assert.Equal(t, i, wqupc.Find(i))
		}
	})
}

// Complex test scenarios
func TestUnionFind_ComplexScenarios(t *testing.T) {
	t.Run("chain union", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(10)
		
		// Create a chain: 0-1-2-3-4
		for i := 0; i < 4; i++ {
			wqupc.Union(i, i+1)
		}
		
		// All elements should be connected
		for i := 0; i <= 4; i++ {
			for j := 0; j <= 4; j++ {
				assert.True(t, wqupc.Connected(i, j), "Elements %d and %d should be connected", i, j)
			}
		}
		
		assert.Equal(t, 6, wqupc.Count()) // 5 connected + 5 separate
	})
	
	t.Run("star union", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(10)
		
		// Create a star: 0 connected to 1,2,3,4
		for i := 1; i <= 4; i++ {
			wqupc.Union(0, i)
		}
		
		// All elements 0-4 should be connected
		for i := 0; i <= 4; i++ {
			for j := 0; j <= 4; j++ {
				assert.True(t, wqupc.Connected(i, j), "Elements %d and %d should be connected", i, j)
			}
		}
		
		assert.Equal(t, 6, wqupc.Count()) // 5 connected + 5 separate
	})
	
	t.Run("multiple components", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(10)
		
		// Create two separate components: {0,1,2} and {3,4,5}
		wqupc.Union(0, 1)
		wqupc.Union(1, 2)
		wqupc.Union(3, 4)
		wqupc.Union(4, 5)
		
		// Elements within each component should be connected
		assert.True(t, wqupc.Connected(0, 2))
		assert.True(t, wqupc.Connected(3, 5))
		
		// Elements from different components should not be connected
		assert.False(t, wqupc.Connected(0, 3))
		assert.False(t, wqupc.Connected(2, 5))
		
		assert.Equal(t, 6, wqupc.Count()) // 2 components of 3 + 4 separate
	})
}

// Performance comparison tests
func TestUnionFind_PerformanceComparison(t *testing.T) {
	size := 1000
	operations := 10000
	
	t.Run("QuickFind performance", func(t *testing.T) {
		qf := unionfind.NewQuickFind(size)
		
		// Perform random unions
		for i := 0; i < operations; i++ {
			p := i % size
			q := (i + 1) % size
			qf.Union(p, q)
		}
		
		assert.True(t, qf.Count() < size)
	})
	
	t.Run("QuickUnion performance", func(t *testing.T) {
		qu := unionfind.NewQuickUnion(size)
		
		// Perform random unions
		for i := 0; i < operations; i++ {
			p := i % size
			q := (i + 1) % size
			qu.Union(p, q)
		}
		
		assert.True(t, qu.Count() < size)
	})
	
	t.Run("WeightedQuickUnion performance", func(t *testing.T) {
		wqu := unionfind.NewWeightedQuickUnion(size)
		
		// Perform random unions
		for i := 0; i < operations; i++ {
			p := i % size
			q := (i + 1) % size
			wqu.Union(p, q)
		}
		
		assert.True(t, wqu.Count() < size)
	})
	
	t.Run("WeightedQuickUnionWithPathCompression performance", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(size)
		
		// Perform random unions
		for i := 0; i < operations; i++ {
			p := i % size
			q := (i + 1) % size
			wqupc.Union(p, q)
		}
		
		assert.True(t, wqupc.Count() < size)
	})
}

// Benchmark tests
func BenchmarkQuickFind_Union(b *testing.B) {
	qf := unionfind.NewQuickFind(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := i % 1000
		q := (i + 1) % 1000
		qf.Union(p, q)
	}
}

func BenchmarkQuickUnion_Union(b *testing.B) {
	qu := unionfind.NewQuickUnion(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := i % 1000
		q := (i + 1) % 1000
		qu.Union(p, q)
	}
}

func BenchmarkWeightedQuickUnion_Union(b *testing.B) {
	wqu := unionfind.NewWeightedQuickUnion(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := i % 1000
		q := (i + 1) % 1000
		wqu.Union(p, q)
	}
}

func BenchmarkWeightedQuickUnionWithPathCompression_Union(b *testing.B) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := i % 1000
		q := (i + 1) % 1000
		wqupc.Union(p, q)
	}
}

func BenchmarkQuickFind_Find(b *testing.B) {
	qf := unionfind.NewQuickFind(1000)
	// Create some unions first
	for i := 0; i < 500; i++ {
		qf.Union(i, i+500)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qf.Find(i % 1000)
	}
}

func BenchmarkWeightedQuickUnionWithPathCompression_Find(b *testing.B) {
	wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(1000)
	// Create some unions first
	for i := 0; i < 500; i++ {
		wqupc.Union(i, i+500)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wqupc.Find(i % 1000)
	}
}

// Property-based tests
func TestUnionFind_Properties(t *testing.T) {
	tdg := utils.NewTestDataGenerator()
	
	t.Run("reflexive property", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(100)
		
		for i := 0; i < 100; i++ {
			assert.True(t, wqupc.Connected(i, i), "Element %d should be connected to itself", i)
		}
	})
	
	t.Run("symmetric property", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(100)
		
		// Perform random unions
		for i := 0; i < 200; i++ {
			p := tdg.GenerateRandomIntArray(1, 100)[0]
			q := tdg.GenerateRandomIntArray(1, 100)[0]
			wqupc.Union(p, q)
		}
		
		// Check symmetry
		for i := 0; i < 100; i++ {
			for j := 0; j < 100; j++ {
				connected1 := wqupc.Connected(i, j)
				connected2 := wqupc.Connected(j, i)
				assert.Equal(t, connected1, connected2, "Connection should be symmetric between %d and %d", i, j)
			}
		}
	})
	
	t.Run("transitive property", func(t *testing.T) {
		wqupc := unionfind.NewWeightedQuickUnionWithPathCompression(50)
		
		// Create transitive connections: 0-1-2, 3-4-5, etc.
		for i := 0; i < 50; i += 3 {
			if i+1 < 50 {
				wqupc.Union(i, i+1)
			}
			if i+2 < 50 {
				wqupc.Union(i+1, i+2)
			}
		}
		
		// Check transitivity
		for i := 0; i < 50; i += 3 {
			if i+2 < 50 {
				assert.True(t, wqupc.Connected(i, i+2), "Transitive connection should exist between %d and %d", i, i+2)
			}
		}
	})
}
