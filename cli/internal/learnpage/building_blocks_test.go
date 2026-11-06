// Package learnpage tests the code samples on the Learn > Building blocks page.
package learnpage

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pagePath = "../../../website/content/learn/building-blocks.md"

type status int

const (
	success status = iota
	failure
	running
)

type node interface{ tick() status }

type fn func() status

func (f fn) tick() status { return f() }

// The ports below mirror the agnostic pseudocode on the page line for line.

type sequence []node

func (s sequence) tick() status {
	for _, c := range s {
		switch c.tick() {
		case running:
			return running
		case failure:
			return failure
		}
	}
	return success
}

type fallback []node

func (f fallback) tick() status {
	for _, c := range f {
		switch c.tick() {
		case running:
			return running
		case success:
			return success
		}
	}
	return failure
}

type parallel struct {
	children []node
	m        int
}

func (p parallel) tick() status {
	successes, failures := 0, 0
	for _, c := range p.children {
		switch c.tick() {
		case success:
			successes++
		case failure:
			failures++
		}
	}
	if successes >= p.m {
		return success
	}
	if failures > len(p.children)-p.m {
		return failure
	}
	return running
}

type inverter struct{ child node }

func (i inverter) tick() status {
	switch i.child.tick() {
	case success:
		return failure
	case failure:
		return success
	}
	return running
}

type retry struct {
	child       node
	maxFailures int
	failures    int
}

func (r *retry) tick() status {
	s := r.child.tick()
	if s == running {
		return running
	}
	if s == failure {
		r.failures++
		if r.failures < r.maxFailures {
			return running
		}
	}
	r.failures = 0
	return s
}

func leaf(s status) node { return fn(func() status { return s }) }

func leaves(ss ...status) []node {
	out := make([]node, len(ss))
	for i, s := range ss {
		out[i] = leaf(s)
	}
	return out
}

var all = []status{success, failure, running}

func TestSequenceAndFallbackStatusTable(t *testing.T) {
	for _, a := range all {
		for _, b := range all {
			wantSeq := success
			switch {
			case a == running, a == success && b == running:
				wantSeq = running
			case a == failure, a == success && b == failure:
				wantSeq = failure
			}
			assert.Equal(t, wantSeq, sequence(leaves(a, b)).tick(), "sequence %v %v", a, b)

			wantFb := failure
			switch {
			case a == running, a == failure && b == running:
				wantFb = running
			case a == success, a == failure && b == success:
				wantFb = success
			}
			assert.Equal(t, wantFb, fallback(leaves(a, b)).tick(), "fallback %v %v", a, b)
		}
	}
}

func TestParallelStatusTable(t *testing.T) {
	for _, a := range all {
		for _, b := range all {
			for _, c := range all {
				for m := 1; m <= 3; m++ {
					var s, f int
					for _, x := range []status{a, b, c} {
						switch x {
						case success:
							s++
						case failure:
							f++
						}
					}
					want := running
					if s >= m {
						want = success
					} else if f > 3-m {
						want = failure
					}
					got := parallel{leaves(a, b, c), m}.tick()
					assert.Equal(t, want, got, "parallel %v %v %v m=%d", a, b, c, m)
				}
			}
		}
	}
	// Every child is ticked every time, even after the outcome is known.
	ticks := 0
	counter := fn(func() status { ticks++; return success })
	parallel{[]node{leaf(failure), counter}, 1}.tick()
	assert.Equal(t, 1, ticks)
}

func TestInverter(t *testing.T) {
	assert.Equal(t, failure, inverter{leaf(success)}.tick())
	assert.Equal(t, success, inverter{leaf(failure)}.tick())
	assert.Equal(t, running, inverter{leaf(running)}.tick())
}

func TestRetryGivesUpThenWorksAgainAndResetsOnSuccess(t *testing.T) {
	next := failure
	r := &retry{child: fn(func() status { return next }), maxFailures: 3}

	run := func() []status {
		return []status{r.tick(), r.tick(), r.tick()}
	}
	assert.Equal(t, []status{running, running, failure}, run(), "gives up after maxFailures")
	assert.Equal(t, []status{running, running, failure}, run(), "counter reset: works again")

	assert.Equal(t, running, r.tick())
	next = success
	assert.Equal(t, success, r.tick(), "success passes through")
	next = failure
	assert.Equal(t, []status{running, running, failure}, run(), "success reset the counter")
}

func page(t *testing.T) string {
	b, err := os.ReadFile(pagePath)
	require.NoError(t, err)
	return string(b)
}

var tabRe = regexp.MustCompile(`(?s)\{\{< tab id="(\w+)" label="[^"]*" lang="[^"]*" >\}\}\n(.*?)\{\{< /tab >\}\}`)

func tabs(t *testing.T) (blocks []map[string]string) {
	p := page(t)
	for _, group := range strings.Split(p, "{{< codetabs >}}")[1:] {
		group = strings.SplitN(group, "{{< /codetabs >}}", 2)[0]
		m := map[string]string{}
		for _, x := range tabRe.FindAllStringSubmatch(group, -1) {
			m[x[1]] = x[2]
		}
		blocks = append(blocks, m)
	}
	require.NotEmpty(t, blocks)
	return blocks
}

func TestPseudocodeMatchesPorts(t *testing.T) {
	var agnostic strings.Builder
	for _, b := range tabs(t) {
		agnostic.WriteString(b["agnostic"])
	}
	src := agnostic.String()
	for _, line := range []string{
		"if childStatus == Running: return Running",
		"else if childStatus == Failure: return Failure",
		"else if childStatus == Success: return Success",
		"if successes >= M: return Success",
		"else if failures > N - M: return Failure",
		"if failures < maxFailures: return Running",
	} {
		assert.Contains(t, src, line)
	}
}

func TestRetryPseudocodeResetsCounterWhenFinished(t *testing.T) {
	var agnostic strings.Builder
	for _, b := range tabs(t) {
		agnostic.WriteString(b["agnostic"])
	}
	assert.Regexp(t, `failures = 0\s+// finished`, agnostic.String())
}

func TestPseudocodeUsesNoNumericIndices(t *testing.T) {
	idx := regexp.MustCompile(`\[\s*\d+\s*\]|\b(first|second)\b.*\[|children\[`)
	for _, b := range tabs(t) {
		assert.False(t, idx.MatchString(b["agnostic"]), "numeric index in:\n%s", b["agnostic"])
	}
}

func TestEveryTabNamesTheSameNodes(t *testing.T) {
	names := regexp.MustCompile(`Action|Condition|Sequence|Fallback|Parallel|Inverter|Retry`)
	set := func(s string) string {
		seen := map[string]bool{}
		for _, n := range names.FindAllString(s, -1) {
			seen[n] = true
		}
		var out []string
		for n := range seen {
			out = append(out, n)
		}
		sort.Strings(out)
		return strings.Join(out, ",")
	}
	for i, b := range tabs(t) {
		want := set(b["agnostic"])
		for id, code := range b {
			assert.Equal(t, want, set(code), "block %d tab %s node names differ from agnostic", i, id)
		}
	}
}
