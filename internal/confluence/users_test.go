package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newUserTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/user" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		q := r.URL.Query()
		user := User{Type: "known", Username: "k77319", UserKey: "1f58644b4d9ba951014dae08fed80006", DisplayName: "Solheim, Bent Andre"}
		if q.Get("key") == user.UserKey || q.Get("username") == user.Username {
			_ = json.NewEncoder(w).Encode(user)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"statusCode":404,"message":"No user found"}`))
	}))
}

func TestGetUserByKey(t *testing.T) {
	srv := newUserTestServer(t)
	defer srv.Close()
	c := NewClient(srv.URL, "secret-token", false)

	u, err := c.GetUserByKey("1f58644b4d9ba951014dae08fed80006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Username != "k77319" || u.DisplayName != "Solheim, Bent Andre" {
		t.Errorf("unexpected user: %+v", u)
	}
}

func TestGetUserByUsername(t *testing.T) {
	srv := newUserTestServer(t)
	defer srv.Close()
	c := NewClient(srv.URL, "secret-token", false)

	u, err := c.GetUserByUsername("k77319")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.UserKey != "1f58644b4d9ba951014dae08fed80006" {
		t.Errorf("unexpected user key: %q", u.UserKey)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	srv := newUserTestServer(t)
	defer srv.Close()
	c := NewClient(srv.URL, "secret-token", false)

	if _, err := c.GetUserByUsername("nobody"); err == nil {
		t.Fatal("expected error for unknown user")
	}
}
