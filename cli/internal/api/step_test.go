package api

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulateStepFlow(t *testing.T) {
	h := newTestServer(t)
	step := func(answers ...string) map[string]any {
		r := do(h, "POST", "/v1/simulate/step", "", map[string]any{"spec": goodSpec, "answers": answers})
		require.Equal(t, 200, r.Code, string(r.Body))
		return r.JSON(t)
	}

	j := step()
	assert.Equal(t, false, j["done"])
	assert.Equal(t, map[string]any{"node": "has_waypoint", "type": "condition"}, j["pending"])

	j = step("SUCCESS")
	assert.Equal(t, false, j["done"])
	assert.Equal(t, "move", j["pending"].(map[string]any)["node"])

	j = step("success", "FAILURE")
	assert.Equal(t, true, j["done"])
	assert.Equal(t, "FAILURE", j["status"])
	assert.Nil(t, j["pending"])
	assert.Contains(t, j["trace"], "resolve root -> FAILURE")

	// A failing first leaf short-circuits the sequence: one answer finishes it.
	j = step("FAILURE")
	assert.Equal(t, true, j["done"])
}

func TestSimulateStepErrors(t *testing.T) {
	h := newTestServer(t)
	post := func(body map[string]any) int { return do(h, "POST", "/v1/simulate/step", "", body).Code }
	assert.Equal(t, 422, post(map[string]any{"spec": goodSpec, "answers": []string{"maybe"}}))
	assert.Equal(t, 422, post(map[string]any{"spec": goodSpec, "answers": []string{"FAILURE", "SUCCESS"}}), "answer after the end")
	assert.Equal(t, 422, post(map[string]any{"spec": ":::"}))
}

func TestSimulateStepMatchesFullSimulation(t *testing.T) {
	for _, ov := range []map[string]string{{}, {"has_waypoint": "FAILURE"}, {"move": "RUNNING"}} {
		full, err := simulateYAML(goodSpec, ov)
		require.NoError(t, err)
		var answers []string
		var res *StepResult
		for {
			res, err = stepYAML(goodSpec, answers)
			require.NoError(t, err)
			if res.Done {
				break
			}
			a := "SUCCESS"
			if o, ok := ov[res.Pending.Node]; ok {
				a = o
			}
			answers = append(answers, a)
		}
		assert.Equal(t, full.Status, res.Status, "%v", ov)
	}
}

func TestMCPStepTools(t *testing.T) {
	h := newTestServer(t)
	call := func(name string, args map[string]any) map[string]any {
		_, r := rpc(h, "", 1, "tools/call", map[string]any{"name": name, "arguments": args})
		text, isErr := toolResult(t, r)
		require.False(t, isErr, text)
		var out map[string]any
		require.NoError(t, json.Unmarshal([]byte(text), &out))
		return out
	}
	j := call("simulate_start", map[string]any{"spec": goodSpec})
	assert.Equal(t, "has_waypoint", j["pending"].(map[string]any)["node"])
	j = call("simulate_answer", map[string]any{"spec": goodSpec, "status": "SUCCESS"})
	assert.Equal(t, "move", j["pending"].(map[string]any)["node"])
	j = call("simulate_answer", map[string]any{"spec": goodSpec, "answers": []string{"SUCCESS"}, "status": "RUNNING"})
	assert.Equal(t, true, j["done"])
	assert.Equal(t, "RUNNING", j["status"])
}

// The docs conversation must match what the server really returns.
func TestStepDocsExample(t *testing.T) {
	doc, err := os.ReadFile("../../../website/content/docs/api/_index.md")
	require.NoError(t, err)
	first, err := stepYAML(goodSpec, nil)
	require.NoError(t, err)
	last, err := stepYAML(goodSpec, []string{"SUCCESS", "FAILURE"})
	require.NoError(t, err)
	for _, want := range []string{
		`"pending":{"node":"` + first.Pending.Node + `","type":"` + first.Pending.Type + `"}`,
		`"status":"` + last.Status + `"`,
		"/v1/simulate/step", "simulate_start", "simulate_answer",
	} {
		assert.Contains(t, string(doc), want)
	}
}
