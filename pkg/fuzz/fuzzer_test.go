package fuzz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNativeFuzzer_Discovery(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})

	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer fuzz-test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Admin Area"))
	})

	mux.HandleFunc("/redirect-target", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	mux.HandleFunc("/secret-forbidden", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := FuzzOptions{
		TargetURL:   server.URL,
		RateLimit:   20,
		Timeout:     3 * time.Second,
		StatusCodes: []int{200, 302, 401, 403},
		Headers: map[string]string{
			"Authorization": "Bearer fuzz-test-token",
		},
		PreferFfuf: false,
	}

	res, err := FuzzEndpoints(ctx, opts)
	if err != nil {
		t.Fatalf("FuzzEndpoints failed: %v", err)
	}

	if res.EngineUsed != "native_go" {
		t.Errorf("Expected engine 'native_go', got '%s'", res.EngineUsed)
	}

	if res.TotalProbed == 0 {
		t.Fatal("Expected >0 requests probed")
	}

	hasMatch := func(path string, expectedStatus int) bool {
		for _, m := range res.Matches {
			if strings.TrimPrefix(m.Path, "/") == strings.TrimPrefix(path, "/") && m.StatusCode == expectedStatus {
				return true
			}
		}
		return false
	}

	if !hasMatch("admin", 200) {
		t.Errorf("Expected /admin to be matched with 200 OK")
	}
}

func TestResolveWordlist(t *testing.T) {
	_, words, err := ResolveWordlist("", "")
	if err != nil {
		t.Fatalf("ResolveWordlist returned error: %v", err)
	}
	if len(words) == 0 {
		t.Errorf("Expected default words list, got empty")
	}
}
