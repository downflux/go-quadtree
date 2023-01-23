package node

import (
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
	ChildNone Child = 0

	ChildN Child = 1 << iota
	ChildE
	ChildS
	ChildW

	ChildNE = ChildN | ChildE
	ChildSE = ChildS | ChildE
	ChildSW = ChildS | ChildW
	ChildNW = ChildN | ChildW
)

type N struct {
	depth    int
	parent   *N
	children map[Child]*N
	aabb     hyperrectangle.R

	lookup map[id.ID]bool
}

func New(aabb hyperrectangle.R) *N {
	return &N{
		aabb:     aabb,
		children: make(map[Child]*N, 4),
	}
}

func (n *N) Edge(c Child) []*N {
	children := make([]*N, 0, 16)
	open := []*N{n}
	var m *N
	for len(open) > 0 {
		m, open = open[0], open[1:]
		if m.IsLeaf() {
			children = append(children, m)
			continue
		}

		if c&ChildN == ChildN {
			open = append(open, m.children[ChildN])
		}
		if c&ChildE == ChildE {
			open = append(open, m.children[ChildE])
		}
		if c&ChildS == ChildS {
			open = append(open, m.children[ChildS])
		}
		if c&ChildW == ChildW {
			open = append(open, m.children[ChildW])
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
		aabb: *hyperrectangle.New(vector.V{px, py}, vector.V{xmax, ymax}),
	}
	n.children[ChildSE] = &N{
		aabb: *hyperrectangle.New(vector.V{px, ymin}, vector.V{xmax, py}),
	}
	n.children[ChildSW] = &N{
		aabb: *hyperrectangle.New(vector.V{xmin, ymin}, vector.V{px, py}),
	}
	n.children[ChildNW] = &N{
		aabb: *hyperrectangle.New(vector.V{xmin, py}, vector.V{px, ymax}),
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

func (n *N) IsLeaf() bool { return n.children[ChildNE] == nil }

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
