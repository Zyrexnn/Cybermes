package nmap

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestExtractHost(t *testing.T) {
	valid := map[string]struct {
		host string
		port int
	}{
		"example.com":                {host: "example.com"},
		"example.com/api/v1":         {host: "example.com"},
		"https://api.example.com/v1": {host: "api.example.com"},
		"http://127.0.0.1:8888/x":    {host: "127.0.0.1", port: 8888},
		"127.0.0.1:8080":             {host: "127.0.0.1", port: 8080},
		"10.0.0.5":                   {host: "10.0.0.5"},
		"::1":                        {host: "::1"},
		"[::1]:443":                  {host: "::1", port: 443},
	}
	for in, want := range valid {
		host, port, err := ExtractHost(in)
		if err != nil {
			t.Errorf("ExtractHost(%q) unexpected error: %v", in, err)
		} else if host != want.host || port != want.port {
			t.Errorf("ExtractHost(%q) = (%q,%d), want (%q,%d)", in, host, port, want.host, want.port)
		}
	}
	for _, in := range []string{
		"", "  ", "exa mple.com", "http://", "foo/bar/baz/qux/../../../etc",
		"-sV", "--script=vuln", "-oN", ".example.com", "-target.com",
	} {
		if _, _, err := ExtractHost(in); err == nil {
			t.Errorf("ExtractHost(%q) expected error, got nil", in)
		}
	}
}

func TestParsePorts(t *testing.T) {
	// Default + explicit port from target.
	if ports, _, _, err := ParsePorts("", 8888); err != nil || len(ports) != 1 || ports[0] != 8888 {
		t.Errorf("default with explicit port: %v, %v", ports, err)
	}
	if ports, useTop, n, err := ParsePorts("top-100", 0); err != nil || !useTop || n != 100 || len(ports) != len(commonPorts) {
		t.Errorf("top-100: len=%d useTop=%v n=%d err=%v", len(ports), useTop, n, err)
	}
	if ports, _, _, err := ParsePorts("80,443, 8080-8081", 0); err != nil {
		t.Fatalf("list/range: %v", err)
	} else if fmt.Sprint(ports) != "[80 443 8080 8081]" {
		t.Errorf("list/range expanded to %v", ports)
	}
	// Flag/argument injection attempts must be rejected.
	for _, evil := range []string{
		"-oN", "-p", "--top-ports", "80; id", "80 && id", "80|80",
		"-sV", "0", "70000", "1-99999", "top-0", "top-5000", "http",
	} {
		if _, _, _, err := ParsePorts(evil, 0); err == nil {
			t.Errorf("ParsePorts(%q) expected rejection, got nil", evil)
		}
	}
}

func openTestListener(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func closedTestPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestNativeScanLocalhost(t *testing.T) {
	openPort, cleanup := openTestListener(t)
	defer cleanup()
	closedPort := closedTestPort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := ScanTarget(ctx, NmapOptions{
		Target:     "127.0.0.1",
		Ports:      fmt.Sprintf("%d,%d", openPort, closedPort),
		RateLimit:  10,
		Timeout:    30 * time.Second,
		PreferNmap: false,
	})
	if err != nil {
		t.Fatalf("ScanTarget: %v", err)
	}
	if res.EngineUsed != "native-go" {
		t.Errorf("engine = %q, want native-go", res.EngineUsed)
	}
	if len(res.OpenPorts) != 1 || res.OpenPorts[0].Port != openPort {
		t.Errorf("expected only port %d open, got %+v", openPort, res.OpenPorts)
	}
	if res.OpenPorts[0].Service == "" {
		t.Errorf("expected non-empty Service for open port %d, got empty", openPort)
	}
}

func TestScanTargetInvalidInput(t *testing.T) {
	ctx := context.Background()
	for _, opts := range []NmapOptions{
		{Target: ""},
		{Target: "-sV"},
		{Target: "--script=vuln"},
		{Target: "127.0.0.1", Ports: "-oN"},
		{Target: "exa mple.com"},
	} {
		if _, err := ScanTarget(ctx, opts); err == nil {
			t.Errorf("ScanTarget(%+v) expected error, got nil", opts)
		}
	}
}

func TestLookupCommonService(t *testing.T) {
	cases := map[int]string{
		21:    "ftp",
		22:    "ssh",
		80:    "http",
		443:   "https",
		3306:  "mysql",
		5432:  "postgresql",
		6379:  "redis",
		8080:  "http-proxy",
		27017: "mongodb",
		49999: "unknown",
	}
	for port, want := range cases {
		if got := lookupCommonService(port); got != want {
			t.Errorf("lookupCommonService(%d) = %q, want %q", port, got, want)
		}
	}
}

func TestIdentifyServiceFromBanner(t *testing.T) {
	cases := []struct {
		banner string
		port   int
		want   string
	}{
		{"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6", 2222, "ssh"},
		{"HTTP/1.1 200 OK\r\nServer: nginx\r\n", 8080, "http"},
		{"HTTP/1.1 200 OK\r\nServer: nginx\r\n", 443, "https"},
		{"<!DOCTYPE html><html><body>Welcome</body></html>", 8000, "http"},
		{"220 (vsFTPd 3.0.3)", 2121, "ftp"},
		{"220 mail.example.com ESMTP Postfix", 2525, "smtp"},
		{"+OK Dovecot ready.", 1100, "pop3"},
		{"* OK [CAPABILITY IMAP4rev1] Courier-IMAP ready.", 1430, "imap"},
		{"RFB 003.008\n", 5901, "vnc"},
		{"5.7.34-0ubuntu0.18.04.1\x00mysql_native_password", 3306, "mysql"},
		{"-ERR unknown command 'HELP'", 6379, "redis"},
		{"UNKNOWN PROTOCOL DATA", 9999, ""},
	}
	for _, tc := range cases {
		if got := identifyServiceFromBanner(tc.banner, tc.port); got != tc.want {
			t.Errorf("identifyServiceFromBanner(%q, %d) = %q, want %q", tc.banner, tc.port, got, tc.want)
		}
	}
}

func TestNmapBinaryIntegration(t *testing.T) {
	if findNmapBinary("") == "" {
		t.Skip("nmap binary not installed")
	}
	// Serve a banner so -sV completes immediately instead of waiting out
	// version-probe timeouts against a silent port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	openPort := ln.Addr().(*net.TCPAddr).Port
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_, _ = c.Write([]byte("SSH-2.0-CybermesTest\r\n"))
			c.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	res, err := ScanTarget(ctx, NmapOptions{
		Target:     "127.0.0.1",
		Ports:      fmt.Sprintf("%d", openPort),
		RateLimit:  25,
		Timeout:    60 * time.Second,
		PreferNmap: true,
	})
	if err != nil {
		t.Fatalf("nmap ScanTarget: %v", err)
	}
	if res.EngineUsed != "nmap" {
		t.Errorf("engine = %q, want nmap", res.EngineUsed)
	}
	found := false
	for _, p := range res.OpenPorts {
		if p.Port == openPort && strings.EqualFold(p.State, "open") {
			found = true
		}
	}
	if !found {
		t.Errorf("nmap missed open listener port %d: %+v", openPort, res.OpenPorts)
	}
}
