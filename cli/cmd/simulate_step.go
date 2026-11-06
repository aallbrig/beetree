package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/aallbrig/beetree-cli/internal/renderer"
	"github.com/aallbrig/beetree-cli/internal/tui"
)

// parseAnswer maps s/f/r (or the full status name) to a Status.
func parseAnswer(s string) (model.Status, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "s", "success":
		return model.StatusSuccess, true
	case "f", "failure":
		return model.StatusFailure, true
	case "r", "running":
		return model.StatusRunning, true
	}
	return 0, false
}

// runStepThrough drives the TUI's SimWalker from line-based input so the CLI
// and the TUI resolve composites identically. Each pending leaf prompts for
// s(uccess), f(ailure) or r(unning); the trace prints at the end.
func runStepThrough(root *model.NodeSpec, in io.Reader, out io.Writer) error {
	w := tui.NewSimWalker(root)
	w.Step()
	sc := bufio.NewScanner(in)
	for w.State == tui.SimWaitingForInput {
		fmt.Fprintf(out, "%s %s [s]uccess / [f]ailure / [r]unning: ", renderer.TypeSigil(w.CurrentNode.Type), w.CurrentNode.Name)
		if !sc.Scan() {
			fmt.Fprintln(out)
			return fmt.Errorf("input ended before the simulation finished")
		}
		status, ok := parseAnswer(sc.Text())
		if !ok {
			fmt.Fprintln(out, "  answer s, f or r")
			continue
		}
		w.Resolve(status)
	}
	fmt.Fprintf(out, "\nSimulation: %s -> %s\n%s\n", root.Name, strings.ToUpper(w.Result.String()), strings.Repeat("-", 50))
	for _, st := range w.Trace {
		suffix := ""
		if st.Event == "resolve" {
			suffix = " -> " + strings.ToUpper(st.Status.String())
		}
		fmt.Fprintf(out, "%s%s %s %s%s\n", strings.Repeat("  ", st.Depth), st.Event, renderer.TypeSigil(st.NodeType), st.NodeName, suffix)
	}
	return nil
}
