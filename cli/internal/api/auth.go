package api

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims identifies the caller after token verification.
type Claims struct {
	Sub      string
	Username string
	Email    string
}

// Authenticator turns a bearer token into Claims.
type Authenticator interface {
	Verify(ctx context.Context, token string) (Claims, error)
}

// ErrUnauthenticated means the token is missing, malformed, or invalid.
var ErrUnauthenticated = errors.New("unauthenticated")

type ctxKey struct{}

// WithClaims stores claims on the context.
func WithClaims(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// ClaimsFrom returns the caller's claims, if authenticated.
func ClaimsFrom(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(ctxKey{}).(Claims)
	return c, ok
}

// bearerToken extracts the token from an Authorization header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if len(h) > 7 && strings.EqualFold(h[:7], "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

// CognitoVerifier validates Cognito-issued ID and access tokens against the
// user pool's JWKS.
type CognitoVerifier struct {
	Region     string
	PoolID     string
	ClientID   string
	HTTPClient *http.Client

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
}

func (v *CognitoVerifier) issuer() string {
	return fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", v.Region, v.PoolID)
}

// Verify implements Authenticator.
func (v *CognitoVerifier) Verify(ctx context.Context, token string) (Claims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		return v.key(ctx, kid)
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(v.issuer()), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return Claims{}, fmt.Errorf("%w: %v", ErrUnauthenticated, err)
	}
	m, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrUnauthenticated
	}
	use, _ := m["token_use"].(string)
	switch use {
	case "id":
		if aud, _ := m["aud"].(string); aud != v.ClientID {
			return Claims{}, fmt.Errorf("%w: wrong audience", ErrUnauthenticated)
		}
	case "access":
		if cid, _ := m["client_id"].(string); cid != v.ClientID {
			return Claims{}, fmt.Errorf("%w: wrong client", ErrUnauthenticated)
		}
	default:
		return Claims{}, fmt.Errorf("%w: unexpected token_use", ErrUnauthenticated)
	}
	c := Claims{}
	c.Sub, _ = m["sub"].(string)
	c.Email, _ = m["email"].(string)
	c.Username, _ = m["cognito:username"].(string)
	if c.Username == "" {
		c.Username, _ = m["username"].(string)
	}
	if c.Sub == "" {
		return Claims{}, ErrUnauthenticated
	}
	return c, nil
}

func (v *CognitoVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	// refetch at most once a minute so unknown kids cannot hammer Cognito
	if v.keys != nil && time.Since(v.fetched) < time.Minute {
		return nil, errors.New("unknown signing key")
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	k, ok := v.keys[kid]
	if !ok {
		return nil, errors.New("unknown signing key")
	}
	return k, nil
}

func (v *CognitoVerifier) refresh(ctx context.Context) error {
	client := v.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.issuer()+"/.well-known/jwks.json", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks: status %d", resp.StatusCode)
	}
	var doc struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	v.keys, v.fetched = keys, time.Now()
	return nil
}
