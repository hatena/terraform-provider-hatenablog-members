package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func setup(t *testing.T) (*http.ServeMux, *httptest.Server, *Client) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	client := NewClient("test", "username", "apikey", "owner", "blog.example.com")

	serverURL, _ := url.Parse(server.URL)
	client.SetHatenablogHost(serverURL.Host)
	if serverURL.Scheme == "http" {
		client.SetInsecure(true)
	}

	return mux, server, client
}

func teardown(server *httptest.Server) {
	server.Close()
}

func parseJSON(t *testing.T, body io.Reader, v interface{}) {
	t.Helper()

	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("failed to parse JSON: %s", err)
	}
}

func assertRequest(t *testing.T, r *http.Request, method string) {
	t.Helper()

	if r.Method != method {
		t.Errorf("unexpected method: %s", r.Method)
	}
	if r.Header.Get("X-WSSE") == "" {
		t.Errorf("missing X-WSSE header")
	}
	if !strings.Contains(r.UserAgent(), "terraform-provider-hatenablog-members") {
		t.Errorf("unexpected User-Agent: %s", r.UserAgent())
	}
}

func TestNewClient(t *testing.T) {
	// can instantiate without error
	var _ *Client = NewClient("test", "username", "apikey", "owner", "blog.example.com")
}

func TestClient_SetHatenablogHost(t *testing.T) {
	client := NewClient("test", "username", "apikey", "owner", "blog.example.com")
	client.SetHatenablogHost("example.com")
	url := client.buildURL()

	if url.Host != "example.com" {
		t.Errorf("unexpected host: %s", url.Host)
	}
}

func TestClient_SetInsecure(t *testing.T) {
	client := NewClient("test", "username", "apikey", "owner", "blog.example.com")
	client.SetInsecure(true)
	url := client.buildURL()

	if url.Scheme != "http" {
		t.Errorf("unexpected scheme: %s", url.Scheme)
	}
}

func TestClient_ListMembers(t *testing.T) {
	mux, server, client := setup(t)
	defer teardown(server)

	mux.HandleFunc("/owner/blog.example.com/api/members", func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "GET")

		fmt.Fprint(w, `{"members":[{"username":"member","role":"admin"}]}`)
	})

	members, err := client.ListMembers()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(members) != 1 {
		t.Errorf("unexpected members: %v", members)
	}
	if members[0].Username != "member" {
		t.Errorf("unexpected username: %s", members[0].Username)
	}
	if members[0].Role != "admin" {
		t.Errorf("unexpected role: %s", members[0].Role)
	}
}

func TestCleint_AddMember(t *testing.T) {
	mux, server, client := setup(t)
	defer teardown(server)

	mux.HandleFunc("/owner/blog.example.com/api/members", func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "POST")

		var data any
		parseJSON(t, r.Body, &data)
		ok := reflect.DeepEqual(data, map[string]interface{}{
			"username": "member",
			"role":     "admin",
		})
		if !ok {
			t.Errorf("unexpected body: %v", data)
		}

		fmt.Fprint(w, `{"username":"member","role":"admin"}`)
	})

	member, err := client.AddMember("member", "admin")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if member.Username != "member" {
		t.Errorf("unexpected username: %s", member.Username)
	}
	if member.Role != "admin" {
		t.Errorf("unexpected role: %s", member.Role)
	}
}

func TestClient_DeleteMember(t *testing.T) {
	mux, server, client := setup(t)
	defer teardown(server)

	mux.HandleFunc("/owner/blog.example.com/api/members/member", func(w http.ResponseWriter, r *http.Request) {
		assertRequest(t, r, "DELETE")
	})

	if err := client.DeleteMember("member"); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
