package cmd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestDescribeModelUsesRuntimeIntrospection(t *testing.T) {
	oldClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/end-user/graphql/org/acme/project/sales/db/maindb/model/users" {
				t.Fatalf("path = %s", req.URL.Path)
			}
			body, _ := io.ReadAll(req.Body)
			// describe uses __schema introspection
			if !bytes.Contains(body, []byte(`__schema`)) {
				t.Fatalf("expected __schema introspection query, got %s", string(body))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: io.NopCloser(bytes.NewBufferString(`{"data":{"__schema":{"types":[{"name":"users","kind":"OBJECT","fields":[{"name":"id","type":{"kind":"NON_NULL","name":null,"ofType":{"kind":"SCALAR","name":"ID","ofType":null}},"args":[]},{"name":"tags","type":{"kind":"LIST","name":null,"ofType":{"kind":"SCALAR","name":"String","ofType":null}},"args":[]}],"inputFields":null}]}}}`)),
			}, nil
		}),
	}
	defer func() { http.DefaultClient = oldClient }()

	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials.json")
	err := os.WriteFile(credPath, []byte(`{"server":"https://gateway.example.com","orgName":"acme","accessToken":"tok","expiresAt":"2099-01-01T00:00:00Z","currentProject":"sales"}`), 0o600)
	if err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	code := Execute(BuildInfo{}, []string{"describe", "sales.maindb.users", "--credentials", credPath}, stdout, stderr)
	if code != 0 {
		t.Fatalf("Execute() code = %d, stdout=%s", code, stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte(`"types"`)) {
		t.Fatalf("missing types payload: %s", stdout.String())
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
