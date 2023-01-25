package quadtree

import (
	"fmt"

	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-quadtree/id"
	"github.com/downflux/go-quadtree/internal/node"
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
