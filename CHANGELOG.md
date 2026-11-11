# Changelog

All notable changes to `github.com/standards-lab/go-cli-sdk` are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the module adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.3.0] - 2026-11-11

Moves go-cli-sdk onto go-core v0.11.0. A command tree is dispatched through `cli.Program`, which
names its root, graph, and lifecycle configuration node as fields; `Command.Leaves` lets a
composition root's layer test build every leaf's set; and a monitored failure is reported on one
line naming the node that failed.

### Added

- `cli.Program{Root, Graph, LifecycleConfig}`, a command tree with the graph its declared nodes
  are built from and the node that configures their lifecycle, and `Program.Run(ctx, args,
  streams)`, which dispatches the arguments and returns the exit code. `Run` panics, naming the
  field, on a nil `Root` or when only one of `Graph` and `LifecycleConfig` is set, and at the
  start of a dispatch when any command in the tree declares `Use` and `Graph` is nil.
- `(*Command).Leaves()`, which yields each leaf at or below a command with its command path and
  the nodes a dispatch to it builds, apart from the `LifecycleConfig` every Build adds, for a
  composition root's layer test.

### Changed

- **Breaking:** a `Monitored` value's failure while a leaf or a scenario step runs is reported on
  one line, `<path>: run: <node>: <error>`, or for a scenario
  `<path>: step N (<intent>): run: <node>: <error>`, in place of `context canceled` followed by
  `run: <error>`. The leaf's context's `context.Cause` is that labelled failure, so a leaf that
  returns only its context's end, or wraps the cause, is reported once; an unrelated leaf error
  is reported beside it.
- **Breaking:** requires go-core v0.11.0, up from v0.8.0, and carries its breaks to a program
  that defines its graph:
  - `graph.Scope.After` is removed; a node's order-only edges are declared at `Define` with
    `graph.After`.
  - `graph.Graph.Observe` takes `func(graph.Ref)`, the node's handle, in place of
    `func(string)`.
  - `graph.Graph.Define`, `Value`, `Fresh`, and `Group` panic on a name containing a comma or a
    newline.
  - `graph.Dependency` gains a `Ref` field, so a composite literal without field names no longer
    compiles.
  - `lifecycle`: a `Monitored` value's failure ends the run as `run: <node>: <error>`; a channel
    registered with `Coordinator.Monitor` still reports `run: <error>`.

  go-core v0.9.0 and v0.10.0 add the `reactor` and `event` packages and break nothing.

### Removed

- **Breaking:** `cli.Run`, `cli.Option`, and `cli.WithGraph`, in favor of `cli.Program`:
  `cli.Run(ctx, root, args, streams, cli.WithGraph(g, cfg))` becomes
  `cli.Program{Root: root, Graph: g, LifecycleConfig: cfg}.Run(ctx, args, streams)`, and
  `cli.Run(ctx, root, args, streams)` becomes `cli.Program{Root: root}.Run(ctx, args, streams)`.

### Migrating from v0.2.0

- **`cli.Run` becomes `cli.Program{...}.Run`.** Move the root and `WithGraph`'s two arguments
  into the Program's fields; the arguments and streams are passed to `Run` as before.
- **A test that matches a monitored failure's stderr** expects one line, the command's path,
  then `run: `, the node's name, and the error, with no `context canceled` line before it.
- **The graph a program defines follows go-core's "Migrating from v0.10.0".** `s.After(x)`
  becomes `graph.After(x)` on the Define line, an observer takes a `graph.Ref`, and a name with
  a comma or a newline must change.

## [v0.2.0] - 2026-10-12

Adds package `scenario`, which runs a tool's narrated scenarios, its guided tours of its own
commands, on package `cli`.

### Added

- `scenario.Scenario`, one narrated tour: a name, a summary, the graph nodes it declares, and
  its ordered `scenario.Step`s, each an intent sentence and an action over the declared nodes.
- `scenario.Commands`, which builds the `scenario` parent with one leaf per scenario, for a root
  to mount with `root.Add`; each leaf declares its scenario's nodes with `Use`, so a scenario
  builds only those. The steps run in order and stop at the first that fails, reported as
  `step i (intent): err`. A step fails when its action returns an error, or when the command's
  context has ended by the time its action returns nil. `Commands` panics on a wiring mistake:
  no scenarios, a duplicate name, or a scenario with no steps.
- `scenario.WriteListing`, which writes the listing of the scenarios and the nodes each
  declares; the scenario parent's help ends with it, and a root's `Footer` can call it.
- `scenario.Reporter`, which narrates a run as plain text: a `[i/n] intent` heading per step,
  `Reporter.Note` for wrapped prose, and `Reporter.Show` for a step's output, indented apart.

## [v0.1.0] - 2026-10-12

The first release: package `cli`, a command dispatcher on the standard library's `flag` whose
commands build only the go-core graph nodes they declare, on go-core v0.8.0.

### Added

- `cli.Command`, one node of a command tree, a parent or a leaf, with `Add` for subcommands and
  `Flags` for its `*flag.FlagSet`. Flags have long names only; flags on the root are accepted at
  any depth, and a leaf takes its flags among its positional arguments, up to `--`. `NoArgs` and
  `ExactArgs` count a leaf's positional arguments, and the root's `PreRun` runs work every
  command shares.
- Generated help on `-h`, `--help`, and a parent run with no subcommand or an unknown one, with
  `Command.Footer` appending a command's own text. There is no help or completion command.
- `Command.Require` for required flags, `Command.Exclusive` for mutually exclusive groups,
  `Invocation.Changed` for whether a flag was given, and `cli.StringsVar` for a repeatable
  string flag.
- `Command.Validate`, which checks a leaf's arguments and flag values as a whole before anything
  is built; its error is always a usage error.
- Per-command composition on go-core v0.8.0: `Command.Use` declares the `graph` nodes a command
  needs, inherited along its path; `cli.WithGraph` gives `Run` the graph and the
  `lifecycle.Config` node; `Run` builds only the union the leaf's path declares and runs the leaf
  under `lifecycle.Coordinator.Exec`; `Invocation.Get` returns a declared node's value. A
  `Monitored` value's failure ends the command's context and is reported with `run:`.
- `cli.Streams`, the standard input, output, and error a dispatch uses, embedded in the
  `Invocation`, and `cli.Run`, which returns go-core's exit codes: `ExitUsage` for requested
  help and usage errors (`cli.UsageError`, `cli.Usagef`), `ExitFailure` for any other error,
  and `ExitOK`. A tree may be dispatched more than once, and each dispatch starts from the flag
  values held before the previous one.
- CI runs `mise run check` on every pull request and push to `main`; a `v*` tag publishes a
  GitHub release.

[Unreleased]: https://github.com/standards-lab/go-cli-sdk/compare/v0.3.0...HEAD
[v0.3.0]: https://github.com/standards-lab/go-cli-sdk/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/standards-lab/go-cli-sdk/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/standards-lab/go-cli-sdk/releases/tag/v0.1.0
