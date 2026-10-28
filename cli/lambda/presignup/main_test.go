package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func event(birthdate string) events.CognitoEventUserPoolsPreSignup {
	ev := events.CognitoEventUserPoolsPreSignup{}
	ev.Request.UserAttributes = map[string]string{"email": "a@b.c"}
	if birthdate != "" {
		ev.Request.UserAttributes["birthdate"] = birthdate
	}
	return ev
}

func TestHandle(t *testing.T) {
	_, err := handle(context.Background(), event("1990-05-05"))
	assert.NoError(t, err)

	for name, bd := range map[string]string{"minor": "2020-01-01", "missing": "", "garbage": "yesterday"} {
		_, err := handle(context.Background(), event(bd))
		assert.ErrorIs(t, err, errUnderage, name)
	}
}
