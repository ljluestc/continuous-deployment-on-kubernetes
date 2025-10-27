package unionfind

// QuickFind implements the Quick Find algorithm
type QuickFind struct {
	id    []int
	count int
}

// QuickUnion implements the Quick Union algorithm
type QuickUnion struct {
	id    []int
	count int
}

// WeightedQuickUnion implements the Weighted Quick Union algorithm
type WeightedQuickUnion struct {
	id    []int
	sz    []int
	count int
}

// WeightedQuickUnionWithPathCompression implements Weighted Quick Union with path compression
type WeightedQuickUnionWithPathCompression struct {
	id    []int
	sz    []int
	count int
}

// UnionFind interface defines common operations
type UnionFind interface {
	Find(p int) int
	Union(p, q int)
	Connected(p, q int) bool
	Count() int
}

// NewQuickFind creates a new QuickFind instance
func NewQuickFind(n int) *QuickFind {
	id := make([]int, n)
	for i := range id {
		id[i] = i
	}
	return &QuickFind{
		id:    id,
		count: n,
	}
}

// Find returns the component identifier for p
func (qf *QuickFind) Find(p int) int {
	return qf.id[p]
}

// Union merges the component containing p with the component containing q
func (qf *QuickFind) Union(p, q int) {
	pID := qf.Find(p)
	qID := qf.Find(q)

	if pID == qID {
		return
	}

	for i := range qf.id {
		if qf.id[i] == pID {
			qf.id[i] = qID
		}
	}
	qf.count--
}

// Connected returns true if p and q are in the same component
func (qf *QuickFind) Connected(p, q int) bool {
	return qf.Find(p) == qf.Find(q)
}

// Count returns the number of components
func (qf *QuickFind) Count() int {
	return qf.count
}

// NewQuickUnion creates a new QuickUnion instance
func NewQuickUnion(n int) *QuickUnion {
	id := make([]int, n)
	for i := range id {
		id[i] = i
	}
	return &QuickUnion{
		id:    id,
		count: n,
	}
}

// Find returns the root of the component containing p
func (qu *QuickUnion) Find(p int) int {
	for p != qu.id[p] {
		p = qu.id[p]
	}
	return p
}

// Union merges the component containing p with the component containing q
func (qu *QuickUnion) Union(p, q int) {
	pRoot := qu.Find(p)
	qRoot := qu.Find(q)

	if pRoot == qRoot {
		return
	}

	qu.id[pRoot] = qRoot
	qu.count--
}

// Connected returns true if p and q are in the same component
func (qu *QuickUnion) Connected(p, q int) bool {
	return qu.Find(p) == qu.Find(q)
}

// Count returns the number of components
func (qu *QuickUnion) Count() int {
	return qu.count
}

// NewWeightedQuickUnion creates a new WeightedQuickUnion instance
func NewWeightedQuickUnion(n int) *WeightedQuickUnion {
	id := make([]int, n)
	sz := make([]int, n)
	for i := range id {
		id[i] = i
		sz[i] = 1
	}
	return &WeightedQuickUnion{
		id:    id,
		sz:    sz,
		count: n,
	}
}

// Find returns the root of the component containing p
func (wqu *WeightedQuickUnion) Find(p int) int {
	for p != wqu.id[p] {
		p = wqu.id[p]
	}
	return p
}

// Union merges the component containing p with the component containing q
func (wqu *WeightedQuickUnion) Union(p, q int) {
	pRoot := wqu.Find(p)
	qRoot := wqu.Find(q)

	if pRoot == qRoot {
		return
	}

	// Make smaller root point to larger one
	if wqu.sz[pRoot] < wqu.sz[qRoot] {
		wqu.id[pRoot] = qRoot
		wqu.sz[qRoot] += wqu.sz[pRoot]
	} else {
		wqu.id[qRoot] = pRoot
		wqu.sz[pRoot] += wqu.sz[qRoot]
	}
	wqu.count--
}

// Connected returns true if p and q are in the same component
func (wqu *WeightedQuickUnion) Connected(p, q int) bool {
	return wqu.Find(p) == wqu.Find(q)
}

// Count returns the number of components
func (wqu *WeightedQuickUnion) Count() int {
	return wqu.count
}

// NewWeightedQuickUnionWithPathCompression creates a new WeightedQuickUnionWithPathCompression instance
func NewWeightedQuickUnionWithPathCompression(n int) *WeightedQuickUnionWithPathCompression {
	id := make([]int, n)
	sz := make([]int, n)
	for i := range id {
		id[i] = i
		sz[i] = 1
	}
	return &WeightedQuickUnionWithPathCompression{
		id:    id,
		sz:    sz,
		count: n,
	}
}

// Find returns the root of the component containing p with path compression
func (wqupc *WeightedQuickUnionWithPathCompression) Find(p int) int {
	root := p
	for root != wqupc.id[root] {
		root = wqupc.id[root]
	}

	// Path compression: make all nodes point directly to root
	for p != root {
		next := wqupc.id[p]
		wqupc.id[p] = root
		p = next
	}

	return root
}

// Union merges the component containing p with the component containing q
func (wqupc *WeightedQuickUnionWithPathCompression) Union(p, q int) {
	pRoot := wqupc.Find(p)
	qRoot := wqupc.Find(q)

	if pRoot == qRoot {
		return
	}

	// Make smaller root point to larger one
	if wqupc.sz[pRoot] < wqupc.sz[qRoot] {
		wqupc.id[pRoot] = qRoot
		wqupc.sz[qRoot] += wqupc.sz[pRoot]
	} else {
		wqupc.id[qRoot] = pRoot
		wqupc.sz[pRoot] += wqupc.sz[qRoot]
	}
	wqupc.count--
}

// Connected returns true if p and q are in the same component
func (wqupc *WeightedQuickUnionWithPathCompression) Connected(p, q int) bool {
	return wqupc.Find(p) == wqupc.Find(q)
}

// Count returns the number of components
func (wqupc *WeightedQuickUnionWithPathCompression) Count() int {
	return wqupc.count
}

// GetComponentSize returns the size of the component containing p
func (wqupc *WeightedQuickUnionWithPathCompression) GetComponentSize(p int) int {
	root := wqupc.Find(p)
	return wqupc.sz[root]
}

// GetAllComponents returns all components as a map
func (wqupc *WeightedQuickUnionWithPathCompression) GetAllComponents() map[int][]int {
	components := make(map[int][]int)
	
	for i := 0; i < len(wqupc.id); i++ {
		root := wqupc.Find(i)
		components[root] = append(components[root], i)
	}
	
	return components
}

// IsValidIndex checks if the given index is valid
func (wqupc *WeightedQuickUnionWithPathCompression) IsValidIndex(p int) bool {
	return p >= 0 && p < len(wqupc.id)
}

// Reset resets the Union-Find structure to initial state
func (wqupc *WeightedQuickUnionWithPathCompression) Reset() {
	n := len(wqupc.id)
	for i := 0; i < n; i++ {
		wqupc.id[i] = i
		wqupc.sz[i] = 1
	}
	wqupc.count = n
}

