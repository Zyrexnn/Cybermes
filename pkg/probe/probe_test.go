package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProbeNative(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx/1.24.0")
		w.Header().Set("X-Powered-By", "Next.js")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.SetCookie(w, &http.Cookie{
			Name:  "laravel_session",
			Value: "xyz123mocktoken",
		})
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Cybermes Dashboard</title>
</head>
<body>
    <div id="__NEXT_DATA__">{}</div>
    <div data-reactroot="">Hello World</div>
</body>
</html>`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := ProbeOptions{
		TargetURL:   server.URL,
		Timeout:     3 * time.Second,
		PreferHttpx: false,
	}

	res, err := ProbeTarget(ctx, opts)
	if err != nil {
		t.Fatalf("ProbeTarget failed: %v", err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	if res.Title != "Cybermes Dashboard" {
		t.Errorf("Expected title 'Cybermes Dashboard', got '%s'", res.Title)
	}

	if res.WebServer != "nginx/1.24.0" {
		t.Errorf("Expected web server 'nginx/1.24.0', got '%s'", res.WebServer)
	}

	if res.EngineUsed != "native_go" {
		t.Errorf("Expected engine 'native_go', got '%s'", res.EngineUsed)
	}

	// Verify tech detections
	hasTech := func(target string) bool {
		for _, tech := range res.Technologies {
			if strings.EqualFold(tech, target) {
				return true
			}
		}
		return false
	}

	if !hasTech("Next.js") {
		t.Errorf("Expected Next.js to be detected in technologies: %v", res.Technologies)
	}
	if !hasTech("React") {
		t.Errorf("Expected React to be detected in technologies: %v", res.Technologies)
	}
	if !hasTech("Nginx") {
		t.Errorf("Expected Nginx to be detected in technologies: %v", res.Technologies)
	}
	if !hasTech("Laravel") {
		t.Errorf("Expected Laravel cookie to be detected in technologies: %v", res.Technologies)
	}
}

func TestExtractTitle(t *testing.T) {
	html1 := "<html><head><title>  My Secure App   </title></head></html>"
	if title := ExtractTitle(html1); title != "My Secure App" {
		t.Errorf("Expected 'My Secure App', got '%s'", title)
	}

	html2 := "<html><head><TITLE>Portal - Home</TITLE></head></html>"
	if title := ExtractTitle(html2); title != "Portal - Home" {
		t.Errorf("Expected 'Portal - Home', got '%s'", title)
	}

	html3 := "<html><body>No title here</body></html>"
	if title := ExtractTitle(html3); title != "" {
		t.Errorf("Expected empty title, got '%s'", title)
	}
}

func TestProbeNative_WithAuthHeadersAndCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		cookieHeader := r.Header.Get("Cookie")
		customHeader := r.Header.Get("X-Custom-Auth")

		if authHeader != "Bearer secret-token-123" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}
		if !strings.Contains(cookieHeader, "session_id=mock-session-456") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("Forbidden"))
			return
		}
		if customHeader != "CybermesSpecial" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Missing custom header"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"authenticated": true, "user": "admin"}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := ProbeOptions{
		TargetURL: server.URL,
		Timeout:   3 * time.Second,
		Headers: map[string]string{
			"Authorization": "Bearer secret-token-123",
			"X-Custom-Auth": "CybermesSpecial",
		},
		Cookies:     "session_id=mock-session-456",
		PreferHttpx: false,
	}

	res, err := ProbeTarget(ctx, opts)
	if err != nil {
		t.Fatalf("ProbeTarget failed: %v", err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("Expected status 200, got %d", res.StatusCode)
	}
}

