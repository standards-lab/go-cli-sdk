# Changelog

All notable changes to `github.com/standards-lab/go-cli-sdk` are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the module adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/standards-lab/go-cli-sdk/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/standards-lab/go-cli-sdk/releases/tag/v0.1.0
