// Package lambdahttp adapts an http.Handler to API Gateway HTTP API (payload v2).
package lambdahttp

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Adapt returns a Lambda handler that serves requests with h.
func Adapt(h http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return func(ctx context.Context, ev events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		body := ev.Body
		if ev.IsBase64Encoded {
			dec, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusBadRequest, Body: `{"error":{"code":"bad_request","message":"invalid body encoding"}}`}, nil
			}
			body = string(dec)
		}
		target := ev.RawPath
		if ev.RawQueryString != "" {
			target += "?" + ev.RawQueryString
		}
		req, err := http.NewRequestWithContext(ctx, ev.RequestContext.HTTP.Method, target, strings.NewReader(body))
		if err != nil {
			return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusBadRequest}, nil
		}
		for k, v := range ev.Headers {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		headers := map[string]string{}
		for k, v := range rec.Header() {
			headers[k] = strings.Join(v, ",")
		}
		return events.APIGatewayV2HTTPResponse{StatusCode: rec.Code, Headers: headers, Body: rec.Body.String()}, nil
	}
}
