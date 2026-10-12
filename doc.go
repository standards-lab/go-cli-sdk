// Package cli is a command dispatcher on the standard library's flag
// package. A program builds a tree of [Command] values, each owning a
// *flag.FlagSet, and hands the root to [Run] with the process arguments and
// its [Streams], standard input, output, and error; Run walks the tree to the
// selected command, parses its flags, runs it, and returns the exit code the
// program passes to os.Exit, following go-core's process convention. Wiring
// mistakes in the tree panic when it is built or when it is dispatched, as
// each symbol's documentation states.
//
// The package exports:
//
//   - [Command], one node of a command tree, a parent or a leaf; its
//     Footer appends the command's own text to its generated help, its
//     Args counts a leaf's positional arguments, its Validate checks a
//     leaf's input, arguments and flag values, before anything is built,
//     and the root's PreRun runs work every command shares
//   - [Command.Add], which attaches subcommands
//   - [Command.Flags], which returns the command's flag set; flags defined
//     on the root are root flags, accepted at any depth
//   - [Command.Require], which marks a leaf's flags as required
//   - [Command.Exclusive], which declares a mutually exclusive flag group
//   - [Command.Use], which declares the graph nodes a command needs
//   - [Streams], the standard input, output, and error a dispatch uses
//   - [Invocation], what a running command receives: its arguments and
//     the Streams Run was given, which it embeds
//   - [Invocation.Get], which returns the value of a node the command's
//     path declares, and panics naming the command and the node when the
//     path does not declare it
//   - [Invocation.Changed], which reports whether a flag was given
//   - [Run], which dispatches the arguments and returns the exit code
//   - [Option], which configures one Run, and [WithGraph], the Option that
//     gives Run the graph that declared nodes are built from
//   - [StringsVar], which defines a repeatable string flag
//   - [NoArgs] and [ExactArgs], the positional-argument validators
//   - [UsageError], an error reported with the command's usage, with
//     [UsageError.Error] and [UsageError.Unwrap]
//   - [Usagef], which returns a formatted UsageError
//
// # Help and flags
//
// Every command's help is generated from the tree and printed on -h or
// --help, or when a parent runs with no subcommand or an unknown one; a
// command's Footer appends its own text. There is no help command and no
// completion command, and flags have long names only: the dispatcher has
// no shorthand flags, with -h beside --help the one exception. Requested
// help and every usage error return go-core's ExitUsage; any other failure
// returns ExitFailure, and success ExitOK.
//
// # Dependencies
//
// A command's dependencies are part of the command: [Command.Use] declares
// the [graph.Node] values it needs, inherited along the command path, and
// [Run] builds only the union a leaf's path declares, from the [WithGraph]
// graph, and runs the leaf under a [lifecycle.Coordinator]'s Exec only once
// the dispatch reaches it, so help, a usage error, or a PreRun error builds
// nothing. The leaf reads each value with [Invocation.Get]. A leaf whose
// path declares no nodes runs with no Build and no lifecycle.
//
// Under the lifecycle, a built value that implements [lifecycle.Monitored]
// ends the running command's context when it reports a failure, and Run
// reports that failure with go-core's "run:" label; a failure to start or
// to shut down the built System is reported with "startup:" or
// "shutdown:". Each returns ExitFailure.
//
// # Dispatch order
//
// A dispatch to a leaf parses its flags and then runs, stopping at the
// first failure: [Command.Args], the [Command.Require] and
// [Command.Exclusive] checks, [Command.Validate], the root's
// [Command.PreRun], the Build of the declared nodes, and [Command.Run].
// Every failure before PreRun is a usage error, and all of them come
// before the Build, so input the dispatcher or Validate refuses builds
// nothing.
//
// The package depends on the standard library and go-core only, whose
// graph, lifecycle, and process packages it builds, runs, and exits with;
// depguard holds that line in the check.
package cli
