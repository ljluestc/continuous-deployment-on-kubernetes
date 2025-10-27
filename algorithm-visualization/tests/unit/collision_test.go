package unit_test

import (
	"testing"

	"algorithm-visualization/algorithms/collision"
	"algorithm-visualization/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAABB_Creation(t *testing.T) {
	tests := []struct {
		name           string
		x, y, w, h     float64
		expectedCenter *collision.Point
		expectedArea   float64
	}{
		{
			name:           "basic AABB",
			x:              0, y: 0, w: 10, h: 10,
			expectedCenter: &collision.Point{X: 5, Y: 5},
			expectedArea:   100,
		},
		{
			name:           "negative coordinates",
			x:              -5, y: -5, w: 10, h: 10,
			expectedCenter: &collision.Point{X: 0, Y: 0},
			expectedArea:   100,
		},
		{
			name:           "zero dimensions",
			x:              0, y: 0, w: 0, h: 0,
			expectedCenter: &collision.Point{X: 0, Y: 0},
			expectedArea:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aabb := collision.NewAABB(tt.x, tt.y, tt.w, tt.h)
			
			assert.Equal(t, tt.x, aabb.X)
			assert.Equal(t, tt.y, aabb.Y)
			assert.Equal(t, tt.w, aabb.Width)
			assert.Equal(t, tt.h, aabb.Height)
			
			center := aabb.GetCenter()
			assert.Equal(t, tt.expectedCenter.X, center.X)
			assert.Equal(t, tt.expectedCenter.Y, center.Y)
			
			area := aabb.GetArea()
			assert.Equal(t, tt.expectedArea, area)
		})
	}
}

func TestCircle_Creation(t *testing.T) {
	tests := []struct {
		name         string
		x, y, radius float64
		expectedArea float64
	}{
		{
			name:         "basic circle",
			x:            0, y: 0, radius: 5,
			expectedArea: 78.53981633974483, // π * 5²
		},
		{
			name:         "zero radius",
			x:            0, y: 0, radius: 0,
			expectedArea: 0,
		},
		{
			name:         "negative coordinates",
			x:            -3, y: -4, radius: 2,
			expectedArea: 12.566370614359172, // π * 2²
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circle := collision.NewCircle(tt.x, tt.y, tt.radius)
			
			assert.Equal(t, tt.x, circle.X)
			assert.Equal(t, tt.y, circle.Y)
			assert.Equal(t, tt.radius, circle.Radius)
			
			area := circle.GetArea()
			assert.InDelta(t, tt.expectedArea, area, 0.0001)
		})
	}
}

func TestPoint_Creation(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
	}{
		{"origin", 0, 0},
		{"positive coordinates", 5, 10},
		{"negative coordinates", -3, -7},
		{"decimal coordinates", 3.14, 2.71},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			point := collision.NewPoint(tt.x, tt.y)
			
			assert.Equal(t, tt.x, point.X)
			assert.Equal(t, tt.y, point.Y)
		})
	}
}

func TestPolygon_Creation(t *testing.T) {
	tests := []struct {
		name    string
		points  []collision.Point
		area    float64
	}{
		{
			name:   "triangle",
			points: []collision.Point{{0, 0}, {3, 0}, {1.5, 2}},
			area:   3.0,
		},
		{
			name:   "square",
			points: []collision.Point{{0, 0}, {2, 0}, {2, 2}, {0, 2}},
			area:   4.0,
		},
		{
			name:   "empty polygon",
			points: []collision.Point{},
			area:   0,
		},
		{
			name:   "two points",
			points: []collision.Point{{0, 0}, {1, 1}},
			area:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			polygon := collision.NewPolygon(tt.points)
			
			assert.Equal(t, tt.points, polygon.Points)
			
			area := polygon.GetArea()
			assert.InDelta(t, tt.area, area, 0.0001)
		})
	}
}

func TestAABBCollision(t *testing.T) {
	tests := []struct {
		name     string
		aabb1    *collision.AABB
		aabb2    *collision.AABB
		expected bool
	}{
		{
			name:     "overlapping AABBs",
			aabb1:    collision.NewAABB(0, 0, 10, 10),
			aabb2:    collision.NewAABB(5, 5, 10, 10),
			expected: true,
		},
		{
			name:     "non-overlapping AABBs",
			aabb1:    collision.NewAABB(0, 0, 5, 5),
			aabb2:    collision.NewAABB(10, 10, 5, 5),
			expected: false,
		},
		{
			name:     "touching AABBs",
			aabb1:    collision.NewAABB(0, 0, 5, 5),
			aabb2:    collision.NewAABB(5, 0, 5, 5),
			expected: true,
		},
		{
			name:     "contained AABB",
			aabb1:    collision.NewAABB(0, 0, 20, 20),
			aabb2:    collision.NewAABB(5, 5, 5, 5),
			expected: true,
		},
		{
			name:     "zero size AABBs",
			aabb1:    collision.NewAABB(0, 0, 0, 0),
			aabb2:    collision.NewAABB(0, 0, 0, 0),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckAABBCollision(tt.aabb1, tt.aabb2)
			assert.Equal(t, tt.expected, result)
			
			// Test symmetry
			result2 := collision.CheckAABBCollision(tt.aabb2, tt.aabb1)
			assert.Equal(t, tt.expected, result2)
		})
	}
}

func TestCircleCollision(t *testing.T) {
	tests := []struct {
		name     string
		circle1  *collision.Circle
		circle2  *collision.Circle
		expected bool
	}{
		{
			name:     "overlapping circles",
			circle1:  collision.NewCircle(0, 0, 5),
			circle2:  collision.NewCircle(3, 3, 4),
			expected: true,
		},
		{
			name:     "non-overlapping circles",
			circle1:  collision.NewCircle(0, 0, 2),
			circle2:  collision.NewCircle(10, 10, 2),
			expected: false,
		},
		{
			name:     "touching circles",
			circle1:  collision.NewCircle(0, 0, 3),
			circle2:  collision.NewCircle(6, 0, 3),
			expected: true,
		},
		{
			name:     "zero radius circles",
			circle1:  collision.NewCircle(0, 0, 0),
			circle2:  collision.NewCircle(0, 0, 0),
			expected: true,
		},
		{
			name:     "contained circle",
			circle1:  collision.NewCircle(0, 0, 10),
			circle2:  collision.NewCircle(0, 0, 5),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckCircleCollision(tt.circle1, tt.circle2)
			assert.Equal(t, tt.expected, result)
			
			// Test symmetry
			result2 := collision.CheckCircleCollision(tt.circle2, tt.circle1)
			assert.Equal(t, tt.expected, result2)
		})
	}
}

func TestAABBCircleCollision(t *testing.T) {
	tests := []struct {
		name    string
		aabb    *collision.AABB
		circle  *collision.Circle
		expected bool
	}{
		{
			name:    "circle inside AABB",
			aabb:    collision.NewAABB(0, 0, 10, 10),
			circle:  collision.NewCircle(5, 5, 2),
			expected: true,
		},
		{
			name:    "circle outside AABB",
			aabb:    collision.NewAABB(0, 0, 5, 5),
			circle:  collision.NewCircle(10, 10, 2),
			expected: false,
		},
		{
			name:    "circle touching AABB edge",
			aabb:    collision.NewAABB(0, 0, 5, 5),
			circle:  collision.NewCircle(7, 2.5, 2),
			expected: true,
		},
		{
			name:    "circle partially overlapping AABB",
			aabb:    collision.NewAABB(0, 0, 5, 5),
			circle:  collision.NewCircle(3, 3, 3),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckAABBCircleCollision(tt.aabb, tt.circle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPointInAABB(t *testing.T) {
	aabb := collision.NewAABB(0, 0, 10, 10)
	
	tests := []struct {
		name     string
		point    *collision.Point
		expected bool
	}{
		{"point inside", collision.NewPoint(5, 5), true},
		{"point on edge", collision.NewPoint(0, 5), true},
		{"point on corner", collision.NewPoint(0, 0), true},
		{"point outside", collision.NewPoint(15, 15), false},
		{"point on right edge", collision.NewPoint(10, 5), true},
		{"point on bottom edge", collision.NewPoint(5, 10), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckPointInAABB(tt.point, aabb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPointInCircle(t *testing.T) {
	circle := collision.NewCircle(0, 0, 5)
	
	tests := []struct {
		name     string
		point    *collision.Point
		expected bool
	}{
		{"point inside", collision.NewPoint(2, 2), true},
		{"point on edge", collision.NewPoint(5, 0), true},
		{"point at center", collision.NewPoint(0, 0), true},
		{"point outside", collision.NewPoint(6, 6), false},
		{"point on circumference", collision.NewPoint(3, 4), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckPointInCircle(tt.point, circle)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPointInPolygon(t *testing.T) {
	// Square polygon
	square := collision.NewPolygon([]collision.Point{
		{0, 0}, {2, 0}, {2, 2}, {0, 2},
	})
	
	// Triangle polygon
	triangle := collision.NewPolygon([]collision.Point{
		{0, 0}, {3, 0}, {1.5, 2},
	})
	
	tests := []struct {
		name     string
		polygon  *collision.Polygon
		point    *collision.Point
		expected bool
	}{
		{"point inside square", square, collision.NewPoint(1, 1), true},
		{"point outside square", square, collision.NewPoint(3, 3), false},
		{"point on square edge", square, collision.NewPoint(1, 0), true},
		{"point inside triangle", triangle, collision.NewPoint(1.5, 1), true},
		{"point outside triangle", triangle, collision.NewPoint(4, 4), false},
		{"point on triangle vertex", triangle, collision.NewPoint(0, 0), true},
		{"empty polygon", collision.NewPolygon([]collision.Point{}), collision.NewPoint(0, 0), false},
		{"two point polygon", collision.NewPolygon([]collision.Point{{0, 0}, {1, 1}}), collision.NewPoint(0.5, 0.5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.CheckPointInPolygon(tt.point, tt.polygon)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		p1       *collision.Point
		p2       *collision.Point
		expected float64
	}{
		{
			name:     "same point",
			p1:       collision.NewPoint(0, 0),
			p2:       collision.NewPoint(0, 0),
			expected: 0,
		},
		{
			name:     "horizontal distance",
			p1:       collision.NewPoint(0, 0),
			p2:       collision.NewPoint(3, 0),
			expected: 3,
		},
		{
			name:     "vertical distance",
			p1:       collision.NewPoint(0, 0),
			p2:       collision.NewPoint(0, 4),
			expected: 4,
		},
		{
			name:     "diagonal distance",
			p1:       collision.NewPoint(0, 0),
			p2:       collision.NewPoint(3, 4),
			expected: 5, // 3-4-5 triangle
		},
		{
			name:     "negative coordinates",
			p1:       collision.NewPoint(-1, -1),
			p2:       collision.NewPoint(2, 3),
			expected: 5, // 3-4-5 triangle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.Distance(tt.p1, tt.p2)
			assert.InDelta(t, tt.expected, result, 0.0001)
		})
	}
}

func TestClosestPointOnAABB(t *testing.T) {
	aabb := collision.NewAABB(0, 0, 10, 10)
	
	tests := []struct {
		name     string
		point    *collision.Point
		expected *collision.Point
	}{
		{
			name:     "point inside",
			point:    collision.NewPoint(5, 5),
			expected: collision.NewPoint(5, 5),
		},
		{
			name:     "point outside top-left",
			point:    collision.NewPoint(-2, -3),
			expected: collision.NewPoint(0, 0),
		},
		{
			name:     "point outside bottom-right",
			point:    collision.NewPoint(15, 12),
			expected: collision.NewPoint(10, 10),
		},
		{
			name:     "point outside right",
			point:    collision.NewPoint(12, 5),
			expected: collision.NewPoint(10, 5),
		},
		{
			name:     "point outside bottom",
			point:    collision.NewPoint(5, 12),
			expected: collision.NewPoint(5, 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collision.ClosestPointOnAABB(tt.point, aabb)
			assert.Equal(t, tt.expected.X, result.X)
			assert.Equal(t, tt.expected.Y, result.Y)
		})
	}
}

// Benchmark tests
func BenchmarkAABBCollision(b *testing.B) {
	aabb1 := collision.NewAABB(0, 0, 10, 10)
	aabb2 := collision.NewAABB(5, 5, 10, 10)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collision.CheckAABBCollision(aabb1, aabb2)
	}
}

func BenchmarkCircleCollision(b *testing.B) {
	circle1 := collision.NewCircle(0, 0, 5)
	circle2 := collision.NewCircle(3, 3, 4)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collision.CheckCircleCollision(circle1, circle2)
	}
}

func BenchmarkPointInPolygon(b *testing.B) {
	polygon := collision.NewPolygon([]collision.Point{
		{0, 0}, {10, 0}, {10, 10}, {0, 10},
	})
	point := collision.NewPoint(5, 5)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collision.CheckPointInPolygon(point, polygon)
	}
}

// Property-based tests using testutils
func TestCollisionProperties(t *testing.T) {
	tdg := utils.NewTestDataGenerator()
	
	t.Run("AABB collision symmetry", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			x1 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			y1 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			w1 := float64(tdg.GenerateRandomIntArray(1, 50)[0])
			h1 := float64(tdg.GenerateRandomIntArray(1, 50)[0])
			
			x2 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			y2 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			w2 := float64(tdg.GenerateRandomIntArray(1, 50)[0])
			h2 := float64(tdg.GenerateRandomIntArray(1, 50)[0])
			
			aabb1 := collision.NewAABB(x1, y1, w1, h1)
			aabb2 := collision.NewAABB(x2, y2, w2, h2)
			
			result1 := collision.CheckAABBCollision(aabb1, aabb2)
			result2 := collision.CheckAABBCollision(aabb2, aabb1)
			
			assert.Equal(t, result1, result2, "AABB collision should be symmetric")
		}
	})
	
	t.Run("Circle collision symmetry", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			x1 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			y1 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			r1 := float64(tdg.GenerateRandomIntArray(1, 20)[0])
			
			x2 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			y2 := float64(tdg.GenerateRandomIntArray(1, 100)[0])
			r2 := float64(tdg.GenerateRandomIntArray(1, 20)[0])
			
			circle1 := collision.NewCircle(x1, y1, r1)
			circle2 := collision.NewCircle(x2, y2, r2)
			
			result1 := collision.CheckCircleCollision(circle1, circle2)
			result2 := collision.CheckCircleCollision(circle2, circle1)
			
			assert.Equal(t, result1, result2, "Circle collision should be symmetric")
		}
	})
}
