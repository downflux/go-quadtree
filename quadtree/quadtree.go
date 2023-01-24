package quadtree

import (
	"fmt"

	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-quadtree/id"
	"github.com/downflux/go-quadtree/internal/node"
)

type next struct {
	quadrant node.Child
	next     node.Edge
}

var (
	// fsm is a transition table for looking up cell neighbors. See Yoder
	// 2006 for more information. Here, R, L, D, U are replaced with EdgeE,
	// EdgeW, EdgeS, and EdgeN respectively, and quadrants 0, 1, 2, 3 are
	// replaced with ChildNW, ChildNE, ChildSW, and ChildSE respectively.
	fsm = map[node.Edge]map[node.Child]next{
		node.EdgeE: map[node.Child]next{
			node.ChildNE: next{quadrant: node.ChildNW, next: node.EdgeE},
			node.ChildSE: next{quadrant: node.ChildSW, next: node.EdgeE},
			node.ChildSW: next{quadrant: node.ChildSE, next: node.EdgeNone},
			node.ChildNW: next{quadrant: node.ChildNE, next: node.EdgeNone},
		},
		node.EdgeW: map[node.Child]next{
			node.ChildNE: next{quadrant: node.ChildNW, next: node.EdgeNone},
			node.ChildSE: next{quadrant: node.ChildSW, next: node.EdgeNone},
			node.ChildSW: next{quadrant: node.ChildSE, next: node.EdgeW},
			node.ChildNW: next{quadrant: node.ChildNE, next: node.EdgeW},
		},
		node.EdgeS: map[node.Child]next{
			node.ChildNE: next{quadrant: node.ChildSE, next: node.EdgeNone},
			node.ChildSE: next{quadrant: node.ChildNE, next: node.EdgeS},
			node.ChildSW: next{quadrant: node.ChildNW, next: node.EdgeS},
			node.ChildNW: next{quadrant: node.ChildSW, next: node.EdgeNone},
		},
		node.EdgeN: map[node.Child]next{
			node.ChildNE: next{quadrant: node.ChildSE, next: node.EdgeN},
			node.ChildSE: next{quadrant: node.ChildNE, next: node.EdgeNone},
			node.ChildSW: next{quadrant: node.ChildNW, next: node.EdgeNone},
			node.ChildNW: next{quadrant: node.ChildSW, next: node.EdgeN},
		},
	}
)

type QT struct {
	root *node.N

	aabb map[id.ID]hyperrectangle.R
}

func New(bounds hyperrectangle.R, tolerance float64, floor int) *QT {
	buf := hyperrectangle.New(vector.V{0, 0}, vector.V{0, 0}).M()
	buf.Copy(bounds)
	return &QT{
		root: node.New(buf.R(), tolerance, floor),
		aabb: make(map[id.ID]hyperrectangle.R, 128),
	}
}

func (qt *QT) Insert(x id.ID, aabb hyperrectangle.R) error {
	if _, ok := qt.aabb[x]; ok {
		return fmt.Errorf("cannot insert duplicate key %v", x)
	}

	buf := hyperrectangle.New(vector.V{0, 0}, vector.V{0, 0}).M()
	buf.Copy(aabb)

	qt.aabb[x] = buf.R()
	qt.root.Insert(x, qt.aabb)

	return nil
}

func (qt *QT) Remove(x id.ID) error {
	if _, ok := qt.aabb[x]; !ok {
		return fmt.Errorf("cannot remove non-existent key %v", x)
	}

	qt.root.Remove(x, qt.aabb)
	delete(qt.aabb, x)

	return nil
}

func (qt *QT) Path(s vector.V, g vector.V) []vector.V {
	return nil
}
