package collision

import (
	"math"
)

// AABB represents an Axis-Aligned Bounding Box
type AABB struct {
	X, Y, Width, Height float64
}

// Circle represents a circle for collision detection
type Circle struct {
	X, Y, Radius float64
}

// Point represents a 2D point
type Point struct {
	X, Y float64
}

// Polygon represents a polygon for collision detection
type Polygon struct {
	Points []Point
}

// NewAABB creates a new AABB
func NewAABB(x, y, width, height float64) *AABB {
	return &AABB{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// NewCircle creates a new Circle
func NewCircle(x, y, radius float64) *Circle {
	return &Circle{
		X:      x,
		Y:      y,
		Radius: radius,
	}
}

// NewPoint creates a new Point
func NewPoint(x, y float64) *Point {
	return &Point{X: x, Y: y}
}

// NewPolygon creates a new Polygon
func NewPolygon(points []Point) *Polygon {
	return &Polygon{Points: points}
}

// CheckAABBCollision checks if two AABBs collide
func CheckAABBCollision(a, b *AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

// CheckCircleCollision checks if two circles collide
func CheckCircleCollision(a, b *Circle) bool {
	dx := a.X - b.X
	dy := a.Y - b.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < (a.Radius + b.Radius)
}

// CheckAABBCircleCollision checks if an AABB and circle collide
func CheckAABBCircleCollision(aabb *AABB, circle *Circle) bool {
	// Find the closest point on the AABB to the circle center
	closestX := circle.X
	closestY := circle.Y

	if circle.X < aabb.X {
		closestX = aabb.X
	} else if circle.X > aabb.X+aabb.Width {
		closestX = aabb.X + aabb.Width
	}

	if circle.Y < aabb.Y {
		closestY = aabb.Y
	} else if circle.Y > aabb.Y+aabb.Height {
		closestY = aabb.Y + aabb.Height
	}

	// Calculate distance between circle center and closest point
	dx := circle.X - closestX
	dy := circle.Y - closestY
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance < circle.Radius
}

// CheckPointInAABB checks if a point is inside an AABB
func CheckPointInAABB(point *Point, aabb *AABB) bool {
	return point.X >= aabb.X &&
		point.X <= aabb.X+aabb.Width &&
		point.Y >= aabb.Y &&
		point.Y <= aabb.Y+aabb.Height
}

// CheckPointInCircle checks if a point is inside a circle
func CheckPointInCircle(point *Point, circle *Circle) bool {
	dx := point.X - circle.X
	dy := point.Y - circle.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= circle.Radius
}

// CheckPointInPolygon checks if a point is inside a polygon using ray casting
func CheckPointInPolygon(point *Point, polygon *Polygon) bool {
	if len(polygon.Points) < 3 {
		return false
	}

	inside := false
	j := len(polygon.Points) - 1

	for i := 0; i < len(polygon.Points); i++ {
		if ((polygon.Points[i].Y > point.Y) != (polygon.Points[j].Y > point.Y)) &&
			(point.X < (polygon.Points[j].X-polygon.Points[i].X)*(point.Y-polygon.Points[i].Y)/(polygon.Points[j].Y-polygon.Points[i].Y)+polygon.Points[i].X) {
			inside = !inside
		}
		j = i
	}

	return inside
}

// GetAABBCenter returns the center point of an AABB
func (a *AABB) GetCenter() *Point {
	return &Point{
		X: a.X + a.Width/2,
		Y: a.Y + a.Height/2,
	}
}

// GetAABBArea returns the area of an AABB
func (a *AABB) GetArea() float64 {
	return a.Width * a.Height
}

// GetCircleArea returns the area of a circle
func (c *Circle) GetArea() float64 {
	return math.Pi * c.Radius * c.Radius
}

// GetPolygonArea returns the area of a polygon using the shoelace formula
func (p *Polygon) GetArea() float64 {
	if len(p.Points) < 3 {
		return 0
	}

	area := 0.0
	j := len(p.Points) - 1

	for i := 0; i < len(p.Points); i++ {
		area += (p.Points[j].X + p.Points[i].X) * (p.Points[j].Y - p.Points[i].Y)
		j = i
	}

	return math.Abs(area) / 2.0
}

// Distance calculates the distance between two points
func Distance(p1, p2 *Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// ClosestPointOnAABB finds the closest point on an AABB to a given point
func ClosestPointOnAABB(point *Point, aabb *AABB) *Point {
	closestX := point.X
	closestY := point.Y

	if point.X < aabb.X {
		closestX = aabb.X
	} else if point.X > aabb.X+aabb.Width {
		closestX = aabb.X + aabb.Width
	}

	if point.Y < aabb.Y {
		closestY = aabb.Y
	} else if point.Y > aabb.Y+aabb.Height {
		closestY = aabb.Y + aabb.Height
	}

	return &Point{X: closestX, Y: closestY}
}

