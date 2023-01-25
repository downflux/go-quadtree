package node

import (
	"fmt"

	"github.com/downflux/go-geometry/epsilon"
	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-pq/pq"
	"github.com/downflux/go-quadtree/id"
)

type next struct {
	quadrant Child
	next     Edge
}

var (
	// fsm is a transition table for looking up cell neighbors. See Yoder
	// 2006 for more information. Here, R, L, D, U are replaced with EdgeE,
	// EdgeW, EdgeS, and EdgeN respectively, and quadrants 0, 1, 2, 3 are
	// replaced with ChildNW, ChildNE, ChildSW, and ChildSE respectively.
	fsm = map[Edge]map[Child]next{
		EdgeE: map[Child]next{
			ChildNE: next{quadrant: ChildNW, next: EdgeE},
			ChildSE: next{quadrant: ChildSW, next: EdgeE},
			ChildSW: next{quadrant: ChildSE, next: EdgeNone},
			ChildNW: next{quadrant: ChildNE, next: EdgeNone},
		},
		EdgeW: map[Child]next{
			ChildNE: next{quadrant: ChildNW, next: EdgeNone},
			ChildSE: next{quadrant: ChildSW, next: EdgeNone},
			ChildSW: next{quadrant: ChildSE, next: EdgeW},
			ChildNW: next{quadrant: ChildNE, next: EdgeW},
		},
		EdgeS: map[Child]next{
			ChildNE: next{quadrant: ChildSE, next: EdgeNone},
			ChildSE: next{quadrant: ChildNE, next: EdgeS},
			ChildSW: next{quadrant: ChildNW, next: EdgeS},
			ChildNW: next{quadrant: ChildSW, next: EdgeNone},
		},
		EdgeN: map[Child]next{
			ChildNE: next{quadrant: ChildSE, next: EdgeN},
			ChildSE: next{quadrant: ChildNE, next: EdgeNone},
			ChildSW: next{quadrant: ChildNW, next: EdgeNone},
			ChildNW: next{quadrant: ChildSW, next: EdgeN},
		},
	}
)

func FSM(path []Child, e Edge) []Child {
	if e != EdgeN && e != EdgeE && e != EdgeS && e != EdgeW {
		panic(fmt.Sprintf("invalid edge value %v", e))
	}

	buf := make([]Child, len(path))
	copy(buf, path)
	for i := 0; i < len(path) && e != EdgeNone; i++ {
		j := len(path) - i - 1
		next := fsm[e][path[j]]
		e, buf[j] = next.next, next.quadrant
	}
	// Check for edge condition -- if the FSM is still expecting an
	// ancestor but we have ran out of nodes, then the input node lies on
	// the edge of the tree, and the input edge asks for a node outside the
	// tree, e.g. asking for a southern neighbor when the input path is
	// already on the southern edge.
	if e != EdgeNone {
		return nil
	}

	return buf
}

type Child uint

const (
	ChildNE Child = iota
	ChildSE
	ChildSW
	ChildNW

	ChildNone
)

func (c Child) String() string {
	switch c {
	case ChildNE:
		return "0"
	case ChildSE:
		return "1"
	case ChildSW:
		return "2"
	case ChildNW:
		return "3"
	default:
		panic(fmt.Sprintf("invalid child"))
	}
}

type Edge uint

const (
	EdgeN Edge = iota
	EdgeE
	EdgeS
	EdgeW
	EdgeNE
	EdgeSE
	EdgeSW
	EdgeNW

	EdgeNone
)

func (e Edge) Invert() Edge {
	switch e {
	case EdgeN:
		return EdgeS
	case EdgeNE:
		return EdgeSW
	case EdgeE:
		return EdgeW
	case EdgeSE:
		return EdgeNW
	case EdgeS:
		return EdgeN
	case EdgeSW:
		return EdgeNE
	case EdgeW:
		return EdgeE
	case EdgeNW:
		return EdgeSE
	default:
		panic(fmt.Sprintf("invalid edge value %v", e))
	}
}

type N struct {
	tolerance float64
	floor     int

	depth int

	parent *N

	// corner refers to the appropriate child from the parent
	corner Child

	children  [4]*N
	aabb      hyperrectangle.R
	cachePath []Child
	cacheID   string

	lookup map[id.ID]bool
}

// New returns a root
func New(aabb hyperrectangle.R, tolerance float64, floor int) *N {
	if floor <= 0 {
		panic("floor must be a positive integer")
	}

	return &N{
		aabb:      aabb,
		tolerance: tolerance,
		floor:     floor,
	}
}

func Get(n *N, path []Child) *N {
	for _, c := range path {
		if m := n.children[c]; m != nil {
			n = m
		} else {
			break
		}
	}
	return n
}

func Path(n *N) []Child {
	path := make([]Child, n.depth)
	for m := n; m.parent != nil; m = m.parent {
		path[m.depth-1] = m.corner
	}
	return path
}

func ID(path []Child) string {
	var buf string
	for _, c := range path {
		buf += c.String()
	}
	return buf
}

func (n *N) Path() []Child { return n.cachePath }
func (n *N) ID() string    { return n.cacheID }
func (n *N) IsLeaf() bool  { return n.children[ChildNE] == nil }

func (n *N) Edge(e Edge) []*N {
	children := make([]*N, 0, 16)

	open := []*N{n}
	var m *N
	for len(open) > 0 {
		m, open = open[0], open[1:]
		if m.IsLeaf() {
			children = append(children, m)
			continue
		}

		switch e {
		case EdgeN:
			open = append(open, m.children[ChildNE], m.children[ChildNW])
		case EdgeE:
			open = append(open, m.children[ChildNE], m.children[ChildSE])
		case EdgeS:
			open = append(open, m.children[ChildSE], m.children[ChildSW])
		case EdgeW:
			open = append(open, m.children[ChildSW], m.children[ChildNW])
		case EdgeNE:
			open = append(open, m.children[ChildNE])
		case EdgeSE:
			open = append(open, m.children[ChildSE])
		case EdgeSW:
			open = append(open, m.children[ChildSW])
		case EdgeNW:
			open = append(open, m.children[ChildNW])
		default:
			panic(fmt.Sprintf("invalid edge %v", e))
		}
	}

	return children
}

func (n *N) split(data map[id.ID]hyperrectangle.R) {
	if n.depth == n.floor {
		panic("cannot split past the depth limit")
	}

	if !n.IsLeaf() {
		panic("cannot split a non-leaf node")
	}

	xmin, ymin := n.aabb.Min().X(vector.AXIS_X), n.aabb.Min().X(vector.AXIS_Y)
	xmax, ymax := n.aabb.Max().X(vector.AXIS_X), n.aabb.Max().X(vector.AXIS_Y)

	xmid, ymid := xmin+(xmax-xmin)/2, ymin+(ymax-ymin)/2

	n.children[ChildNE] = &N{
		corner: ChildNE,
		aabb:   *hyperrectangle.New(vector.V{xmid, ymid}, vector.V{xmax, ymax}),
	}
	n.children[ChildSE] = &N{
		corner: ChildSE,
		aabb:   *hyperrectangle.New(vector.V{xmid, ymin}, vector.V{xmax, ymid}),
	}
	n.children[ChildSW] = &N{
		corner: ChildSW,
		aabb:   *hyperrectangle.New(vector.V{xmin, ymin}, vector.V{xmid, ymid}),
	}
	n.children[ChildNW] = &N{
		corner: ChildNW,
		aabb:   *hyperrectangle.New(vector.V{xmin, ymid}, vector.V{xmid, ymax}),
	}

	for _, c := range n.children {
		c.depth = n.depth + 1
		c.parent = n
		c.lookup = make(map[id.ID]bool, len(n.lookup))
		c.tolerance = n.tolerance
		c.floor = n.floor
		c.cachePath = Path(c)
		c.cacheID = ID(c.cachePath)

		for x := range n.lookup {
			aabb := data[x]
			if !hyperrectangle.Disjoint(c.aabb, aabb) {
				c.lookup[x] = true
			}
		}
	}

	n.lookup = map[id.ID]bool{}
}

func (n *N) Insert(x id.ID, data map[id.ID]hyperrectangle.R) {
	aabb := data[x]

	open := []*N{n}
	var m *N
	for len(open) > 0 {
		m, open = open[0], open[1:]

		if hyperrectangle.Disjoint(aabb, m.aabb) {
			continue
		}

		if !m.IsLeaf() {
			open = append(
				open,
				m.children[ChildNE],
				m.children[ChildSE],
				m.children[ChildSW],
				m.children[ChildNW],
			)
			continue
		}

		if m.depth >= m.floor || epsilon.Absolute(m.tolerance).Within(
			hyperrectangle.V(m.aabb),
			hyperrectangle.V(aabb),
		) {
			m.lookup[x] = true
		} else {
			m.split(data)
			open = append(
				open,
				m.children[ChildNE],
				m.children[ChildSE],
				m.children[ChildSW],
				m.children[ChildNW],
			)
		}
	}
}

func (n *N) Remove(x id.ID, data map[id.ID]hyperrectangle.R) {
	aabb := data[x]

	candidates := pq.New[*N](0, pq.PMax)

	open := []*N{n}
	var m *N
	for len(open) > 0 {
		m, open = open[0], open[1:]

		if hyperrectangle.Disjoint(aabb, m.aabb) {
			continue
		}

		if !m.IsLeaf() {
			open = append(
				open,
				m.children[ChildNE],
				m.children[ChildSE],
				m.children[ChildSW],
				m.children[ChildNW],
			)
			continue
		}

		if m.lookup[x] {
			delete(m.lookup, x)
		}

		if len(m.lookup) == 0 {
			candidates.Push(m, float64(m.depth))
		}
	}

	// Collapse parents.
	for !candidates.Empty() {
		m, _ = candidates.Pop()
		if p := m.parent; p != nil {
			empty := true
			for _, c := range p.children {
				if len(c.lookup) != 0 {
					empty = false
				}
			}

			if empty {
				for x, c := range p.children {
					c.parent = nil
					p.children[x] = nil
				}
				candidates.Push(p, float64(p.depth))
			}
		}
	}
}

func (n *N) Root() *N {
	var m *N
	for m = n; m.parent != nil; m = m.parent {
	}
	return m
}

func (n *N) Neighbors() []*N {
	root := n.Root()
	p := n.Path()

	paths := make([][]Child, 8)

	paths[EdgeN] = FSM(p, EdgeN)
	paths[EdgeE] = FSM(p, EdgeE)
	paths[EdgeS] = FSM(p, EdgeS)
	paths[EdgeW] = FSM(p, EdgeW)

	paths[EdgeNE] = FSM(paths[EdgeN], EdgeE)
	paths[EdgeSE] = FSM(paths[EdgeS], EdgeE)
	paths[EdgeSW] = FSM(paths[EdgeS], EdgeW)
	paths[EdgeNW] = FSM(paths[EdgeN], EdgeW)

	ns := make([]*N, 0, 16)
	ids := make(map[string]bool, 16)
	for e, p := range paths {
		if len(p) > 0 {
			for _, m := range Get(root, p).Edge(Edge(e).Invert()) {
				if _, ok := ids[m.ID()]; !ok {
					ns = append(ns, m)
				}
				ids[m.ID()] = true
			}
		}
	}
	return ns
}
