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
			got := c.n.Get(c.path)
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
			got := c.n.Path()
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
			}
			want.children = [4]*N{
				&N{
					depth:  1,
					parent: want,
					corner: ChildNE,
					aabb:   *hyperrectangle.New(vector.V{50, 50}, vector.V{100, 100}),
					lookup: map[id.ID]bool{},
				},
				&N{
					depth:  1,
					parent: want,
					corner: ChildSE,
					aabb:   *hyperrectangle.New(vector.V{50, 0}, vector.V{100, 50}),
					lookup: map[id.ID]bool{},
				},
				&N{
					depth:  1,
					parent: want,
					corner: ChildSW,
					aabb:   *hyperrectangle.New(vector.V{0, 0}, vector.V{50, 50}),
					lookup: map[id.ID]bool{100: true},
				},
				&N{
					depth:  1,
					parent: want,
					corner: ChildNW,
					aabb:   *hyperrectangle.New(vector.V{0, 50}, vector.V{50, 100}),
					lookup: map[id.ID]bool{100: true},
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
