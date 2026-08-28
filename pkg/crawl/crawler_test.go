package crawl

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCrawlNative(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Test App</title></head>
<body>
    <a href="/login">Login</a>
    <a href="/dashboard">Dashboard</a>
    <a href="https://external.example.com">External</a>
    <script src="/static/bundle.js"></script>
    <form action="/api/v1/submit" method="POST"></form>
</body>
</html>`))
	})

	mux.HandleFunc("/static/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`
const apiUser = "/api/v2/users";
const apiAuth = "/api/v1/auth/token";
const logout = "/logout";
fetch("/api/v1/internal/config");
`))
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<a href="/admin/settings">Admin</a>`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	tmpDir := t.TempDir()
	reconDir := filepath.Join(tmpDir, "recon", "test_app")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := CrawlOptions{
		TargetURL:    server.URL,
		Depth:        2,
		MaxEndpoints: 15,
		Timeout:      5 * time.Second,
		OutputDir:    reconDir,
		PreferKatana: false,
	}

	res, err := CrawlTarget(ctx, opts)
	if err != nil {
		t.Fatalf("CrawlTarget failed: %v", err)
	}

	if res.TotalEndpointsFound == 0 {
		t.Fatal("Expected endpoints to be discovered, got 0")
	}

	if res.EngineUsed != "native_go" {
		t.Errorf("Expected engine native_go, got %s", res.EngineUsed)
	}

	// Verify endpoints discovered
	hasEndpointContaining := func(sub string) bool {
		for _, item := range res.TopEndpoints {
			if strings.Contains(item.Text, sub) {
				return true
			}
		}
		return false
	}

	if !hasEndpointContaining("/login") {
		t.Errorf("Expected /login to be discovered")
	}
	if !hasEndpointContaining("/api/v1/submit") {
		t.Errorf("Expected /api/v1/submit to be discovered")
	}
	if !hasEndpointContaining("/api/v2/users") {
		t.Errorf("Expected /api/v2/users from JS bundle to be discovered")
	}
	if !hasEndpointContaining("/api/v1/auth/token") {
		t.Errorf("Expected /api/v1/auth/token from JS bundle to be discovered")
	}

	// Verify katana.txt file was written to disk
	savedFile := filepath.Join(reconDir, "katana.txt")
	content, err := os.ReadFile(savedFile)
	if err != nil {
		t.Fatalf("Failed to read saved file %s: %v", savedFile, err)
	}

	if len(content) == 0 {
		t.Errorf("Expected katana.txt to have content")
	}
}

func TestCrawlNative_WithAuthHeadersAndCookies(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
    <a href="/public">Public</a>
    <a href="/protected/admin">Protected Admin</a>
</body>
</html>`))
	})

	mux.HandleFunc("/protected/admin", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		cookie := r.Header.Get("Cookie")
		if auth != "Bearer token-xyz" || !strings.Contains(cookie, "role=superadmin") {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<body>
    <a href="/api/v1/internal/secrets">Internal Secrets</a>
    <a href="/api/v1/users/export">Export Users</a>
</body>
</html>`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := CrawlOptions{
		TargetURL:    server.URL,
		Depth:        2,
		MaxEndpoints: 20,
		Timeout:      3 * time.Second,
		Headers: map[string]string{
			"Authorization": "Bearer token-xyz",
		},
		Cookies:      "role=superadmin",
		PreferKatana: false,
	}

	res, err := CrawlTarget(ctx, opts)
	if err != nil {
		t.Fatalf("CrawlTarget failed: %v", err)
	}

	hasEndpointContaining := func(sub string) bool {
		for _, item := range res.TopEndpoints {
			if strings.Contains(item.Text, sub) {
				return true
			}
		}
		return false
	}

	if !hasEndpointContaining("/protected/admin") {
		t.Errorf("Expected /protected/admin to be discovered")
	}
	if !hasEndpointContaining("/api/v1/internal/secrets") {
		t.Errorf("Expected authenticated endpoint /api/v1/internal/secrets to be discovered")
	}
	if !hasEndpointContaining("/api/v1/users/export") {
		t.Errorf("Expected authenticated endpoint /api/v1/users/export to be discovered")
	}
}

