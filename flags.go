package cli

// See docs/dispatch.md

import (
	"flag"
	"fmt"
	"slices"
	"strings"
)

func prepareTree(root *Command, hasGraph bool) {
	var walk func(c *Command)
	walk = func(c *Command) {
		if c.Args != nil && c.isParent() {
			panic(fmt.Sprintf("cli: %s: Args set on a parent command", c.path()))
		}
		if c.Validate != nil && c.isParent() {
			panic(fmt.Sprintf("cli: %s: Validate set on a parent command", c.path()))
		}
		if c.PreRun != nil && c != root {
			panic(fmt.Sprintf("cli: %s: PreRun set below the root", c.path()))
		}
		if len(c.uses) > 0 && !hasGraph {
			panic(fmt.Sprintf("cli: %s: Use declared but Program.Graph is nil", c.path()))
		}
		c.checkWiring()
		for _, sub := range c.children {
			sub.inherit(root.Flags())
			walk(sub)
		}
	}
	walk(root)
}

func (c *Command) inherit(from *flag.FlagSet) {
	from.VisitAll(func(f *flag.Flag) {
		if c.inherited[f.Name] {
			return
		}
		if c.Flags().Lookup(f.Name) != nil {
			panic(fmt.Sprintf("cli: %s: flag --%s redefines a root flag", c.path(), f.Name))
		}
		c.Flags().Var(f.Value, f.Name, f.Usage)
		if c.inherited == nil {
			c.inherited = make(map[string]bool)
		}
		c.inherited[f.Name] = true
	})
}

func (c *Command) restoreFlags() {
	for _, undo := range slices.Backward(c.undo) {
		undo()
	}
	c.undo = nil
}

func undoer(cmd *Command, f *flag.Flag) func() {
	if v, ok := f.Value.(*stringsValue); ok {
		values, replacedDefault := slices.Clone(*v.p), v.replacedDefault
		return func() { *v.p, v.replacedDefault = values, replacedDefault }
	}
	if _, ok := f.Value.(flag.Getter); !ok {
		return nil
	}
	value := f.Value.String()
	return func() {
		if err := f.Value.Set(value); err != nil {
			panic(fmt.Sprintf("cli: %s: flag --%s cannot be restored to %q: %v", cmd.path(), f.Name, value, err))
		}
	}
}

func interleave(fs *flag.FlagSet, args []string) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return flags, append(positional, args[i+1:]...)
		case len(arg) < 2 || arg[0] != '-':
			positional = append(positional, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimPrefix(arg[1:], "-")
		if strings.Contains(name, "=") {
			continue
		}
		if f := fs.Lookup(name); f != nil && !isBool(f) && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return flags, positional
}

func isBool(f *flag.Flag) bool {
	b, ok := f.Value.(interface{ IsBoolFlag() bool })
	return ok && b.IsBoolFlag()
}

// StringsValue returns a repeatable string flag.Value bound to p, for
// registering with fs.Var:
//
//	fs.Var(cli.StringsValue(&tags, []string{"a"}), "tag", "tag to attach")
//
// Each occurrence appends its value to *p, in the order given, wherever it
// falls among the positional arguments. A value is taken whole and never
// split on commas. StringsValue sets *p to a copy of value before it
// returns, so fs.Var records that default for help and *p never aliases
// value; the first occurrence replaces the default rather than appending to
// it. Help lists the flag as "--name string" with "(repeatable)" after its
// usage.
func StringsValue(p *[]string, value []string) flag.Value {
	*p = append([]string(nil), value...)
	return &stringsValue{p: p}
}

type stringsValue struct {
	p               *[]string
	replacedDefault bool
}

func (v *stringsValue) Set(s string) error {
	if !v.replacedDefault {
		*v.p = nil
		v.replacedDefault = true
	}
	*v.p = append(*v.p, s)
	return nil
}

func (v *stringsValue) String() string {
	if v == nil || v.p == nil {
		return "[]"
	}
	return "[" + strings.Join(*v.p, ",") + "]"
}
