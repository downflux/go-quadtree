# go-quadtree
Golang quadtree implementation

## API


```
const (
	tolerance = ...
	depthLimit = ...
)

type Child int

const (
	ChildNone Child = iota
	ChildNE
	ChildSE
	ChildSw
	ChildNW
)

type datum[T any] struct {
	aabb  hyperrectangle.R
	datum T
}

type QT[T any] struct {
	root *Node

	data map[id.ID]datum
}

type Node struct {
	depth int
	parent *Node
	children [4]*Node
	bounds hyperrectangle.R

	lookup []id.ID
}

func (n *N) IsLeaf() bool { return children[ChildNE] == nil }

func (qt *QT[T]) Insert(x id.ID, aabb hyperrectangle.R, d T) {
	buf := hyperrectangle.New(vector.V{0, 0}, vector.V{0, 0})
	buf.Copy(aabb)

	qt.data[x] = datum{
		aabb: buf.R(),
		datum: d,
	}

	open := []*node{root}
	var n *Node
	for len(open) > 0 {
		n, open = open[0], open[1:]

		if hyperrectangle.Disjoint(n.bounds, buf.R()) {
			continue
		}

		if !n.IsLeaf() {
			open = append(
				open,
				n.children[ChildNE],
				n.children[ChildSE],
				n.children[ChildSW],
				n.children[ChildNW],
			)
			continue
		}

		if n.depth >= depthLimit || epsilon.Absolute(tolerance).Within(
			hyperrectangle.V(qt.bounds),
			hyperrectangle.V(buf.R()),
		) {
			n.lookup[x] = append(n.lookup[x], n)
		} else {
			xmax, xmin := n.bounds.Max().X(vector.AXIS_X), n.bounds.Min().X(vector.AXIS_Y)
			ymax, ymin := n.bounds.Max().X(vector.AXIS_X), n.bounds.Min().X(vector.AXIS_Y)
			xmid, ymid := xmin + (xmax - xmin) / 2, ymin + (ymax - ymin) / 2

			n.children[ChildNE] = &node{
				depth: n.depth + 1,
				parent: n,
				bounds: hyperrectangle.New(
					vector.V{xmid, ymid},
					vector.V{xmax, ymax},
				)
			}
			n.children[ChildSE] = &node{
				depth: n.depth + 1,
				parent: n,
				bounds: hyperrectangle.New(
					vector.V{xmid, ymin},
					vector.V{xmax, ymid},
				)
			}
			n.children[ChildSW] = &node{
				depth: n.depth + 1,
				parent: n,
				bounds: hyperrectangle.New(
					vector.V{xmin, ymin},
					vector.V{xmid, ymid},
				)
			}
			n.children[ChildNW] = &node{
				depth: n.depth + 1,
				parent: n,
				bounds: hyperrectangle.New(
					vector.V{xmin, ymid}
					vector.V{xmid, ymax},
				)
			}

			data := n.lookup
			for _, x := range data {
				... // insert child
			}

			open = append(open, n.children[...])
		}
	}
	if !hyperrectangle.Disjoint(qt.bounds, aabb) {
		if qt.depth >= depthLimit || epsilon.Absolute(tolerance).Within(hyperrectangle.Volume(qt.aabb), hyperrectangle.V(qt.bounds)) {
			qt.data[x] = T
		}
		else {
			if qt.children[ChildNE] == (QT{}) {
				qt.children[ChildNE] = ...
			}
			...

			qt.children[ChildNE].Insert(x, aabb, data)
			...
		}				
	}
}

func (qt *QT[T]) Delete(x id.ID) {
	...
}
```
