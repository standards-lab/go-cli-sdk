package scenario_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/standards-lab/go-cli-sdk"
	"github.com/standards-lab/go-cli-sdk/scenario"
	"github.com/standards-lab/go-core/config"
	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-core/process"
)

// The scenario parent driven through cli.Run over buffers, mounted under a
// test root over a small graph: db, and cache, which uses db. Each node's
// value is a subsystem that records its Start and Shutdown, and each
// constructor records its build, so a test reads what a run built and
// started from the recorder.

// recorder collects the events of one dispatch, safe for the concurrent
// starts and shutdowns of a layer.
type recorder struct {
	mu     sync.Mutex
	events []string
}

func (r *recorder) add(event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
}

func (r *recorder) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.events)
}

// matching returns the recorded events that start with prefix, in order.
func (r *recorder) matching(prefix string) []string {
	var out []string
	for _, e := range r.list() {
		if strings.HasPrefix(e, prefix) {
			out = append(out, e)
		}
	}
	return out
}

// subsystem is a node's value: it records its Start and Shutdown, and its
// Start fails with startErr when that is set.
type subsystem struct {
	name     string
	rec      *recorder
	startErr error
}

func (s *subsystem) Start(context.Context) error {
	s.rec.add("start " + s.name)
	return s.startErr
}

func (s *subsystem) Shutdown(context.Context) error {
	s.rec.add("stop " + s.name)
	return nil
}

// harness is the test graph and what a run records over it.
type harness struct {
	g     *graph.Graph
	cfg   *graph.Node[lifecycle.Config]
	db    *graph.Node[*subsystem]
	cache *graph.Node[*subsystem]
	rec   *recorder

	cacheStartErr error // set before a dispatch to fail cache's Start
}

func newHarness() *harness {
	h := &harness{g: graph.New(), rec: &recorder{}}
	h.cfg = h.g.Define("lifecycle", func(*graph.Scope) (lifecycle.Config, error) {
		h.rec.add("build lifecycle")
		return lifecycle.Config{ShutdownTimeout: config.Duration(5 * time.Second)}, nil
	})
	h.db = h.g.Define("db", func(*graph.Scope) (*subsystem, error) {
		h.rec.add("build db")
		return &subsystem{name: "db", rec: h.rec}, nil
	})
	h.cache = h.g.Define("cache", func(s *graph.Scope) (*subsystem, error) {
		s.Use(h.db)
		h.rec.add("build cache")
		return &subsystem{name: "cache", rec: h.rec, startErr: h.cacheStartErr}, nil
	})
	return h
}

// root mounts scenario.Commands(scenarios...) under a test root.
func (h *harness) root(scenarios ...scenario.Scenario) *cli.Command {
	return (&cli.Command{Name: "prog"}).Add(scenario.Commands(scenarios...)...)
}

type result struct {
	code           int
	stdout, stderr string
}

// run dispatches args over root with ctx, fresh buffers, and the harness's
// graph.
func (h *harness) run(ctx context.Context, root *cli.Command, args ...string) result {
	var stdout, stderr bytes.Buffer
	streams := cli.Streams{Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr}
	code := cli.Run(ctx, root, args, streams, cli.WithGraph(h.g, h.cfg))
	return result{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

// builds returns the nodes the run built, in the order it built them.
func (h *harness) builds() []string {
	return h.rec.matching("build ")
}

// step returns a step that records its intent and succeeds.
func (h *harness) step(intent string) scenario.Step {
	return scenario.Step{Intent: intent, Action: func(context.Context, *cli.Invocation, *scenario.Reporter) error {
		h.rec.add("step " + intent)
		return nil
	}}
}

// pair is two scenarios: first declares db alone, second cache and db.
func (h *harness) pair() []scenario.Scenario {
	return []scenario.Scenario{
		{Name: "first", Summary: "Show the first tour", Nodes: []graph.Ref{h.db}, Steps: []scenario.Step{h.step("one")}},
		{Name: "second-tour", Summary: "Show the second tour", Nodes: []graph.Ref{h.cache, h.db}, Steps: []scenario.Step{h.step("two")}},
	}
}

// pairListing is the listing of pair's scenarios.
const pairListing = "" +
	"Scenarios:\n" +
	"  first         Show the first tour\n" +
	"                uses db\n" +
	"  second-tour   Show the second tour\n" +
	"                uses cache, db\n"

func TestCommands_ParentHelpListsTheLeavesInOrderAndEndsWithTheListing(t *testing.T) {
	for _, args := range [][]string{{"scenario"}, {"scenario", "-h"}, {"scenario", "--help"}} {
		h := newHarness()
		root := h.root(h.pair()...)

		r := h.run(context.Background(), root, args...)

		if r.code != process.ExitUsage {
			t.Errorf("%v: code = %d, want %d", args, r.code, process.ExitUsage)
		}
		// The Commands section and the listing pad the name column to one
		// width, so each scenario's line starts the same in both.
		commands := "Commands:\n" +
			"  first         Show the first tour\n" +
			"  second-tour   Show the second tour\n"
		if !strings.Contains(r.stdout, "Usage:\n  prog scenario <command> [flags]\n") || !strings.Contains(r.stdout, commands) {
			t.Errorf("%v: help:\n%s\nwant the usage line and:\n%s", args, r.stdout, commands)
		}
		if !strings.HasSuffix(r.stdout, "\n\n"+pairListing) {
			t.Errorf("%v: help:\n%s\nwant it to end with:\n%s", args, r.stdout, pairListing)
		}
		if r.stderr != "" {
			t.Errorf("%v: stderr = %q, want empty", args, r.stderr)
		}
		if got := h.rec.list(); len(got) != 0 {
			t.Errorf("%v: events = %q, want nothing built", args, got)
		}
	}
}

func TestCommands_ALeafsHelpIsItsOwn(t *testing.T) {
	h := newHarness()
	root := h.root(h.pair()...)

	r := h.run(context.Background(), root, "scenario", "second-tour", "--help")

	if r.code != process.ExitUsage {
		t.Errorf("code = %d, want %d", r.code, process.ExitUsage)
	}
	if want := "Show the second tour\n\nUsage:\n  prog scenario second-tour [flags]\n"; !strings.HasPrefix(r.stdout, want) {
		t.Errorf("help:\n%s\nwant it to start:\n%s", r.stdout, want)
	}
	if strings.Contains(r.stdout, "Scenarios:") {
		t.Errorf("help:\n%s\nwant no listing: the footer is the parent's", r.stdout)
	}
	if got := h.rec.list(); len(got) != 0 {
		t.Errorf("events = %q, want nothing built", got)
	}
}

func TestCommands_AUsageErrorOnALeafBuildsNothing(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stderr string // the start of the report
	}{
		{"an argument", []string{"scenario", "first", "extra"}, "prog scenario first: "},
		{"an unknown flag", []string{"scenario", "second-tour", "--bogus"}, "prog scenario second-tour: "},
		{"an unknown scenario", []string{"scenario", "bogus"}, "prog scenario: unknown command \"bogus\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness()
			root := h.root(h.pair()...)

			r := h.run(context.Background(), root, tt.args...)

			if r.code != process.ExitUsage {
				t.Errorf("code = %d, want %d; stderr = %q", r.code, process.ExitUsage, r.stderr)
			}
			if !strings.HasPrefix(r.stderr, tt.stderr) {
				t.Errorf("stderr = %q, want it to start %q", r.stderr, tt.stderr)
			}
			if got := h.rec.list(); len(got) != 0 {
				t.Errorf("events = %q, want nothing built or run", got)
			}
		})
	}
}

func TestCommands_ALeafBuildsOnlyItsScenariosNodes(t *testing.T) {
	h := newHarness()
	root := h.root(h.pair()...)

	r := h.run(context.Background(), root, "scenario", "first")

	if r.code != process.ExitOK {
		t.Fatalf("code = %d, want %d; stderr = %q", r.code, process.ExitOK, r.stderr)
	}
	// first declares db alone, and the parent declares nothing, so cache,
	// which second-tour declares, is never built.
	if got, want := h.builds(), []string{"build db", "build lifecycle"}; !slices.Equal(got, want) {
		t.Errorf("builds = %q, want %q", got, want)
	}
	if got := h.rec.matching("step "); !slices.Equal(got, []string{"step one"}) {
		t.Errorf("steps = %q, want first's one step", got)
	}
}

func TestCommands_BuildsEachDeclaredNodeOnceUnderTheCoordinator(t *testing.T) {
	h := newHarness()
	// db is declared twice and reached again through cache; each step
	// reads db, and both must read the one value the Build made.
	var seen []*subsystem
	read := func(intent string) scenario.Step {
		return scenario.Step{Intent: intent, Action: func(_ context.Context, inv *cli.Invocation, _ *scenario.Reporter) error {
			seen = append(seen, inv.Get(h.db))
			h.rec.add("step " + intent)
			return nil
		}}
	}
	root := h.root(scenario.Scenario{
		Name:  "tour",
		Nodes: []graph.Ref{h.db, h.cache, h.db},
		Steps: []scenario.Step{read("one"), read("two")},
	})

	r := h.run(context.Background(), root, "scenario", "tour")

	if r.code != process.ExitOK {
		t.Fatalf("code = %d, want %d; stderr = %q", r.code, process.ExitOK, r.stderr)
	}
	if got, want := h.builds(), []string{"build db", "build cache", "build lifecycle"}; !slices.Equal(got, want) {
		t.Errorf("builds = %q, want %q", got, want)
	}
	if len(seen) != 2 || seen[0] != seen[1] {
		t.Errorf("the steps read db as %v, want one value twice", seen)
	}
	// The Coordinator starts db's layer before cache's, runs the steps,
	// and shuts down in reverse.
	want := []string{"start db", "start cache", "step one", "step two", "stop cache", "stop db"}
	var got []string
	for _, e := range h.rec.list() {
		if !strings.HasPrefix(e, "build ") {
			got = append(got, e)
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("events = %q, want %q", got, want)
	}
}

func TestCommands_ADeclaredNodeThatFailsToStartFailsTheRunBeforeTheFirstHeading(t *testing.T) {
	h := newHarness()
	h.cacheStartErr = errors.New("down")
	root := h.root(h.pair()...)

	r := h.run(context.Background(), root, "scenario", "second-tour")

	if r.code != process.ExitFailure {
		t.Errorf("code = %d, want %d", r.code, process.ExitFailure)
	}
	if want := "prog scenario second-tour: startup: cache: down\n"; r.stderr != want {
		t.Errorf("stderr = %q, want %q", r.stderr, want)
	}
	if r.stdout != "" {
		t.Errorf("stdout = %q, want no heading", r.stdout)
	}
	if got := h.rec.matching("step "); len(got) != 0 {
		t.Errorf("steps = %q, want none run", got)
	}
	if got := h.rec.matching("stop db"); len(got) != 1 {
		t.Errorf("events = %q, want db, which started, shut down once", h.rec.list())
	}
}

func TestWriteListing_WritesTheScenariosAndTheNodesEachDeclares(t *testing.T) {
	h := newHarness()
	var out bytes.Buffer

	scenario.WriteListing(&out, h.pair()...)

	if out.String() != pairListing {
		t.Errorf("WriteListing wrote:\n%s\nwant:\n%s", out.String(), pairListing)
	}
	if got := h.rec.list(); len(got) != 0 {
		t.Errorf("events = %q, want nothing built", got)
	}
}

func TestWriteListing_LeavesOutTheUsesLineOfAScenarioThatDeclaresNoNodes(t *testing.T) {
	h := newHarness()
	var out bytes.Buffer

	scenario.WriteListing(&out, scenario.Scenario{Name: "bare", Summary: "Declare nothing", Steps: []scenario.Step{h.step("one")}})

	if want := "Scenarios:\n  bare   Declare nothing\n"; out.String() != want {
		t.Errorf("WriteListing wrote %q, want %q", out.String(), want)
	}
}

func TestWriteListing_ARootFooterEndsTheRootHelpWithTheListing(t *testing.T) {
	h := newHarness()
	scenarios := h.pair()
	root := h.root(scenarios...)
	root.Footer = func(w io.Writer) { scenario.WriteListing(w, scenarios...) }

	r := h.run(context.Background(), root, "--help")

	if r.code != process.ExitUsage {
		t.Errorf("code = %d, want %d", r.code, process.ExitUsage)
	}
	if !strings.HasSuffix(r.stdout, "\n\n"+pairListing) {
		t.Errorf("help:\n%s\nwant it to end with:\n%s", r.stdout, pairListing)
	}
	if got := h.rec.list(); len(got) != 0 {
		t.Errorf("events = %q, want nothing built", got)
	}
}

func TestRun_NarratesEachStepUnderItsHeading(t *testing.T) {
	h := newHarness()
	errRender := errors.New("render failed")
	var showErr error
	root := h.root(scenario.Scenario{
		Name:  "tour",
		Nodes: []graph.Ref{h.db},
		Steps: []scenario.Step{
			{Intent: "Open the database", Action: func(_ context.Context, _ *cli.Invocation, r *scenario.Reporter) error {
				r.Note("The database node was built and started before this step ran, so the step reads its value with %s and works over a System that is already up.", "inv.Get")
				return nil
			}},
			{Intent: "Read a value", Action: func(_ context.Context, _ *cli.Invocation, r *scenario.Reporter) error {
				r.Note("Read it.")
				return r.Show(func(w io.Writer) error {
					_, err := fmt.Fprint(w, "line one\nline two\n")
					return err
				})
			}},
			{Intent: "Fail to render", Action: func(_ context.Context, _ *cli.Invocation, r *scenario.Reporter) error {
				showErr = r.Show(func(w io.Writer) error {
					_, _ = fmt.Fprint(w, "partial")
					return errRender
				})
				r.Note("After.")
				return nil
			}},
		},
	})

	r := h.run(context.Background(), root, "scenario", "tour")

	if r.code != process.ExitOK {
		t.Fatalf("code = %d, want %d; stderr = %q", r.code, process.ExitOK, r.stderr)
	}
	// The note wraps at 80 columns, its indent included: the first line
	// is 75 wide, and the next word would take it to 81.
	want := "" +
		"[1/3] Open the database\n" +
		"  The database node was built and started before this step ran, so the step\n" +
		"  reads its value with inv.Get and works over a System that is already up.\n" +
		"\n" +
		"[2/3] Read a value\n" +
		"  Read it.\n" +
		"    line one\n" +
		"    line two\n" +
		"\n" +
		"[3/3] Fail to render\n" +
		"  After.\n"
	if r.stdout != want {
		t.Errorf("stdout:\n%s\nwant:\n%s", r.stdout, want)
	}
	if !errors.Is(showErr, errRender) {
		t.Errorf("Show returned %v, want the render's error", showErr)
	}
}

func TestRun_NoteKeepsParagraphsAndWrapsByRune(t *testing.T) {
	h := newHarness()
	// 26 two-rune words take 77 columns, under the 78 a note's line holds
	// past its indent, though their bytes alone would take 129.
	wide := strings.Repeat("éé ", 25) + "éé"
	root := h.root(scenario.Scenario{
		Name: "tour",
		Steps: []scenario.Step{
			{Intent: "Speak", Action: func(_ context.Context, _ *cli.Invocation, r *scenario.Reporter) error {
				r.Note("First paragraph.\n\nSecond paragraph.")
				r.Note("%s", wide)
				return nil
			}},
		},
	})

	r := h.run(context.Background(), root, "scenario", "tour")

	if r.code != process.ExitOK {
		t.Fatalf("code = %d, want %d; stderr = %q", r.code, process.ExitOK, r.stderr)
	}
	want := "" +
		"[1/1] Speak\n" +
		"  First paragraph.\n" +
		"\n" +
		"  Second paragraph.\n" +
		"  " + wide + "\n"
	if r.stdout != want {
		t.Errorf("stdout:\n%s\nwant:\n%s", r.stdout, want)
	}
}

func TestRun_AFailedStepStopsTheRunAndIsReportedOnce(t *testing.T) {
	h := newHarness()
	root := h.root(scenario.Scenario{
		Name:  "tour",
		Nodes: []graph.Ref{h.db},
		Steps: []scenario.Step{
			h.step("one"),
			{Intent: "Break", Action: func(context.Context, *cli.Invocation, *scenario.Reporter) error {
				return errors.New("boom")
			}},
			h.step("three"),
		},
	})

	r := h.run(context.Background(), root, "scenario", "tour")

	if r.code != process.ExitFailure {
		t.Errorf("code = %d, want %d", r.code, process.ExitFailure)
	}
	if want := "prog scenario tour: step 2 (Break): boom\n"; r.stderr != want {
		t.Errorf("stderr = %q, want %q", r.stderr, want)
	}
	if want := "[1/3] one\n\n[2/3] Break\n"; r.stdout != want {
		t.Errorf("stdout = %q, want %q", r.stdout, want)
	}
	if got := h.rec.matching("step "); !slices.Equal(got, []string{"step one"}) {
		t.Errorf("steps = %q, want step three not run", got)
	}
	if got := h.rec.matching("stop db"); len(got) != 1 {
		t.Errorf("events = %q, want db shut down once", h.rec.list())
	}
}

func TestRun_AContextThatEndsFailsTheStepUnderWay(t *testing.T) {
	h := newHarness()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	root := h.root(scenario.Scenario{
		Name:  "tour",
		Nodes: []graph.Ref{h.db},
		Steps: []scenario.Step{
			{Intent: "Wait", Action: func(ctx context.Context, _ *cli.Invocation, _ *scenario.Reporter) error {
				cancel()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(5 * time.Second):
					return errors.New("the step's context did not end with the run's")
				}
			}},
			h.step("after"),
		},
	})

	r := h.run(ctx, root, "scenario", "tour")

	if r.code != process.ExitFailure {
		t.Errorf("code = %d, want %d", r.code, process.ExitFailure)
	}
	if want := "prog scenario tour: step 1 (Wait): " + context.Canceled.Error() + "\n"; r.stderr != want {
		t.Errorf("stderr = %q, want %q", r.stderr, want)
	}
	if got := h.rec.matching("step "); len(got) != 0 {
		t.Errorf("steps = %q, want the step after not run", got)
	}
}

func TestCommands_PanicsOnAWiringMistake(t *testing.T) {
	h := newHarness()
	steps := []scenario.Step{h.step("one")}
	tests := []struct {
		name      string
		scenarios []scenario.Scenario
		want      string // what the panic message names
	}{
		{"no scenarios", nil, "no scenarios"},
		{"two with one name", []scenario.Scenario{{Name: "twin", Steps: steps}, {Name: "twin", Steps: steps}}, `"twin"`},
		{"one with no steps", []scenario.Scenario{{Name: "empty"}}, `"empty"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				msg, ok := recover().(string)
				if !ok || !strings.HasPrefix(msg, "scenario: ") || !strings.Contains(msg, tt.want) {
					t.Errorf("panic = %q, want a scenario: message naming %s", msg, tt.want)
				}
			}()
			scenario.Commands(tt.scenarios...)
		})
	}
}
