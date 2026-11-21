# How the dispatcher works

`doc.go` states what `cli` and `scenario` promise their callers. This page holds the reasoning
behind their unexported machinery: how a dispatch shares, parses, and restores flags, and how a
monitored failure reaches the running command.

## Root flags and wiring checks

Every dispatch first prepares the whole tree (`prepareTree`). Each flag defined on the root's flag
set is defined again on the flag set of every command below it with the same `flag.Value`
(`Command.inherit`), so a root flag is accepted at any depth and sets the one value the root
defined. Only the root's flags are shared: a flag defined on a nested parent stays local to that
level, and a leaf sees its own flags and the root flags but not a nested parent's, so a leaf flag
of the same name is not shadowed.

Preparing the whole tree on every dispatch makes a wiring mistake panic on the first run, whatever
path the user takes: a command that defines a flag with a root flag's name, a parent with an `Args`
or `Validate` function it would never call, a `PreRun` below the root, a flag requirement or group
the command cannot honour (checked once the command's flag set holds the root flags it inherits),
or a `Use` on any command when the `Program` has no `Graph`. Sharing is idempotent, so a tree can be
dispatched more than once.

A command with neither subcommands nor `Run` counts as a parent, so running it prints its help
instead of doing nothing.

## Parsing

Each dispatch parses with a flag set made for that dispatch which shares the Values of the
command's own (`parse`), so the flags it reports as set count only this dispatch's; a
`flag.FlagSet` parsed before would not forget them. A parent's flags precede its subcommand, so
parsing stops at the first non-flag argument, which names it. A leaf's flags may come anywhere
among its positional arguments (`interleave`), classified by the flag package's own rules: `--`
ends the flags and everything after it is positional; `-` and any argument not starting with `-`
is positional; a flag without `=` whose Value is not a boolean takes the next argument as its
value, whatever it looks like. Unknown and malformed flags stay flag arguments, so the flag set
reports them as it would have.

## Restoring flags between dispatches

Because a tree can be dispatched more than once, as a test or a scenario run does, each dispatch
first undoes the flags the previous one set (`Command.restoreFlags`), latest first, so a root flag
set at two levels ends at the value it held before either; a flag the previous dispatch did not
set is never touched. Each undo function (`undoer`) puts its flag back to the value it held before
the parse:

- A `StringsValue` flag gets back its values and whether its next occurrence replaces them.
- Any other flag that holds a value, which the flag package's own flags show by implementing
  `flag.Getter`, is `Set` to its earlier `String`; the undo panics when `Set` refuses that string.
- A Value without `Get`, such as one from `flag.Func`, carries its own state and is not restored.

The undo functions are kept even when parsing fails part way. A flag whose `Set` failed is not
visited, yet the flag package's numeric Values write their variable before reporting the error, so
any other restorable flag whose `String` changed during the parse is restored too.

A `StringsValue` renders an empty or unbound value as `[]`, which help treats as no default, and
help gives it the placeholder `string`, since each occurrence takes one.

## The leaf's context and a monitored failure

When a leaf's path declares nodes, the leaf runs under go-core's `lifecycle.Coordinator`, which
ends the run's context when a `lifecycle.Monitored` value reports a failure and reports that
failure labelled `run:`. The aim is that a monitored failure is reported once.

- **The leaf's context carries the labelled cause.** A context derived from the run's context
  would carry the failure unlabelled, so the leaf's context (`runContext`, `leafContext`) is
  detached from the run context's cancellation and is cancelled, by a `context.AfterFunc`, with
  the failure labelled `run:` (`runFailure`). Its values and deadline are the run context's, and
  its `Err` is the run context's error, read after the cancelling function has run, so the cause is
  in place whenever the end is seen and a caller's deadline still reports
  `context.DeadlineExceeded`.
- **The cause matches the end.** `runFailure` also matches `context.Canceled`, the end it gives
  the context, so the leaf, or a scenario step, may report the cause in place of that end.
- **Only a monitored failure replaces the end.** A leaf error that is only its context's end is
  replaced by the cause only when the run context ended while the caller's did not. A caller's
  own cause does not match the end and is left out, so a context the caller ended reports
  `context.Canceled` or `context.DeadlineExceeded`.
- **One report.** When the leaf's error carries the failure, because it returned or wrapped the
  cause, that error replaces the Coordinator's own `run:` report (`without`). A leaf error
  unrelated to the failure is reported beside it.

A scenario step follows the same rule (`scenario.run`): after each action the runner checks the
context itself, so an action that ignores it, or ends it and returns nil, still fails its step
before the next heading prints, and a step whose error is only the context's end fails with the
context's cause when that cause matches the end.

## A leaf's nodes

`uses` is the one rule for the nodes a leaf's path declares: the union along the path, root
first, each node once in the order first declared. The dispatcher builds that set, and
`Command.Leaves` yields it.
