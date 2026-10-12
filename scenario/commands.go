package scenario

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/standards-lab/go-cli-sdk"
)

// Commands builds the scenario command surface over scenarios: the
// "scenario" parent, its one element, returned as a slice so a root mounts
// it as it mounts any package's commands, with
// root.Add(scenario.Commands(...)...). The parent has one leaf per
// scenario, in the order given, each declaring its own nodes with Use; the
// parent declares none, so a scenario brings up only what it declares.
// Run alone, the parent prints its help, which ends with the listing
// [WriteListing] writes.
//
// A leaf runs its scenario's steps in order, each under a heading its
// [Reporter] prints, and stops at the first step that fails, which it
// returns as "step i (intent): err"; the dispatcher labels it with the
// command's path, so the report names the scenario and the step, and the
// run exits with go-core's ExitFailure. A step fails when its action
// returns an error, or when the command's context has ended by the time
// its action returns, whatever the action returned: the runner checks the
// context after each action, not the action, and err is then the
// context's error. No heading or step runs after the step that fails.
//
// Commands panics on a wiring mistake: no scenarios, two scenarios with
// one name, or a scenario with no steps, naming the scenario.
func Commands(scenarios ...Scenario) []*cli.Command {
	if len(scenarios) == 0 {
		panic("scenario: Commands with no scenarios")
	}
	parent := &cli.Command{
		Name:    "scenario",
		Summary: "Run a narrated scenario over the commands",
		Footer:  func(w io.Writer) { WriteListing(w, scenarios...) },
	}
	seen := make(map[string]bool, len(scenarios))
	for _, s := range scenarios {
		if seen[s.Name] {
			panic(fmt.Sprintf("scenario: duplicate scenario %q", s.Name))
		}
		seen[s.Name] = true
		if len(s.Steps) == 0 {
			panic(fmt.Sprintf("scenario: scenario %q has no steps", s.Name))
		}
		parent.Add(command(s))
	}
	return []*cli.Command{parent}
}

// WriteListing writes the listing of scenarios to w, as a help Footer
// prints it: a Scenarios heading, then for each scenario, in the order
// given, a line with its name and summary, followed by a uses line naming
// each node the scenario declares once, in the order it first declares
// it, as the dispatcher builds each once. A scenario that declares no
// nodes gets no uses line. It is
// the one source of the listing, which the scenario parent's help ends
// with, and which a root's Footer can call so the root's help ends with it
// too. It prints what a scenario declares, not the nodes those reach: the
// graph discovers a node's dependencies only by running its constructor,
// and the listing builds nothing.
//
// The columns are laid out as package cli lays out a help section, by a
// tabwriter with a padding of 3, so in the scenario parent's help the
// listing's names line up with the Commands section above it, where the
// same names appear.
func WriteListing(w io.Writer, scenarios ...Scenario) {
	_, _ = io.WriteString(w, "Scenarios:\n")
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, s := range scenarios {
		var names []string
		for _, n := range s.Nodes {
			if !slices.Contains(names, n.Name()) {
				names = append(names, n.Name())
			}
		}
		_, _ = fmt.Fprintf(tw, "  %s\t%s\n", s.Name, s.Summary)
		if len(names) > 0 {
			_, _ = fmt.Fprintf(tw, "  \tuses %s\n", strings.Join(names, ", "))
		}
	}
	_ = tw.Flush()
}
