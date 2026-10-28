// Command presignup is the Cognito pre-sign-up trigger. It rejects sign-ups
// that do not carry a birthdate showing the user is at least 18, so the age
// gate cannot be bypassed by calling Cognito directly.
package main

import (
	"context"
	"errors"
	"time"

	"github.com/aallbrig/beetree-cli/internal/api"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var errUnderage = errors.New("You must be 18 or older to create an account.")

func handle(_ context.Context, ev events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	if !api.IsAdult(ev.Request.UserAttributes["birthdate"], time.Now()) {
		return ev, errUnderage
	}
	return ev, nil
}

func main() { lambda.Start(handle) }
