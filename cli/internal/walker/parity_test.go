package walker

import (
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/aallbrig/beetree-cli/internal/model"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite web/src/testdata/walker-golden.json")

const (
	templatesPath = "../../../web/src/testdata/walker-templates.json"
	goldenPath    = "../../../web/src/testdata/walker-golden.json"
)

type goldenStep struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Depth  int    `json:"depth"`
	Event  string `json:"event"`
	Status string `json:"status"`
}

type goldenCase struct {
	Template string       `json:"template"`
	Policy   string       `json:"policy"`
	Result   string       `json:"result"`
	Trace    []goldenStep `json:"trace"`
}

func answer(policy string, i int) model.Status {
	switch policy {
	case "failure":
		return model.StatusFailure
	case "running":
		return model.StatusRunning
	case "cycle":
		return []model.Status{model.StatusSuccess, model.StatusFailure, model.StatusRunning}[i%3]
	}
	return model.StatusSuccess
}

// TestParityGolden runs the Go walker over the web app's template trees. The
// TypeScript walker's test compares against the same file, so the two cannot
// drift apart. Regenerate with: go test ./internal/walker -run Parity -update-golden
func TestParityGolden(t *testing.T) {
	raw, err := os.ReadFile(templatesPath)
	require.NoError(t, err)
	var templates map[string]model.NodeSpec
	require.NoError(t, json.Unmarshal(raw, &templates))

	var cases []goldenCase
	// The TS fixture lists templates in SAMPLES order; keep that order via the file's key order.
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	_, _ = dec.Token()
	var order []string
	for dec.More() {
		k, _ := dec.Token()
		order = append(order, k.(string))
		var skip json.RawMessage
		require.NoError(t, dec.Decode(&skip))
	}
	for _, id := range order {
		root := templates[id]
		for _, policy := range []string{"success", "failure", "running", "cycle"} {
			w := NewSimWalker(&root)
			w.Step()
			for i := 0; w.State == SimWaitingForInput; i++ {
				w.Resolve(answer(policy, i))
			}
			c := goldenCase{Template: id, Policy: policy, Result: strings.ToLower(w.Result.String())}
			for _, s := range w.Trace {
				st := goldenStep{Name: s.NodeName, Type: s.NodeType, Depth: s.Depth, Event: s.Event}
				if s.Event == "resolve" {
					st.Status = strings.ToLower(s.Status.String())
				}
				c.Trace = append(c.Trace, st)
			}
			cases = append(cases, c)
		}
	}
	out, err := json.MarshalIndent(cases, "", "  ")
	require.NoError(t, err)
	out = append(out, '\n')
	if *updateGolden {
		require.NoError(t, os.WriteFile(goldenPath, out, 0o644))
	}
	want, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(out))
}
