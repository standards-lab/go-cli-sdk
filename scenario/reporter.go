package scenario

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Reporter writes a scenario's narration as plain text, with no color, to
// the scenario command's standard output. The runner prints a heading per
// step, set off from the step before it by a blank line; under it, the
// step prints prose saying what it does and why with [Reporter.Note], and
// the output its operation produced with [Reporter.Show]. The runner makes
// the Reporter each step receives; a Reporter's zero value is not usable.
type Reporter struct {
	w       io.Writer
	atBlank bool
	started bool
}

const indent = "  "

const columns = 80

func (r *Reporter) heading(i, n int, intent string) {
	r.blank()
	r.printf("[%d/%d] %s\n", i, n, intent)
}

// Note prints prose formatted as fmt.Sprintf formats it, indented two
// spaces and wrapped at 80 columns, the indent included, counting a rune
// as one column and breaking only at whitespace: one call is one description, however many lines it takes. A
// newline in the prose is a hard break, and an empty line between two
// paragraphs prints as one blank line.
func (r *Reporter) Note(format string, args ...any) {
	for _, line := range wrap(fmt.Sprintf(format, args...), columns-len(indent)) {
		if line == "" {
			r.blank()
			continue
		}
		r.printf("%s%s\n", indent, line)
	}
}

// Show prints what render writes as a block indented four spaces, past the
// prose, so a step's result reads apart from its narration: render is
// typically the same output the direct command the step stands for writes.
// A blank line in the output prints empty, with no indent. Nothing is
// printed when render fails, and its error is returned.
func (r *Reporter) Show(render func(io.Writer) error) error {
	var buf bytes.Buffer
	if err := render(&buf); err != nil {
		return err
	}
	text := strings.TrimRight(buf.String(), "\n")
	if text == "" {
		return nil
	}
	for line := range strings.SplitSeq(text, "\n") {
		if line == "" {
			r.printf("\n")
			continue
		}
		r.printf("%s%s%s\n", indent, indent, line)
	}
	return nil
}

func (r *Reporter) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(r.w, format, args...)
	r.atBlank = false
	r.started = true
}

func (r *Reporter) blank() {
	if r.atBlank || !r.started {
		return
	}
	_, _ = io.WriteString(r.w, "\n")
	r.atBlank = true
}

func wrap(text string, width int) []string {
	var lines []string
	for para := range strings.SplitSeq(text, "\n") {
		lines = append(lines, wrapParagraph(para, width)...)
	}
	return lines
}

func wrapParagraph(para string, width int) []string {
	words := strings.Fields(para)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	line, n := words[0], utf8.RuneCountInString(words[0])
	for _, word := range words[1:] {
		w := utf8.RuneCountInString(word)
		if n+1+w > width {
			lines = append(lines, line)
			line, n = word, w
			continue
		}
		line, n = line+" "+word, n+1+w
	}
	return append(lines, line)
}
