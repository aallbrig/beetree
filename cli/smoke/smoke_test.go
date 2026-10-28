//go:build smoke

// Package smoke exercises the deployed stack end to end. Run with `task smoke`.
// It creates one throwaway Cognito user and deletes everything it created.
package smoke

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const spec = `version: "1.0"
metadata:
  name: smoke
tree:
  type: sequence
  name: root
  children:
    - {type: condition, name: has_target}
    - {type: action, name: chase}
`

func aws(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("aws", append(args, "--region", "us-east-1", "--output", "json")...).CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func stackOutput(t *testing.T, key string) string {
	t.Helper()
	out := aws(t, "cloudformation", "describe-stacks", "--stack-name", "beetreecraft", "--query",
		fmt.Sprintf("Stacks[0].Outputs[?OutputKey=='%s'].OutputValue|[0]", key))
	var v string
	require.NoError(t, json.Unmarshal([]byte(out), &v))
	return v
}

func cognito(t *testing.T, clientID, target string, body map[string]any) (int, map[string]any) {
	t.Helper()
	body["ClientId"] = clientID
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://cognito-idp.us-east-1.amazonaws.com/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", "AWSCognitoIdentityProviderService."+target)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp.StatusCode, m
}

type client struct {
	t     *testing.T
	base  string
	token string
}

func (c client) do(method, path string, body any) (int, map[string]any) {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.base+path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)
	defer func() { _ = resp.Body.Close() }()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp.StatusCode, m
}

func TestLiveStack(t *testing.T) {
	apiURL := stackOutput(t, "ApiUrl")
	poolID := stackOutput(t, "UserPoolId")
	clientID := stackOutput(t, "UserPoolClientId")
	table := stackOutput(t, "TableName")
	anon := client{t: t, base: apiURL}

	// 1. The age gate holds even when the web form is bypassed.
	name := fmt.Sprintf("smoke%d", time.Now().Unix())
	code, resp := cognito(t, clientID, "SignUp", map[string]any{
		"Username": name + "kid", "Password": "Sup3rSecretPass",
		"UserAttributes": []map[string]string{{"Name": "email", "Value": name + "kid@example.com"}, {"Name": "birthdate", "Value": "2015-06-01"}},
	})
	assert.NotEqual(t, 200, code, "an 11-year-old must not be able to sign up")
	assert.Contains(t, fmt.Sprint(resp["message"]), "18 or older")
	code, _ = cognito(t, clientID, "SignUp", map[string]any{
		"Username": name + "nodob", "Password": "Sup3rSecretPass",
		"UserAttributes": []map[string]string{{"Name": "email", "Value": name + "nodob@example.com"}},
	})
	assert.NotEqual(t, 200, code, "sign-up without a birthdate must be rejected")

	// 2. A confirmed adult user (admin path skips the email step, not the trigger).
	pass := "Sup3rSecretPass1"
	aws(t, "cognito-idp", "admin-create-user", "--user-pool-id", poolID, "--username", name, "--message-action", "SUPPRESS",
		"--user-attributes", "Name=email,Value="+name+"@example.com", "Name=email_verified,Value=true", "Name=birthdate,Value=1990-01-01")
	t.Cleanup(func() {
		_ = exec.Command("aws", "cognito-idp", "admin-delete-user", "--user-pool-id", poolID, "--username", name, "--region", "us-east-1").Run()
	})
	aws(t, "cognito-idp", "admin-set-user-password", "--user-pool-id", poolID, "--username", name, "--password", pass, "--permanent")

	code, auth := cognito(t, clientID, "InitiateAuth", map[string]any{"AuthFlow": "USER_PASSWORD_AUTH", "AuthParameters": map[string]string{"USERNAME": name, "PASSWORD": pass}})
	require.Equal(t, 200, code, fmt.Sprint(auth))
	token := auth["AuthenticationResult"].(map[string]any)["IdToken"].(string)
	me := client{t: t, base: apiURL, token: token}

	code, bad := cognito(t, clientID, "InitiateAuth", map[string]any{"AuthFlow": "USER_PASSWORD_AUTH", "AuthParameters": map[string]string{"USERNAME": name, "PASSWORD": "wrong-password-1"}})
	assert.NotEqual(t, 200, code, fmt.Sprint(bad))

	// 3. Authenticated REST flows.
	code, _ = anon.do("GET", "/v1/me", nil)
	assert.Equal(t, 401, code)
	code, prof := me.do("GET", "/v1/me", nil)
	require.Equal(t, 200, code, fmt.Sprint(prof))
	userID := prof["userId"].(string)
	assert.Equal(t, name, prof["displayName"])

	code, tree := me.do("POST", "/v1/trees", map[string]any{"spec": spec, "visibility": "public", "name": "Smoke " + name, "tags": []string{"smoke"}})
	require.Equal(t, 201, code, fmt.Sprint(tree))
	treeID := tree["id"].(string)
	t.Cleanup(func() { _, _ = me.do("DELETE", "/v1/trees/"+treeID, nil) })

	code, list := anon.do("GET", "/v1/trees?q=smoke+"+name, nil)
	assert.Equal(t, 200, code)
	assert.Len(t, list["items"], 1, "public tree is discoverable anonymously")

	code, like := me.do("PUT", "/v1/trees/"+treeID+"/like", nil)
	assert.Equal(t, 200, code)
	assert.EqualValues(t, 1, like["likeCount"])
	code, cm := me.do("POST", "/v1/trees/"+treeID+"/comments", map[string]string{"body": "smoke test comment"})
	assert.Equal(t, 201, code, fmt.Sprint(cm))
	_, got := anon.do("GET", "/v1/trees/"+treeID, nil)
	assert.EqualValues(t, 1, got["likeCount"])
	assert.EqualValues(t, 1, got["commentCount"])

	_, stats := me.do("GET", "/v1/me/stats", nil)
	assert.EqualValues(t, 1, stats["trees"])
	assert.EqualValues(t, 1, stats["likes"])

	code, thread := me.do("POST", "/v1/forum/threads", map[string]string{"title": "smoke thread", "body": "hello"})
	require.Equal(t, 201, code)
	threadID := thread["id"].(string)
	t.Cleanup(func() { purge(t, table, "THREAD#"+threadID, "USER#"+userID) })
	code, _ = me.do("POST", "/v1/forum/threads/"+threadID+"/posts", map[string]string{"body": "reply"})
	assert.Equal(t, 201, code)

	code, scene := me.do("POST", "/v1/scenes", map[string]any{"name": "smoke scene", "agents": []map[string]any{{"name": "a", "model": "character-a", "treeId": treeID, "x": 1, "z": 2}}})
	require.Equal(t, 201, code, fmt.Sprint(scene))
	sceneID := scene["id"].(string)
	t.Cleanup(func() { _, _ = me.do("DELETE", "/v1/scenes/"+sceneID, nil) })

	// 4. MCP over the same auth.
	call := func(c client, tool string, args map[string]any) (int, map[string]any) {
		return c.do("POST", "/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": tool, "arguments": args}})
	}
	code, rpc := call(me, "validate_tree", map[string]any{"spec": spec})
	assert.Equal(t, 200, code)
	assert.Contains(t, fmt.Sprint(rpc["result"]), "valid")
	_, rpc = call(me, "list_my_trees", nil)
	assert.Contains(t, fmt.Sprint(rpc["result"]), treeID)
	_, rpc = call(anon, "list_my_trees", nil)
	assert.Contains(t, fmt.Sprint(rpc["result"]), "signed-in")

	// 5. Cleanup is verified, not assumed.
	code, _ = me.do("DELETE", "/v1/trees/"+treeID, nil)
	assert.Equal(t, 204, code)
	code, _ = anon.do("GET", "/v1/trees/"+treeID, nil)
	assert.Equal(t, 404, code)
	if v := os.Getenv("SMOKE_VERBOSE"); v != "" {
		t.Logf("user %s tree %s thread %s", name, treeID, threadID)
	}
}

// purge removes every item in the given partitions (best effort; test data only).
func purge(t *testing.T, table string, pks ...string) {
	t.Helper()
	for _, pk := range pks {
		out, err := exec.Command("aws", "dynamodb", "query", "--table-name", table, "--region", "us-east-1",
			"--key-condition-expression", "pk = :p", "--expression-attribute-values", fmt.Sprintf(`{":p":{"S":%q}}`, pk),
			"--projection-expression", "pk, sk", "--output", "json").Output()
		if err != nil {
			continue
		}
		var q struct {
			Items []map[string]map[string]string `json:"Items"`
		}
		if json.Unmarshal(out, &q) != nil {
			continue
		}
		for _, it := range q.Items {
			key, _ := json.Marshal(it)
			_ = exec.Command("aws", "dynamodb", "delete-item", "--table-name", table, "--region", "us-east-1", "--key", string(key)).Run()
		}
	}
}
