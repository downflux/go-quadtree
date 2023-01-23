package node

import (
	"testing"

	"github.com/downflux/go-geometry/nd/hyperrectangle"
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
