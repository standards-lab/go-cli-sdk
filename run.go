package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/standards-lab/go-core/graph"
	"github.com/standards-lab/go-core/lifecycle"
	"github.com/standards-lab/go-core/process"
)

// Program is a command tree ready to dispatch: its root, and the graph that
// the nodes its commands declare with [Command.Use] are built from.
//
// Root is required. Graph and LifecycleConfig go together: a tree that
// declares Use sets both, and one that declares none may leave both nil.
type Program struct {
	// Root is the root of the command tree.
	Root *Command

	// Graph is the graph that the nodes a leaf's path declares with
	// [Command.Use] are built from.
	Graph *graph.Graph

	// LifecycleConfig is the node whose value configures the
	// [lifecycle.Coordinator] that runs what was built. It is added to
	// every Build, so its constructor supplies the shutdown timeout, and
	// it must return a finalized Config: lifecycle.New panics on one that
	// is not.
	LifecycleConfig *graph.Node[lifecycle.Config]
}

// Run dispatches args, the program arguments without the program name, over
// the tree rooted at p.Root, and returns the process exit code. streams are
// the process's streams, or a test's buffers: a running command reads Stdin
// and writes Stdout and Stderr through its [Invocation], which embeds them,
// and the dispatcher prints to Stdout and Stderr itself but never reads
// Stdin.
//
// A dispatch to a leaf goes in this order, stopping at the first failure:
//
//  1. parse each level's flags and select the leaf
//  2. count its positional arguments with [Command.Args]
//  3. check its required flags, [Command.Require]
//  4. check its exclusive groups, [Command.Exclusive]
//  5. validate its input with [Command.Validate]
//  6. run the root's [Command.PreRun]
//  7. build the nodes its path declares with [Command.Use]
//  8. run the leaf with [Command.Run]
//
// When the leaf's path has nodes declared with Use, the Build constructs
// their union and p.LifecycleConfig from p.Graph; the System is started
// with a [lifecycle.Coordinator], the leaf runs under it, reading each
// declared node's value with [Invocation.Get], and the System is shut
// down. Nothing is built when the dispatch ends before the Build.
//
// Run owns every line the dispatcher prints, and returns go-core's process
// exit codes:
//
//   - a parent command run with no subcommand, or any command given -h or
//     --help, prints its generated help to stdout and returns ExitUsage
//   - an unknown subcommand, an unknown flag, a malformed flag value, a
//     rejected argument count, a missing required flag, a broken exclusive
//     group, any error Validate returns, or a [UsageError] that PreRun or
//     the command returns is reported on stderr with the command's usage
//     and returns ExitUsage
//   - any other error PreRun or the command returns, a Build error, or a
//     failure the [lifecycle.Coordinator] reports, to start, of a
//     [lifecycle.Monitored] value while the leaf runs, which also ends the
//     leaf's context, or to shut down, is reported once on stderr and
//     returns ExitFailure; a leaf that returns only its context's end, or
//     the cause [context.Cause] gives for it, after a monitored failure
//     ended it reports one line, the failure labelled "run:" and its
//     node's name, and a leaf error unrelated to the failure is reported
//     beside it
//   - a command that succeeds returns ExitOK
//
// A tree may be dispatched more than once. Each dispatch starts from the
// flag values the tree held before the previous one: Run restores every
// flag the previous dispatch set on the command line to the value it held
// just before that dispatch parsed it, and touches no other flag, so a
// value the program gave a flag variable after defining it survives, and
// a single dispatch calls no Set beyond the command line's. A flag from
// [StringsVar] gets back its values, and its next occurrence replaces them
// again. A flag whose Value does not implement flag.Getter, such as one
// from flag.Func or flag.BoolFunc, carries its own state and is not
// restored. [Invocation.Changed], [Command.Require], and
// [Command.Exclusive] see only the flags set in the current dispatch.
//
// Run panics, naming the field, when p.Root is nil, and when only one of
// p.Graph and p.LifecycleConfig is set. It panics at the start of the
// dispatch when any command in the tree has declared [Command.Use] and
// p.Graph is nil, whichever command is selected, and when a flag being
// restored from the previous dispatch refuses, in its Set, the String it
// had before that dispatch.
func (p Program) Run(ctx context.Context, args []string, streams Streams) int {
	p.check()
	root := p.Root
	prepareTree(root, p.Graph != nil)
	root.restoreFlags()
	cmd := root
	var path []*Command
	changed := make(map[string]bool)
	for {
		path = append(path, cmd)
		rest, set, undo, code, ok := parse(cmd, args, streams)
		root.undo = append(root.undo, undo...)
		if !ok {
			return code
		}
		// The leaf sees its own flags and the root flags, given at
		// whichever level they were parsed. A nested parent's local flags
		// are left out, so a leaf flag of the same name is not shadowed.
		for _, name := range set {
			if !cmd.isParent() || root.Flags().Lookup(name) != nil {
				changed[name] = true
			}
		}
		if !cmd.isParent() {
			return execute(ctx, p, path, rest, changed, streams)
		}
		if len(rest) == 0 {
			return process.Usage(streams.Stdout, help(cmd))
		}
		sub := cmd.child(rest[0])
		if sub == nil {
			msg := fmt.Sprintf("%s: unknown command %q\n\n%s", cmd.path(), rest[0], help(cmd))
			return process.Usage(streams.Stderr, msg)
		}
		cmd, args = sub, rest[1:]
	}
}

// check panics, naming the field, when p.Root is nil or when only one of
// p.Graph and p.LifecycleConfig is set.
func (p Program) check() {
	switch {
	case p.Root == nil:
		panic("cli: Program.Root is nil")
	case p.Graph != nil && p.LifecycleConfig == nil:
		panic("cli: Program.LifecycleConfig is nil but Program.Graph is set")
	case p.Graph == nil && p.LifecycleConfig != nil:
		panic("cli: Program.Graph is nil but Program.LifecycleConfig is set")
	}
}

// parse parses args against cmd's flags and returns the positional
// arguments left, the names of the flags args set, and for each of those
// the dispatcher restores, a function that puts it back to its value
// before the parse. The undo functions are returned even when parsing
// fails part way, and then also cover a flag whose malformed value its
// Set wrote anyway. It parses with a flag set made for this dispatch that
// shares the Values of cmd's own, so the flags set count only this
// dispatch's, which a flag.FlagSet parsed before would not forget. A
// parent's flags precede its subcommand, so parsing stops at the first
// non-flag argument, which names it; a leaf's flags may come anywhere
// among its positional arguments, up to a "--". When parsing ends the
// dispatch, because help was asked for or a flag was wrong, it reports
// that and returns ok false with the exit code.
func parse(cmd *Command, args []string, streams Streams) (rest, set []string, undo []func(), code int, ok bool) {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	undoers := make(map[string]func())
	before := make(map[string]string) // each flag's String before the parse
	cmd.Flags().VisitAll(func(f *flag.Flag) {
		fs.Var(f.Value, f.Name, f.Usage)
		undoers[f.Name] = undoer(cmd, f)
		before[f.Name] = f.Value.String()
	})
	var positional []string
	if !cmd.isParent() {
		args, positional = interleave(fs, args)
	}
	err := fs.Parse(args)
	fs.Visit(func(f *flag.Flag) {
		set = append(set, f.Name)
		if u := undoers[f.Name]; u != nil {
			undo = append(undo, u)
		}
	})
	if err != nil {
		// A flag whose Set failed is not visited, yet the flag package's
		// numeric Values write their variable before reporting the error,
		// so any other restorable flag whose String changed is kept too.
		fs.VisitAll(func(f *flag.Flag) {
			if u := undoers[f.Name]; u != nil && !slices.Contains(set, f.Name) && f.Value.String() != before[f.Name] {
				undo = append(undo, u)
			}
		})
	}
	switch {
	case errors.Is(err, flag.ErrHelp):
		return nil, nil, undo, process.Usage(streams.Stdout, help(cmd)), false
	case err != nil:
		return nil, nil, undo, usageError(cmd, streams.Stderr, err), false
	}
	if cmd.isParent() {
		return fs.Args(), set, undo, process.ExitOK, true
	}
	// interleave left only flags and their values, so fs.Args() is empty.
	return positional, set, undo, process.ExitOK, true
}

// execute checks the leaf's positional argument count, its flag groups,
// and its input with Validate, runs the root's PreRun and then the leaf,
// under a lifecycle when the path declares nodes with Use, and maps the
// first error to an exit code. path holds the commands from the root to
// the leaf, and changed the flags the leaf sees set on the command line.
func execute(ctx context.Context, p Program, path []*Command, args []string, changed map[string]bool, streams Streams) int {
	root, cmd := path[0], path[len(path)-1]
	if cmd.Args != nil {
		if err := cmd.Args(args); err != nil {
			return usageError(cmd, streams.Stderr, err)
		}
	}
	inv := &Invocation{Streams: streams, Args: args, cmd: cmd, uses: uses(path), changed: changed}
	if err := cmd.checkFlags(inv.changed); err != nil {
		return usageError(cmd, streams.Stderr, err)
	}
	if cmd.Validate != nil {
		// Validate's error is a usage error whatever its type.
		if err := cmd.Validate(inv); err != nil {
			return usageError(cmd, streams.Stderr, err)
		}
	}
	var err error
	if root.PreRun != nil {
		err = root.PreRun(ctx, inv)
	}
	if err == nil {
		err = p.run(ctx, cmd, inv)
	}
	if err == nil {
		return process.ExitOK
	}
	if _, ok := errors.AsType[*UsageError](err); ok {
		return usageError(cmd, streams.Stderr, err)
	}
	return process.Fail(streams.Stderr, cmd.path(), err)
}

// run runs the leaf cmd with inv. When inv's path declares no nodes it
// calls Run directly; otherwise it builds them and the lifecycle
// configuration node, and runs cmd under a Coordinator for the System,
// with the System set on inv for [Invocation.Get]. The error is the
// Build's, or the Coordinator's: startup's or the leaf's, joined with the
// shutdown's.
//
// A monitored failure is reported once. The leaf runs under a context
// whose [context.Cause], once a monitored failure ends it, is that failure
// labelled "run:", as the Coordinator labels it. When the leaf's error is
// only its context's end, it is replaced by that cause; when the leaf's
// error then carries the failure, by returning the cause or wrapping it,
// that error stands in place of the Coordinator's own "run:" report. A
// leaf error unrelated to the failure is reported beside it, and a
// context the caller ended is reported as its own error.
func (p Program) run(ctx context.Context, cmd *Command, inv *Invocation) error {
	if len(inv.uses) == 0 {
		return cmd.Run(ctx, inv)
	}
	sys, err := p.Graph.Build(append(slices.Clone(inv.uses), p.LifecycleConfig)...)
	if err != nil {
		return err
	}
	var leafErr, failure error // failure is the monitored failure the leaf's error carries
	err = lifecycle.New(sys, sys.Get(p.LifecycleConfig)).Exec(ctx, func(runCtx context.Context) error {
		inv.system = sys
		leafCtx, stop := runContext(ctx, runCtx)
		defer stop()
		leafErr = cmd.Run(leafCtx, inv)
		if leafErr != nil && leafErr == leafCtx.Err() {
			leafErr = context.Cause(leafCtx)
		}
		// The run context ended while the caller's did not: the
		// Coordinator failed it with a monitored failure, its cause.
		if ctx.Err() == nil && runCtx.Err() != nil && errors.Is(leafErr, context.Cause(runCtx)) {
			failure = context.Cause(runCtx)
			return nil
		}
		return leafErr
	})
	if failure == nil {
		return err
	}
	return errors.Join(leafErr, without(err, failure))
}

// runContext returns the context the leaf runs under: one that carries
// runCtx's values and deadline and ends when runCtx does, and whose cause,
// when runCtx ends while ctx, the caller's, has not, is runCtx's cause
// labelled "run:", the Coordinator's label for a monitored failure. A
// context derived from runCtx would carry the cause unlabelled, so the
// context is detached from runCtx's cancellation and ended, with the
// labelled cause, by a function run after runCtx ends. stop releases the
// context.
func runContext(ctx, runCtx context.Context) (leaf context.Context, stop func()) {
	detached, cancel := context.WithCancelCause(context.WithoutCancel(runCtx))
	stopAfter := context.AfterFunc(runCtx, func() {
		cause := context.Cause(runCtx)
		if ctx.Err() == nil {
			cause = fmt.Errorf("run: %w", cause)
		}
		cancel(cause)
	})
	return leafContext{Context: detached, run: runCtx}, func() {
		stopAfter()
		cancel(nil)
	}
}

// leafContext is the context runContext returns: Context, detached from
// run's cancellation and ended after run with the labelled cause, which
// [context.Cause] reads from it, with run's deadline.
type leafContext struct {
	context.Context
	run context.Context
}

// Deadline returns run's deadline, which ends the context with run.
func (c leafContext) Deadline() (time.Time, bool) {
	return c.run.Deadline()
}

// Err reports run's error once run has ended, after waiting for the
// function that ends the context after it, so the context ends as soon as
// run does, with the cause in place: an action that ends the run's context
// and returns sees the end at once. It is run's error, not the detached
// context's, which is always [context.Canceled], so a caller's deadline
// reports [context.DeadlineExceeded], as [leafContext.Deadline] implies.
func (c leafContext) Err() error {
	if err := c.run.Err(); err != nil {
		<-c.Done()
		return err
	}
	return c.Context.Err()
}

// without returns err with the Coordinator's report of the monitored
// failure, the error that wraps failure, removed from the errors err
// joins, and nil when nothing is left.
func without(err, failure error) error {
	if err == nil || errors.Is(errors.Unwrap(err), failure) {
		return nil
	}
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return err
	}
	var kept []error
	for _, e := range joined.Unwrap() {
		kept = append(kept, without(e, failure))
	}
	return errors.Join(kept...)
}

// uses returns the union of the nodes declared with Use along path, root
// first, each node once in the order first declared, in a slice of its
// own. It is the one rule for a leaf's set: execute builds it, and
// [Command.Leaves] yields it.
func uses(path []*Command) []graph.Ref {
	var refs []graph.Ref
	for _, cmd := range path {
		for _, r := range cmd.uses {
			if !slices.Contains(refs, r) {
				refs = append(refs, r)
			}
		}
	}
	return refs
}

// usageError reports err with cmd's short usage on stderr and returns
// ExitUsage.
func usageError(cmd *Command, stderr io.Writer, err error) int {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %v\n", cmd.path(), err)
	fmt.Fprintf(&b, "Usage: %s\n", usageLine(cmd))
	fmt.Fprintf(&b, "Run '%s --help' for details.", cmd.path())
	return process.Usage(stderr, b.String())
}
