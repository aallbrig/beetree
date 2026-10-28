package lambdahttp

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdaptRoundTrip(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Seen", r.Header.Get("Authorization")+"|"+r.URL.Query().Get("q"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(r.Method + " " + r.URL.Path + " " + string(b)))
	})
	ev := events.APIGatewayV2HTTPRequest{
		RawPath: "/v1/x", RawQueryString: "q=1",
		Headers: map[string]string{"authorization": "Bearer t"},
		Body:    base64.StdEncoding.EncodeToString([]byte("hello")), IsBase64Encoded: true,
	}
	ev.RequestContext.HTTP.Method = "POST"
	res, err := Adapt(h)(context.Background(), ev)
	require.NoError(t, err)
	assert.Equal(t, 201, res.StatusCode)
	assert.Equal(t, "POST /v1/x hello", res.Body)
	assert.Equal(t, "Bearer t|1", res.Headers["X-Seen"])
}

func TestAdaptBadBase64(t *testing.T) {
	ev := events.APIGatewayV2HTTPRequest{Body: "!!!", IsBase64Encoded: true}
	ev.RequestContext.HTTP.Method = "POST"
	res, err := Adapt(http.NotFoundHandler())(context.Background(), ev)
	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
}
