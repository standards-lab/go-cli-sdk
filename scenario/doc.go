// Package scenario runs a tool's narrated scenarios, its guided tours of
// its own commands, as "<tool> scenario <name>" leaves. Each scenario is an
// ordered list of steps over the graph nodes it declares; the package holds
// no scenarios itself, only the runner, the narration, and the listing.
//
// The package exports:
//
//   - [Scenario], one narrated tour: its name, its summary, the graph
//     nodes it declares, and its steps
//   - [Step], one beat of the narration: an intent sentence and the
//     action that does it
//   - [Reporter], which writes the narration, with [Reporter.Note] for a
//     step's prose and [Reporter.Show] for the output its operation
//     produced
//   - [Commands], which builds the "scenario" parent with one leaf per
//     scenario, for a root to mount
//   - [WriteListing], which writes the listing of the scenarios that the
//     scenario parent's help ends with
//
// # Opting in
//
// A tool keeps its scenarios in one root-level package named scenarios,
// and opts in by mounting them at its root with
// root.Add(scenario.Commands(scenarios...)...). So that the root's help
// ends with the listing too, the root's own Footer calls
// [WriteListing] with the same scenarios.
//
// Each scenario's leaf declares the scenario's nodes with
// [cli.Command.Use], and the parent declares none, so the dispatcher builds
// and starts only the nodes the scenario declares, before its first step:
// a dependency that cannot be reached fails the run at start, naming its
// node. A step's action reads the built values with [cli.Invocation.Get].
//
// # Narration
//
// The narration is plain text on the command's standard output, with no
// color. Each step opens with a heading of its number, the step count, and
// its intent, set off from the step before it by a blank line; under it,
// [Reporter.Note] prints prose wrapped at 80 columns and indented two
// spaces, and [Reporter.Show] prints a block indented four, so a step's
// result reads apart from its narration:
//
//	[2/3] List the files
//	  The list command reads the store and prints one file per line.
//	    notes.txt
//	    todo.txt
//
// The steps run in order and stop at the first action that returns an
// error, which the leaf returns as "step i (intent): err"; the dispatcher
// labels it with the command's path, so the report names the scenario and
// the step, and the run exits with go-core's ExitFailure.
//
// [Commands] panics on a wiring mistake: no scenarios, two scenarios with
// one name, or a scenario with no steps.
package scenario
