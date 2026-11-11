package cli_test

import (
	"bytes"
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/standards-lab/go-cli-sdk"
	"github.com/standards-lab/go-core/config"
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-core/process"
)

// leafTree is a test tree with Use at several levels:
//
//	prog            Use a
//	  data          Use b, a   (a again: inherited, counted once)
//	    get         Use c, b   (b again)
//	    put
//	  tools
//	    deep        Use d
//	      x         Use a      (a again, from two levels up)
//	  plain
//
// Every node is a plain value with no dependencies, so a Build reaches
// exactly the roots it is given, in the order given.
type leafTree struct {
	root       *cli.Command
	get        *cli.Command
	g          *graph.Graph
	cfg        *graph.Node[lifecycle.Config]
	a, b, c, d *graph.Node[int]
}

func newLeafTree() *leafTree {
	t := &leafTree{g: graph.New()}
	t.cfg = t.g.Value("lifecycle", lifecycle.Config{ShutdownTimeout: config.Duration(5 * time.Second)})
	node := func(name string) *graph.Node[int] { return t.g.Value(name, 0) }
	t.a, t.b, t.c, t.d = node("a"), node("b"), node("c"), node("d")
	run := func(context.Context, *cli.Invocation) error { return nil }
	t.get = (&cli.Command{Name: "get", Run: run}).Use(t.c, t.b)
	t.root = (&cli.Command{Name: "prog"}).Use(t.a).Add(
		(&cli.Command{Name: "data"}).Use(t.b, t.a).Add(
			t.get,
			&cli.Command{Name: "put", Run: run},
		),
		(&cli.Command{Name: "tools"}).Add(
			(&cli.Command{Name: "deep"}).Use(t.d).Add(
				(&cli.Command{Name: "x", Run: run}).Use(t.a),
			),
		),
		&cli.Command{Name: "plain", Run: run},
	)
	return t
}

// leaf is one pair Leaves yielded, with the nodes as names.
type leaf struct {
	path  string
	nodes []string
}

// names returns the names of refs, in order.
func names(refs []graph.Ref) []string {
	var out []string
	for _, r := range refs {
		out = append(out, r.Name())
	}
	return out
}

// collect ranges over cmd.Leaves() to the end.
func collect(cmd *cli.Command) []leaf {
	var out []leaf
	for path, refs := range cmd.Leaves() {
		out = append(out, leaf{path, names(refs)})
	}
	return out
}

func TestLeaves_YieldsEachLeafDepthFirstInAddOrderWithItsPathsUnion(t *testing.T) {
	tr := newLeafTree()
	want := []leaf{
		{"prog data get", []string{"a", "b", "c"}},
		{"prog data put", []string{"a", "b"}},
		{"prog tools deep x", []string{"a", "d"}},
		{"prog plain", []string{"a"}},
	}
	if got := collect(tr.root); !slices.EqualFunc(got, want, equalLeaf) {
		t.Errorf("Leaves = %v, want %v", got, want)
	}
}

func equalLeaf(x, y leaf) bool {
	return x.path == y.path && slices.Equal(x.nodes, y.nodes)
}

func TestLeaves_EachSetIsWhatADispatchToTheLeafBuilds(t *testing.T) {
	tr := newLeafTree()
	var mu sync.Mutex
	var built []graph.Ref
	tr.g.Observe(func(r graph.Ref) {
		mu.Lock()
		defer mu.Unlock()
		built = append(built, r)
	})
	for path, refs := range tr.root.Leaves() {
		mu.Lock()
		built = nil
		mu.Unlock()
		var stdout, stderr bytes.Buffer
		streams := cli.Streams{Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr}
		args := strings.Fields(path)[1:]
		if code := (cli.Program{Root: tr.root, Graph: tr.g, LifecycleConfig: tr.cfg}).Run(context.Background(), args, streams); code != process.ExitOK {
			t.Fatalf("Run(%q) = %d, stderr %q", args, code, stderr.String())
		}
		// The nodes have no dependencies, so the Build reaches its roots in
		// the order given: the leaf's set, then the lifecycle node.
		mu.Lock()
		want := append(slices.Clone(refs), graph.Ref(tr.cfg))
		if !slices.Equal(built, want) {
			t.Errorf("%s: dispatch built %v, Leaves yielded %v plus lifecycle", path, names(built), names(refs))
		}
		mu.Unlock()
	}
}

func TestLeaves_StopsWhenTheConsumerStops(t *testing.T) {
	tr := newLeafTree()
	var paths []string
	for path := range tr.root.Leaves() {
		paths = append(paths, path)
		if len(paths) == 2 {
			break
		}
	}
	if want := []string{"prog data get", "prog data put"}; !slices.Equal(paths, want) {
		t.Errorf("paths = %q, want %q", paths, want)
	}
}

func TestLeaves_AYieldedSliceIsTheCallersOwn(t *testing.T) {
	tr := newLeafTree()
	for _, refs := range tr.root.Leaves() {
		for i := range refs {
			refs[i] = tr.d
		}
		_ = append(refs[:0], tr.c, tr.c, tr.c, tr.c)
	}
	want := []leaf{
		{"prog data get", []string{"a", "b", "c"}},
		{"prog data put", []string{"a", "b"}},
		{"prog tools deep x", []string{"a", "d"}},
		{"prog plain", []string{"a"}},
	}
	if got := collect(tr.root); !slices.EqualFunc(got, want, equalLeaf) {
		t.Errorf("after mutating the yields, Leaves = %v, want %v", got, want)
	}
}

func TestLeaves_OnALeafYieldsItselfWithItsWholePathsUnion(t *testing.T) {
	tr := newLeafTree()
	want := []leaf{{"prog data get", []string{"a", "b", "c"}}}
	if got := collect(tr.get); !slices.EqualFunc(got, want, equalLeaf) {
		t.Errorf("get.Leaves = %v, want %v", got, want)
	}

	solo := &cli.Command{Name: "solo", Run: func(context.Context, *cli.Invocation) error { return nil }}
	want = []leaf{{"solo", nil}}
	if got := collect(solo); !slices.EqualFunc(got, want, equalLeaf) {
		t.Errorf("solo.Leaves = %v, want %v", got, want)
	}
}

func TestLeaves_OnASubtreeYieldsItsLeavesWithTheirPathsFromTheRoot(t *testing.T) {
	tr := newLeafTree()
	run := func(context.Context, *cli.Invocation) error { return nil }
	tools := (&cli.Command{Name: "tools"}).Use(tr.d).Add(
		(&cli.Command{Name: "x", Run: run}).Use(tr.a, tr.d),
	)
	(&cli.Command{Name: "prog"}).Use(tr.a).Add(tools)
	want := []leaf{{"prog tools x", []string{"a", "d"}}}
	if got := collect(tools); !slices.EqualFunc(got, want, equalLeaf) {
		t.Errorf("tools.Leaves = %v, want %v", got, want)
	}
}

func TestLeaves_AParentWithNoLeavesYieldsNothing(t *testing.T) {
	empty := (&cli.Command{Name: "prog"}).Add(&cli.Command{Name: "idle"})
	if got := collect(empty); len(got) != 0 {
		t.Errorf("Leaves = %v, want none", got)
	}
}
