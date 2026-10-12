// Package scenario runs a tool's narrated scenarios, its guided tours of
// its own commands, as "<tool> scenario <name>" leaves. A scenario is an
// ordered list of steps over the graph nodes it declares. The package holds
// no scenarios itself, only the runner, the narration, and the listing.
//
// The package exports:
//
//   - [Scenario], one narrated tour: its name, its summary, the graph
//     nodes it declares, and its steps
//   - [Step], one beat of the narration: an intent sentence and the
//     action that does it
//   - [Reporter], which writes a step's narration under its heading
//   - [Reporter.Note], which prints a step's prose
//   - [Reporter.Show], which prints the output a step's operation produced
//   - [Commands], which builds the "scenario" parent with one leaf per
//     scenario, for a root to mount
//   - [WriteListing], which writes the listing of the scenarios that the
//     scenario parent's help ends with
//
// # Opting in
//
// A tool keeps its scenarios in one root-level package named scenarios
// and mounts them at its root with root.Add(scenario.Commands(scenarios...)...).
// For the root's help to end with the listing too, the root's own Footer
// calls [WriteListing] with the same scenarios. A step's action reads the
// values of the scenario's nodes with [cli.Invocation.Get].
//
// # Narration
//
// The narration is plain text on the command's standard output. Each step
// opens with a heading of its number, the step count, and its intent; under
// it, [Reporter.Note] prints the step's prose and [Reporter.Show] its
// result, indented past the prose:
//
//	[2/3] List the files
//	  The list command reads the store and prints one file per line.
//	    notes.txt
//	    todo.txt
//
// [Commands] states when a run stops at a failing step and which wiring
// mistakes panic.
package scenario
