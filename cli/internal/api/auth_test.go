package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rewriteTransport struct{ target string }

func (rt rewriteTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.URL.Scheme = "http"
	r2.URL.Host = rt.target
	return http.DefaultTransport.RoundTrip(r2)
}

func newTestVerifier(t *testing.T) (*CognitoVerifier, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": "k1",
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(jwks.Close)
	v := &CognitoVerifier{
		Region: "us-east-1", PoolID: "pool", ClientID: "client",
		HTTPClient: &http.Client{Transport: rewriteTransport{target: jwks.Listener.Addr().String()}},
	}
	return v, key
}

func sign(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "k1"
	s, err := tok.SignedString(key)
	require.NoError(t, err)
	return s
}

func TestCognitoVerifier(t *testing.T) {
	v, key := newTestVerifier(t)
	iss := "https://cognito-idp.us-east-1.amazonaws.com/pool"
	exp := time.Now().Add(time.Hour).Unix()

	t.Run("id token", func(t *testing.T) {
		c, err := v.Verify(context.Background(), sign(t, key, jwt.MapClaims{
			"iss": iss, "exp": exp, "token_use": "id", "aud": "client",
			"sub": "u1", "email": "a@b.c", "cognito:username": "alice"}))
		require.NoError(t, err)
		assert.Equal(t, Claims{Sub: "u1", Username: "alice", Email: "a@b.c"}, c)
	})
	t.Run("access token", func(t *testing.T) {
		c, err := v.Verify(context.Background(), sign(t, key, jwt.MapClaims{
			"iss": iss, "exp": exp, "token_use": "access", "client_id": "client", "sub": "u1", "username": "alice"}))
		require.NoError(t, err)
		assert.Equal(t, "alice", c.Username)
	})
	t.Run("rejects", func(t *testing.T) {
		bad := map[string]jwt.MapClaims{
			"wrong issuer":   {"iss": "https://evil", "exp": exp, "token_use": "id", "aud": "client", "sub": "u"},
			"expired":        {"iss": iss, "exp": time.Now().Add(-time.Hour).Unix(), "token_use": "id", "aud": "client", "sub": "u"},
			"wrong audience": {"iss": iss, "exp": exp, "token_use": "id", "aud": "other", "sub": "u"},
			"wrong client":   {"iss": iss, "exp": exp, "token_use": "access", "client_id": "other", "sub": "u"},
			"no token_use":   {"iss": iss, "exp": exp, "sub": "u"},
			"no sub":         {"iss": iss, "exp": exp, "token_use": "id", "aud": "client"},
		}
		for name, claims := range bad {
			_, err := v.Verify(context.Background(), sign(t, key, claims))
			assert.ErrorIs(t, err, ErrUnauthenticated, name)
		}
	})
	t.Run("rejects forged signature", func(t *testing.T) {
		other, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)
		_, err = v.Verify(context.Background(), sign(t, other, jwt.MapClaims{
			"iss": iss, "exp": exp, "token_use": "id", "aud": "client", "sub": "u"}))
		assert.ErrorIs(t, err, ErrUnauthenticated)
	})
	t.Run("rejects garbage", func(t *testing.T) {
		_, err := v.Verify(context.Background(), "not.a.jwt")
		assert.ErrorIs(t, err, ErrUnauthenticated)
	})
}
