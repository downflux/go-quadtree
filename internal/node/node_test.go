package node

import (
	"testing"

	"github.com/downflux/go-geometry/nd/hyperrectangle"
	"github.com/downflux/go-geometry/nd/vector"
	"github.com/downflux/go-quadtree/id"
	"github.com/google/go-cmp/cmp"
)

var (
	opts = []cmp.Option{
		cmp.AllowUnexported(N{}, hyperrectangle.R{}),
	}
)

func TestEdge(t *testing.T) {
	type config struct {
		name string
		n    *N
		e    Edge
		want []*N
	}

	configs := []config{
		func() config {
			n := &N{}
			return config{
				name: "Trivial",
				n:    n,
				e:    EdgeNE,
				want: []*N{n},
			}
		}(),
	}
	configs = append(configs, func() []config {
		ne := &N{}
		se := &N{}
		sw := &N{}
		nw := &N{}
		root := &N{
			children: [4]*N{ne, se, sw, nw},
		}
		return []config{
			{
				name: "Simple/N",
				n:    root,
				e:    EdgeN,
				want: []*N{ne, nw},
			},
			{
				name: "Simple/NE",
				n:    root,
				e:    EdgeNE,
				want: []*N{ne},
			},
			{
				name: "Simple/E",
				n:    root,
				e:    EdgeE,
				want: []*N{ne, se},
			},
			{
				name: "Simple/SE",
				n:    root,
				e:    EdgeSE,
				want: []*N{se},
			},
			{
				name: "Simple/S",
				n:    root,
				e:    EdgeS,
				want: []*N{se, sw},
			},
			{
				name: "Simple/SW",
				n:    root,
				e:    EdgeSW,
				want: []*N{sw},
			},
			{
				name: "Simple/W",
				n:    root,
				e:    EdgeW,
				want: []*N{sw, nw},
			},
			{
				name: "Simple/NW",
				n:    root,
				e:    EdgeNW,
				want: []*N{nw},
			},
		}
	}()...,
	)

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			got := c.n.Edge(c.e)
			if diff := cmp.Diff(c.want, got, opts...); diff != "" {
				t.Errorf("Edge() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type config struct {
		name string
		n    *N
		path []Child
		want *N
	}

	configs := []config{
		func() config {
			child := &N{}
			root := &N{
				children: [4]*N{child, &N{}, &N{}, &N{}},
			}
			return config{
				name: "Simple",
				n:    root,
				path: []Child{ChildNE},
				want: child,
			}
		}(),
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			got := Get(c.n, c.path)
			if diff := cmp.Diff(c.want, got, opts...); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestPath(t *testing.T) {
	type config struct {
		name string
		n    *N
		want []Child
	}

	configs := []config{
		{
			name: "Root",
			n:    &N{},
			want: []Child{},
		},
		func() config {
			root := &N{}
			child := &N{
				depth: 1,

				parent: root,
				corner: ChildNE,
			}
			return config{
				name: "Simple",
				n:    child,
				want: []Child{ChildNE},
			}
		}(),
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			got := Path(c.n)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("Path() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestIsLeaf(t *testing.T) {
	type config struct {
		name string
		n    *N
		want bool
	}

	configs := []config{
		{
			name: "Simple",
			n:    &N{},
			want: true,
		},
		{
			name: "Simple/NoLeaf",
			n: &N{
				children: [4]*N{&N{}, &N{}, &N{}, &N{}},
			},
		},
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			if got := c.n.IsLeaf(); got != c.want {
				t.Errorf("IsLeaf() = %v, want = %v", got, c.want)
			}
		})
	}
}

func TestSplit(t *testing.T) {
	type config struct {
		name string
		n    *N
		data map[id.ID]hyperrectangle.R
		want *N
	}

	configs := []config{
		func() config {
			want := &N{
				aabb: *hyperrectangle.New(
					vector.V{0, 0},
					vector.V{100, 100},
				),
				lookup: map[id.ID]bool{},
				floor:  1,
			}
			want.children = [4]*N{
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildNE,
					aabb:      *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
					lookup:    map[id.ID]bool{},
					floor:     1,
					cachePath: []Child{ChildNE},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildSE,
					aabb:      *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
					lookup:    map[id.ID]bool{},
					floor:     1,
					cachePath: []Child{ChildSE},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildSW,
					aabb:      *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					lookup:    map[id.ID]bool{100: true},
					floor:     1,
					cachePath: []Child{ChildSW},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildNW,
					aabb:      *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
					lookup:    map[id.ID]bool{100: true},
					floor:     1,
					cachePath: []Child{ChildNW},
				},
			}
			data := map[id.ID]hyperrectangle.R{
				100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 100}),
			}
			return config{
				name: "Trivial",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{100: true},
					floor:  1,
				},
				data: data,
				want: want,
			}
		}(),
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			c.n.split(c.data)
			if diff := cmp.Diff(c.want, c.n, opts...); diff != "" {
				t.Errorf("split() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	type config struct {
		name string
		n    *N
		x    id.ID
		data map[id.ID]hyperrectangle.R
		want *N
	}

	configs := []config{
		func() config {
			data := map[id.ID]hyperrectangle.R{
				100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
				101: *hyperrectangle.New(vector.V{99, 99}, vector.V{100, 100}),
			}
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{100: true},
			}

			return config{
				name: "Simple",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{100: true, 101: true},
				},
				x:    101,
				data: data,
				want: want,
			}
		}(),
		func() config {
			data := map[id.ID]hyperrectangle.R{
				100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
			}
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{},
			}

			return config{
				name: "Root/NoCollapse",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{100: true},
				},
				x:    100,
				data: data,
				want: want,
			}
		}(),
		func() config {
			data := map[id.ID]hyperrectangle.R{
				100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
			}
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{},
			}

			n := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{},
			}
			n.children = [4]*N{
				&N{
					depth:  1,
					corner: ChildNE,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
				},
				&N{
					depth:  1,
					corner: ChildSE,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
				},
				&N{
					depth:  1,
					corner: ChildSW,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					lookup: map[id.ID]bool{100: true},
				},
				&N{
					depth:  1,
					corner: ChildNW,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
				},
			}

			return config{
				name: "Child/Collapse",
				n:    n,
				x:    100,
				data: data,
				want: want,
			}
		}(),
		func() config {
			data := map[id.ID]hyperrectangle.R{
				100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
				101: *hyperrectangle.New(vector.V{99, 99}, vector.V{100, 100}),
			}
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{},
			}
			want.children = [4]*N{
				&N{
					depth:  1,
					corner: ChildNE,
					parent: want,
					aabb:   *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
					lookup: map[id.ID]bool{101: true},
				},
				&N{
					depth:  1,
					corner: ChildSE,
					parent: want,
					aabb:   *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
					lookup: map[id.ID]bool{},
				},
				&N{
					depth:  1,
					corner: ChildSW,
					parent: want,
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					lookup: map[id.ID]bool{},
				},
				&N{
					depth:  1,
					corner: ChildNW,
					parent: want,
					aabb:   *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
					lookup: map[id.ID]bool{},
				},
			}

			n := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{},
			}
			n.children = [4]*N{
				&N{
					depth:  1,
					corner: ChildNE,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
					lookup: map[id.ID]bool{101: true},
				},
				&N{
					depth:  1,
					corner: ChildSE,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
					lookup: map[id.ID]bool{},
				},
				&N{
					depth:  1,
					corner: ChildSW,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					lookup: map[id.ID]bool{100: true},
				},
				&N{
					depth:  1,
					corner: ChildNW,
					parent: n,
					aabb:   *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
					lookup: map[id.ID]bool{},
				},
			}

			return config{
				name: "Child/NoCollapse",
				n:    n,
				x:    100,
				data: data,
				want: want,
			}
		}(),
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			c.n.Remove(c.x, c.data)
			if diff := cmp.Diff(c.want, c.n, opts...); diff != "" {
				t.Errorf("Remove() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	type config struct {
		name string
		n    *N
		x    id.ID
		data map[id.ID]hyperrectangle.R
		want *N
	}

	configs := []config{
		func() config {
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{100: true},
			}

			return config{
				name: "Simple",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{},
				},
				x: 100,
				data: map[id.ID]hyperrectangle.R{
					100: *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				},
				want: want,
			}
		}(),
		func() config {
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup: map[id.ID]bool{100: true},
				floor:  0,
			}

			return config{
				name: "Simple/AtFloor",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{},
				},
				x: 100,
				data: map[id.ID]hyperrectangle.R{
					100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
				},
				want: want,
			}
		}(),
		func() config {
			want := &N{
				aabb:      *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				lookup:    map[id.ID]bool{100: true},
				floor:     1,
				tolerance: 100,
			}

			return config{
				name: "Simple/TightFit",
				n: &N{
					aabb:      *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup:    map[id.ID]bool{},
					floor:     1,
					tolerance: 100,
				},
				x: 100,
				data: map[id.ID]hyperrectangle.R{
					100: *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 99.1}),
				},
				want: want,
			}
		}(),
		func() config {
			want := &N{
				aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
				floor:  1,
				lookup: map[id.ID]bool{},
			}
			want.children = [4]*N{
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildNE,
					aabb:      *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
					floor:     1,
					lookup:    map[id.ID]bool{},
					cachePath: []Child{ChildNE},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildSE,
					aabb:      *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
					floor:     1,
					lookup:    map[id.ID]bool{},
					cachePath: []Child{ChildSE},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildSW,
					aabb:      *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					floor:     1,
					lookup:    map[id.ID]bool{100: true},
					cachePath: []Child{ChildSW},
				},
				&N{
					depth:     1,
					parent:    want,
					corner:    ChildNW,
					aabb:      *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
					floor:     1,
					lookup:    map[id.ID]bool{},
					cachePath: []Child{ChildNW},
				},
			}

			return config{
				name: "Simple/InsertLayer",
				n: &N{
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{100, 100}),
					lookup: map[id.ID]bool{},
					floor:  1,
				},
				x: 100,
				data: map[id.ID]hyperrectangle.R{
					100: *hyperrectangle.New(vector.V{0, 0}, vector.V{1, 1}),
				},
				want: want,
			}
		}(),
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			c.n.Insert(c.x, c.data)
			if diff := cmp.Diff(c.want, c.n, opts...); diff != "" {
				t.Errorf("Insert() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestFSM(t *testing.T) {
	type config struct {
		name string
		path []Child
		e    Edge
		want []Child
	}

	configs := []config{
		{
			name: "Trivial",
			path: []Child{},
			e:    EdgeN,
			want: nil,
		},
		{
			name: "Edge",
			path: []Child{ChildNE},
			e:    EdgeN,
			want: nil,
		},
		{
			name: "Edge/Recursive",
			path: []Child{ChildSE, ChildSW},
			e:    EdgeS,
			want: nil,
		},
		{
			name: "Sibling",
			path: []Child{ChildSE},
			e:    EdgeN,
			want: []Child{ChildNE},
		},
		{
			name: "Sibling/Recursive",
			path: []Child{ChildSE, ChildSW},
			e:    EdgeW,
			want: []Child{ChildSW, ChildSE},
		},
		{
			name: "Sibling/Recursive/Sibling",
			path: []Child{ChildSE, ChildSE},
			e:    EdgeN,
			want: []Child{ChildSE, ChildNE},
		},
		{
			name: "Sibling/Composite",
			path: FSM([]Child{ChildNE, ChildSW}, EdgeS),
			e:    EdgeW,
			want: []Child{ChildSW, ChildNE},
		},
	}

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			got := FSM(c.path, c.e)
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("FSM() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}

func TestNeighbors(t *testing.T) {
	type config struct {
		name string
		n    *N
		want []*N
	}

	configs := []config{
		{
			name: "Root/NoNeighbors",
			n:    &N{},
			want: []*N{},
		},
		func() config {
			root := &N{}
			root.children = [4]*N{
				&N{corner: ChildNE, parent: root, depth: 1, cachePath: []Child{ChildNE}},
				&N{corner: ChildSE, parent: root, depth: 1, cachePath: []Child{ChildSE}},
				&N{corner: ChildSW, parent: root, depth: 1, cachePath: []Child{ChildSW}},
				&N{corner: ChildNW, parent: root, depth: 1, cachePath: []Child{ChildNW}},
			}

			return config{
				name: "Simple/FilterEdge",
				n:    root.children[ChildNE],
				want: []*N{
					root.children[ChildSE],
					root.children[ChildNW],
					root.children[ChildSW],
				},
			}
		}(),
	}
	configs = append(configs, func() []config {
		root := &N{}
		root.children = [4]*N{
			&N{corner: ChildNE, parent: root, depth: 1, cachePath: []Child{ChildNE}},
			&N{corner: ChildSE, parent: root, depth: 1, cachePath: []Child{ChildSE}},
			&N{corner: ChildSW, parent: root, depth: 1, cachePath: []Child{ChildSW}},
			&N{corner: ChildNW, parent: root, depth: 1, cachePath: []Child{ChildNW}},
		}
		root.children[ChildNE].children = [4]*N{
			&N{corner: ChildNE, parent: root, depth: 2, cachePath: []Child{ChildNE, ChildNE}},
			&N{corner: ChildSE, parent: root, depth: 2, cachePath: []Child{ChildNE, ChildSE}},
			&N{corner: ChildSW, parent: root, depth: 2, cachePath: []Child{ChildNE, ChildSW}},
			&N{corner: ChildNW, parent: root, depth: 2, cachePath: []Child{ChildNE, ChildNW}},
		}

		return []config{
			{
				name: "Large/SmallNeighbors",
				n:    root.children[ChildSE],
				want: []*N{
					root.children[ChildNE].children[ChildSE],
					root.children[ChildNE].children[ChildSW],
					root.children[ChildSW],
					root.children[ChildNW],
				},
			},
			{
				name: "Small/LargeNeighbors",
				n:    root.children[ChildNE].children[ChildSE],
				want: []*N{
					root.children[ChildNE].children[ChildNE],
					root.children[ChildSE],
					root.children[ChildNE].children[ChildSW],
					root.children[ChildNE].children[ChildNW],
				},
			},
			{
				name: "Large/SmallCorner",
				n:    root.children[ChildSW],
				want: []*N{
					root.children[ChildNW],
					root.children[ChildSE],
					root.children[ChildNE].children[ChildSW],
				},
			},
		}
	}()...)

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			got := c.n.Neighbors()
			if diff := cmp.Diff(c.want, got, opts...); diff != "" {
				t.Errorf("Neighbors() mismatch (-want +got):\n%v", diff)
			}
		})
	}
}
