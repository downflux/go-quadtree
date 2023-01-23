package node

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Diff(n *N, m *N) string {
	return cmp.Diff(n, m)
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
