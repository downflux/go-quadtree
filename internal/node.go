package node

import (
	"fmt"

	"github.com/downflux/go-geometry/epsilon"
	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-pq/pq"
	"github.com/downflux/go-quadtree/id"
)

type Child uint

const (
	ChildNE Child = iota
	ChildSE
	ChildSW
	ChildNW
)

type next struct {
	child Child
	edge  Edge
}

var (
	// fsm is a transition table for looking up cell neighbors. See Yoder
	// 2006 for more information. Here, R, L, D, U are replaced with EdgeE,
	// EdgeW, EdgeS, and EdgeN respectively, and quadrants 0, 1, 2, 3 are
	// replaced with ChildNW, ChildNE, ChildSW, and ChildSE respectively.
	fsm = map[Edge]map[Child]next{
		EdgeE: map[Child]next{
			ChildNE: next{Child: ChildNW, Edge: EdgeE},
			ChildSE: next{Child: ChildSW, Edge: EdgeE},
			ChildSW: next{Child: ChildSE, Edge: EdgeNone},
			ChildNW: next{Child: ChildNE, Edge: EdgeNone},
		},
		EdgeW: map[Child]next{
			ChildNE: next{Child: ChildNW, Edge: EdgeNone},
			ChildSE: next{Child: ChildSW, Edge: EdgeNone},
			ChildSW: next{Child: ChildSE, Edge: EdgeW},
			ChildNW: next{Child: ChildNE, Edge: EdgeW},
		},
		EdgeS: map[Child]next{
			ChildNE: next{Child: ChildSE, Edge: EdgeNone},
			ChildSE: next{Child: ChildNE, Edge: EdgeS},
			ChildSW: next{Child: ChildNW, Edge: EdgeS},
			ChildNW: next{Child: ChildSW, Edge: EdgeNone},
		},
		EdgeN: map[Child]next{
			ChildNE: next{Child: ChildSE, Edge: EdgeN},
			ChildSE: next{Child: ChildNE, Edge: EdgeNone},
			ChildSW: next{Child: ChildNW, Edge: EdgeNone},
			ChildNW: next{Child: ChildSW, Edge: EdgeN},
		},
	}
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
	tolerance float64
	floor     int

	depth int

	parent *N

	// corner refers to the appropriate child from the parent node.
	corner Child

	children [4]*N
	aabb     hyperrectangle.R

	lookup map[id.ID]bool
}

// New returns a root node.
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
		path[m.depth-1] = m.corner
	}
	return path
}

func (n *N) IsLeaf() bool { return n.children[ChildNE] == nil }

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

func (n *N) Neighbors() *N {
	return nil
}
