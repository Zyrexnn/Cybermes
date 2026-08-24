package secrets

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	SevCritical = "critical"
	SevHigh     = "high"
	SevMedium   = "medium"
	SevLow      = "low"
)

type Finding struct {
	Pattern  string `json:"pattern"`
	Severity string `json:"severity"`
	Category string `json:"category"`
	Match    string `json:"match"`
	Source   string `json:"source"`
	Line     int    `json:"line"`
}

type patternDef struct {
	name     string
	severity string
	category string
	re       *regexp.Regexp
}

var compiledPatterns []patternDef

func init() {
	raw := []struct {
		name     string
		severity string
		category string
		expr     string
	}{
		// AWS
		{"AWS_ACCESS_KEY", SevCritical, "aws", `\b(AKIA|ASIA)[0-9A-Z]{16}\b`},
		{"AWS_SECRET_TYPED", SevCritical, "aws", `(?i)aws[_\-]?secret[_\-]?access[_\-]?key['"\s:=]+([A-Za-z0-9/+=]{40})`},
		{"AWS_SECRET_LOOSE", SevHigh, "aws", `(?i)aws(.{0,20})?(secret|sk)['"=: ]+([0-9a-z/+=]{40})`},

		// Google Cloud Platform
		{"GCP_SERVICE_ACCOUNT", SevCritical, "gcp", `"type"\s*:\s*"service_account"`},
		{"GOOGLE_API_KEY", SevHigh, "gcp", `\bAIza[0-9A-Za-z_\-]{35}\b`},

		// GitHub
		{"GH_PAT_CLASSIC", SevCritical, "github", `\bghp_[A-Za-z0-9]{36}\b`},
		{"GH_PAT_FINEGRAINED", SevCritical, "github", `\bgithub_pat_[A-Za-z0-9_]{82}\b`},
		{"GH_OAUTH", SevHigh, "github", `\bgho_[A-Za-z0-9]{36}\b`},
		{"GH_S2S", SevHigh, "github", `\bgh[usr]_[A-Za-z0-9]{36,}\b`},

		// Stripe
		{"STRIPE_LIVE", SevCritical, "stripe", `\bsk_live_[0-9A-Za-z]{24,}\b`},
		{"STRIPE_TEST", SevLow, "stripe", `\bsk_test_[0-9A-Za-z]{24,}\b`},

		// Slack
		{"SLACK_TOKEN", SevHigh, "slack", `\bxox[abpors]-[0-9A-Za-z\-]{10,48}\b`},
		{"SLACK_WEBHOOK", SevMedium, "slack", `https://hooks\.slack\.com/services/T[A-Z0-9]+/B[A-Z0-9]+/[A-Za-z0-9]+`},

		// Email service providers
		{"SENDGRID", SevHigh, "email_svc", `\bSG\.[A-Za-z0-9_\-]{22}\.[A-Za-z0-9_\-]{43}\b`},
		{"MAILGUN_V1", SevHigh, "email_svc", `\bkey-[0-9a-zA-Z]{32}\b`},
		{"MAILGUN_LOOSE", SevHigh, "email_svc", `\bkey-[0-9a-f]{32}\b`},

		// Twilio
		{"TWILIO_API", SevHigh, "twilio", `\bSK[0-9a-fA-F]{32}\b`},
		{"TWILIO_SID", SevMedium, "twilio", `\bAC[a-f0-9]{32}\b`},
		{"TWILIO_AUTH", SevHigh, "twilio", `(?i)twilio(.{0,20})?(auth|token)['"=: ]+([a-f0-9]{32})`},

		// PaaS
		{"HEROKU_API", SevMedium, "paas", `(?i)heroku(.{0,20})?api['"=: ]+([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`},

		// Firebase
		{"FIREBASE_URL", SevLow, "firebase", `\bhttps?://[a-z0-9\-]+\.firebaseio\.com\b`},

		// Tokens / auth headers
		{"JWT", SevMedium, "jwt", `\beyJ[A-Za-z0-9_\-]{10,}\.eyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\b`},
		{"BEARER_AUTH", SevMedium, "bearer", `(?i)authorization['"=: ]+bearer\s+[A-Za-z0-9._\-]{20,}`},
		{"BASIC_AUTH_URL", SevMedium, "basic_auth", `https?://[^/\s:@]+:[^/\s:@]+@[^/\s]+`},

		// Private keys
		{"RSA_PRIVKEY", SevCritical, "private_key", `-----BEGIN RSA PRIVATE KEY-----`},
		{"EC_PRIVKEY", SevCritical, "private_key", `-----BEGIN EC PRIVATE KEY-----`},
		{"OPENSSH_PRIVKEY", SevCritical, "private_key", `-----BEGIN OPENSSH PRIVATE KEY-----`},
		{"GENERIC_PRIVKEY", SevCritical, "private_key", `-----BEGIN (DSA |PGP |)PRIVATE KEY-----`},

		// Generic
		{"GENERIC_API_KEY", SevMedium, "generic", `(?i)(?:api[_\-]?key|apikey|api_secret|access_token|secret[_\-]?token)['"\s:=]+["']([A-Za-z0-9+/=_\-]{24,})["']`},

		// Modern AI APIs
		{"ANTHROPIC_API", SevCritical, "ai_api", `\bsk-ant-(?:api03|admin01)-[A-Za-z0-9_\-]{93,}\b`},
		{"OPENAI_LEGACY", SevCritical, "ai_api", `\bsk-[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20}\b`},
		{"OPENAI_PROJECT", SevCritical, "ai_api", `\bsk-proj-[A-Za-z0-9_\-]{40,}T3BlbkFJ[A-Za-z0-9_\-]{40,}\b`},
		{"OPENAI_SESSION", SevHigh, "ai_api", `\bsess-[A-Za-z0-9]{40}\b`},
		{"HUGGINGFACE", SevHigh, "ai_api", `\bhf_[A-Za-z0-9]{30,}\b`},

		// Cloud infra
		{"CLOUDFLARE_API", SevCritical, "infra_api", `(?i)cf[_\-]?api[_\-]?key['"\s:=]+([a-f0-9]{37})`},
		{"DIGITALOCEAN", SevHigh, "infra_api", `\bdop_v1_[a-f0-9]{64}\b`},

		// Package registries
		{"NPM_TOKEN", SevHigh, "package_registry", `\bnpm_[A-Za-z0-9]{36}\b`},
		{"PYPI_TOKEN", SevHigh, "package_registry", `\bpypi-AgENdGV[A-Za-z0-9_\-]+\b`},
		{"DOCKER_HUB_PAT", SevHigh, "package_registry", `\bdckr_pat_[A-Za-z0-9_\-]{27,}\b`},

		// SaaS
		{"ATLASSIAN_TOKEN", SevHigh, "saas_api", `\bATATT3xFfGF0[A-Za-z0-9_\-]{180,}\b`},
		{"LINEAR_API", SevMedium, "saas_api", `\blin_api_[A-Za-z0-9]{40}\b`},

		// Observability
		{"NEWRELIC_LICENSE", SevMedium, "observability", `\b(?:NRAA|NRAK|NRBR)-[A-F0-9]{27}\b`},
		{"DATADOG_API", SevHigh, "observability", `(?i)dd[_\-]?api[_\-]?key['"\s:=]+([a-f0-9]{32})`},
		{"SENTRY_DSN", SevLow, "observability", `https://[a-f0-9]+@o[0-9]+\.ingest\.sentry\.io/[0-9]+`},

		// Tunneling
		{"NGROK_AUTH", SevMedium, "tunneling", `\b[12][A-Za-z0-9]{26}_[A-Za-z0-9]{32,}\b`},

		// Bot tokens
		{"DISCORD_BOT", SevHigh, "bot_token", `\b[MN][A-Za-z\d]{23}\.[\w\-]{6}\.[\w\-]{27}\b`},
		{"TELEGRAM_BOT", SevHigh, "bot_token", `\b\d{8,10}:[A-Za-z0-9_\-]{35}\b`},
	}

	for _, p := range raw {
		re := regexp.MustCompile(p.expr)
		compiledPatterns = append(compiledPatterns, patternDef{
			name:     p.name,
			severity: p.severity,
			category: p.category,
			re:       re,
		})
	}
}

func ScanLine(line string, lineNo int, source string) []Finding {
	var findings []Finding
	for _, p := range compiledPatterns {
		matches := p.re.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			matchText := m[0]
			if len(m) > 1 && m[1] != "" {
				matchText = m[1]
			}
			findings = append(findings, Finding{
				Pattern:  p.name,
				Severity: p.severity,
				Category: p.category,
				Match:    matchText,
				Source:   source,
				Line:     lineNo,
			})
		}
	}
	return findings
}

func ScanReader(r io.Reader, source string) []Finding {
	var findings []Finding
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 20*1024*1024)

	lineNo := 1
	for scanner.Scan() {
		line := scanner.Text()
		f := ScanLine(line, lineNo, source)
		if len(f) > 0 {
			findings = append(findings, f...)
		}
		lineNo++
	}
	return findings
}

func ScanText(text string, source string) []Finding {
	return ScanReader(strings.NewReader(text), source)
}

func ScanFile(filePath string) ([]Finding, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ScanReader(f, filePath), nil
}

func ScanDirectory(dirPath string, maxWorkers int) ([]Finding, error) {
	if maxWorkers <= 0 {
		maxWorkers = 8
	}

	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	fileChan := make(chan string, len(files))
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	var mu sync.Mutex
	var allFindings []Finding
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				findings, err := ScanFile(file)
				if err == nil && len(findings) > 0 {
					mu.Lock()
					allFindings = append(allFindings, findings...)
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return allFindings, nil
}
