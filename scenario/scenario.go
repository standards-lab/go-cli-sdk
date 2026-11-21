package scenario

// See docs/dispatch.md

import (
	"context"
	"errors"
	"fmt"

	"github.com/standards-lab/go-cli-sdk"
	"github.com/standards-lab/go-core/graph"
)

// Scenario is one narrated tour: what it shows, the graph nodes it
// declares, and the ordered steps that show it. It states no preconditions
// of its own: the nodes are its dependencies, which its command declares
// with Use, so the dispatcher builds and starts exactly those before the
// first step, and a dependency that cannot be reached fails the run at
// start, naming its node.
type Scenario struct {
	Name    string      // the word after "scenario" that runs it
	Summary string      // the line the scenario parent's help and the listing print
	Nodes   []graph.Ref // the nodes the command declares, in order, which the listing names
	Steps   []Step      // the beats the command runs, in order; at least one
}

// Step is one beat of the narration: the sentence saying what is about to
// happen, and the action that does it. The action closes over the nodes it
// reads and reads their values with inv.Get, from the System the
// dispatcher built and started for the scenario's command, and reports
// what it observed through r.
type Step struct {
	Intent string
	Action func(ctx context.Context, inv *cli.Invocation, r *Reporter) error
}

func command(s Scenario) *cli.Command {
	return (&cli.Command{
		Name:    s.Name,
		Summary: s.Summary,
		Args:    cli.NoArgs,
		Run: func(ctx context.Context, inv *cli.Invocation) error {
			return run(ctx, s, inv)
		},
	}).Use(s.Nodes...)
}

func run(ctx context.Context, s Scenario, inv *cli.Invocation) error {
	r := &Reporter{w: inv.Stdout}
	for i, step := range s.Steps {
		r.heading(i+1, len(s.Steps), step.Intent)
		err := step.Action(ctx, inv, r)
		if err == nil {
			err = ctx.Err()
		}
		if err != nil && err == ctx.Err() {
			if cause := context.Cause(ctx); errors.Is(cause, err) {
				err = cause
			}
		}
		if err != nil {
			return fmt.Errorf("step %d (%s): %w", i+1, step.Intent, err)
		}
	}
	return nil
}
