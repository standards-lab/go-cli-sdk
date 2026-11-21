package cli

import (
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
)

func help(cmd *Command) string {
	var b strings.Builder
	if cmd.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", cmd.Summary)
	}
	fmt.Fprintf(&b, "Usage:\n  %s\n", usageLine(cmd))

	var commands []row
	for _, sub := range cmd.children {
		commands = append(commands, row{sub.Name, sub.Summary})
	}
	writeSection(&b, "Commands", commands)

	own := flagsOf(cmd, func(name string) bool { return !cmd.inherited[name] })
	own = append(own, row{"--help", "Show help for " + cmd.path()})
	writeSection(&b, "Flags", own)
	writeSection(&b, "Global flags", flagsOf(cmd, func(name string) bool { return cmd.inherited[name] }))

	if len(cmd.children) > 0 {
		fmt.Fprintf(&b, "\nRun '%s <command> --help' for help on a command.\n", cmd.path())
	}
	if cmd.Footer != nil {
		var footer strings.Builder
		cmd.Footer(&footer)
		if footer.Len() > 0 {
			fmt.Fprintf(&b, "\n%s", footer.String())
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func usageLine(cmd *Command) string {
	parts := []string{cmd.path()}
	if len(cmd.children) > 0 {
		parts = append(parts, "<command>")
	}
	parts = append(parts, "[flags]")
	if cmd.Synopsis != "" {
		parts = append(parts, cmd.Synopsis)
	}
	return strings.Join(parts, " ")
}

type row struct {
	name string
	text string
}

func flagsOf(cmd *Command, keep func(name string) bool) []row {
	var lines []row
	cmd.Flags().VisitAll(func(f *flag.Flag) {
		if keep(f.Name) {
			lines = append(lines, describe(f, cmd.isRequired(f.Name)))
		}
	})
	return lines
}

func describe(f *flag.Flag, required bool) row {
	kind, usage := flag.UnquoteUsage(f)
	_, repeatable := f.Value.(*stringsValue)
	if repeatable {
		if kind == "value" {
			kind = "string"
		}
		usage += " (repeatable)"
	}
	if required {
		usage += " (required)"
	}
	name := "--" + f.Name
	if kind != "" {
		name += " " + kind
	}
	switch f.DefValue {
	case "", "false", "0", "[]":
	default:
		if kind == "string" && !repeatable {
			usage += fmt.Sprintf(" (default %q)", f.DefValue)
		} else {
			usage += fmt.Sprintf(" (default %s)", f.DefValue)
		}
	}
	return row{name: name, text: usage}
}

func writeSection(b *strings.Builder, title string, rows []row) {
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s:\n", title)
	tw := tabwriter.NewWriter(b, 0, 0, 3, ' ', 0)
	for _, r := range rows {
		_, _ = fmt.Fprintf(tw, "  %s\t%s\n", r.name, r.text)
	}
	_ = tw.Flush()
}
