package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsAdult(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	cases := map[string]bool{
		"2008-10-27": true,  // exactly 18 today
		"2008-10-28": false, // turns 18 tomorrow
		"1990-01-01": true,
		"2020-05-05": false,
		"":           false,
		"not-a-date": false,
		"10/27/2000": false,
	}
	for in, want := range cases {
		assert.Equal(t, want, IsAdult(in, now), in)
	}
}
