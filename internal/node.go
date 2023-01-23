package node

import (
	"fmt"

	"github.com/downflux/go-geometry/epsilon"
	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-pq/pq"
	"github.com/downflux/go-quadtree/id"
)

const (
	depthLimit = 10
	tolerance  = 1
)

type Child uint

const (
	ChildNE Child = iota
	ChildSE
	ChildSW
	ChildNW
)

type Edge uint

const (
	EdgeNone Edge = 0

	EdgeN Edge = 1 << iota
	EdgeE
	EdgeS
	EdgeW

	EdgeNE = EdgeN | EdgeE
	EdgeSE = EdgeS | EdgeE
	EdgeSW = EdgeS | EdgeW
	EdgeNW = EdgeN | EdgeW
)

type N struct {
	depth int

	parent *N

	// corner refers to the appropriate child from the parent node.
	corner Child

	children [4]*N
	aabb     hyperrectangle.R

	lookup map[id.ID]bool
}

// New returns a root node.
func New(aabb hyperrectangle.R) *N {
	return &N{
		aabb: aabb,
	}
}

func (n *N) Get(path []Child) *N {
	m := n
	for _, c := range path {
		m = m.children[c]
	}
	return m
}

func (n *N) Path() []Child {
	path := make([]Child, n.depth)
	for m := n; m.parent != nil; m = m.parent {
		path[m.depth] = m.corner
	}
	return path
}

func (n *N) IsLeaf() bool { return n.children[ChildNE] == nil }

func (n *N) Edge(c Edge) []*N {
	children := make([]*N, 0, 16)

	open := []*N{n}
	var m *N
	for len(open) > 0 {
		m, open = open[0], open[1:]
		if m.IsLeaf() {
			children = append(children, m)
			continue
		}

		switch c {
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
			panic(fmt.Sprintf("invalid edge %v", c))
		}
	}

	return children
}

func (n *N) split(p vector.V, data map[id.ID]hyperrectangle.R) {
	if n.depth == depthLimit {
		panic("cannot split past the depth limit")
	}

	if !n.IsLeaf() {
		panic("cannot split a non-leaf node")
	}

	xmin, ymin := n.aabb.Min().X(vector.AXIS_X), n.aabb.Min().X(vector.AXIS_Y)
	xmax, ymax := n.aabb.Max().X(vector.AXIS_X), n.aabb.Max().X(vector.AXIS_Y)

	px, py := p.X(vector.AXIS_X), p.X(vector.AXIS_Y)

	n.children[ChildNE] = &N{
		corner: ChildNE,
		aabb:   *hyperrectangle.New(vector.V{px, py}, vector.V{xmax, ymax}),
	}
	n.children[ChildSE] = &N{
		corner: ChildSE,
		aabb:   *hyperrectangle.New(vector.V{px, ymin}, vector.V{xmax, py}),
	}
	n.children[ChildSW] = &N{
		corner: ChildSW,
		aabb:   *hyperrectangle.New(vector.V{xmin, ymin}, vector.V{px, py}),
	}
	n.children[ChildNW] = &N{
		corner: ChildNW,
		aabb:   *hyperrectangle.New(vector.V{xmin, py}, vector.V{px, ymax}),
	}

	for _, c := range n.children {
		c.depth = n.depth + 1
		c.parent = n
		c.lookup = make(map[id.ID]bool, len(n.lookup))

		for x := range n.lookup {
			aabb := data[x]
			if !hyperrectangle.Disjoint(c.aabb, aabb) {
				c.lookup[x] = true
			}
		}
	}

	n.lookup = make(map[id.ID]bool, 16)
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

		if m.depth >= depthLimit || epsilon.Absolute(tolerance).Within(
			hyperrectangle.V(m.aabb),
			hyperrectangle.V(aabb),
		) {
			m.lookup[x] = true
		} else {
			xmin, ymin := n.aabb.Min().X(vector.AXIS_X), n.aabb.Min().X(vector.AXIS_Y)
			xmax, ymax := n.aabb.Max().X(vector.AXIS_X), n.aabb.Max().X(vector.AXIS_Y)

			xmid, ymid := xmin+(xmax-xmin)/2, ymin+(ymax-ymin)/2
			m.split(vector.V{xmid, ymid}, data)
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
