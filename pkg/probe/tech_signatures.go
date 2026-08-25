package probe

import (
	"net/http"
	"strings"
)

// TechSignature defines a technology pattern match rule.
type TechSignature struct {
	Name        string
	Category    string
	Headers     map[string]string // Header Name (lowercase) -> Substring pattern
	Cookies     []string          // Cookie name substring
	BodyMatches []string          // Body substring patterns
}

var defaultSignatures = []TechSignature{
	{
		Name:     "Next.js",
		Category: "Frontend / Fullstack Framework",
		Headers: map[string]string{
			"x-powered-by": "next.js",
		},
		BodyMatches: []string{
			"__NEXT_DATA__",
			"/_next/static/",
			"/_next/data/",
		},
	},
	{
		Name:     "React",
		Category: "Frontend Library",
		BodyMatches: []string{
			"data-reactroot",
			"react-dom",
			"react.production.min.js",
		},
	},
	{
		Name:     "Vue.js",
		Category: "Frontend Framework",
		BodyMatches: []string{
			"data-v-",
			"vue.min.js",
			"vue.esm-browser.js",
			"__vue__",
		},
	},
	{
		Name:     "Angular",
		Category: "Frontend Framework",
		BodyMatches: []string{
			"ng-version=",
			"ng-app=",
			"ng-controller=",
		},
	},
	{
		Name:     "Laravel",
		Category: "Backend Framework (PHP)",
		Cookies: []string{
			"laravel_session",
			"XSRF-TOKEN",
		},
		BodyMatches: []string{
			"laravel",
			"csrf-token",
		},
	},
	{
		Name:     "Spring Boot",
		Category: "Backend Framework (Java)",
		Headers: map[string]string{
			"x-application-context": "",
		},
		BodyMatches: []string{
			"Whitelabel Error Page",
			"This application has no explicit mapping for /error",
		},
	},
	{
		Name:     "Django",
		Category: "Backend Framework (Python)",
		Cookies: []string{
			"csrftoken",
			"sessionid",
		},
		BodyMatches: []string{
			"csrfmiddlewaretoken",
		},
	},
	{
		Name:     "Express.js",
		Category: "Backend Framework (Node.js)",
		Headers: map[string]string{
			"x-powered-by": "express",
		},
	},
	{
		Name:     "FastAPI / Uvicorn",
		Category: "Backend Framework (Python)",
		Headers: map[string]string{
			"server": "uvicorn",
		},
	},
	{
		Name:     "WordPress",
		Category: "CMS (PHP)",
		Headers: map[string]string{
			"x-pingback": "",
		},
		Cookies: []string{
			"wordpress_logged_in",
			"wp-settings",
		},
		BodyMatches: []string{
			"/wp-content/",
			"/wp-includes/",
			"wp-json",
		},
	},
	{
		Name:     "PHP",
		Category: "Programming Language",
		Headers: map[string]string{
			"x-powered-by": "php",
		},
		Cookies: []string{
			"PHPSESSID",
		},
	},
	{
		Name:     "ASP.NET / IIS",
		Category: "Backend / Web Server (Microsoft)",
		Headers: map[string]string{
			"x-aspnet-version": "",
			"x-powered-by":     "asp.net",
			"server":           "microsoft-iis",
		},
		Cookies: []string{
			"ASP.NET_SessionId",
			"__RequestVerificationToken",
		},
	},
	{
		Name:     "Nginx",
		Category: "Web Server / Reverse Proxy",
		Headers: map[string]string{
			"server": "nginx",
		},
	},
	{
		Name:     "Apache HTTP Server",
		Category: "Web Server",
		Headers: map[string]string{
			"server": "apache",
		},
	},
	{
		Name:     "Cloudflare",
		Category: "CDN / WAF",
		Headers: map[string]string{
			"server":  "cloudflare",
			"cf-ray":  "",
			"cf-cache-status": "",
		},
	},
	{
		Name:     "Swagger / OpenAPI",
		Category: "API Documentation",
		BodyMatches: []string{
			"swagger-ui",
			"swagger-ui-bundle",
			"openapi:",
			"/swagger/v1/swagger.json",
		},
	},
	{
		Name:     "GraphQL",
		Category: "API Interface",
		BodyMatches: []string{
			"__schema",
			"GraphiQL",
			"graphql-ws",
		},
	},
}

// DetectTechnologies analyzes HTTP headers, cookies, and response body against known signatures.
func DetectTechnologies(headers http.Header, cookies []*http.Cookie, body string) []string {
	matched := make(map[string]bool)
	lowerBody := strings.ToLower(body)

	lowerHeaders := make(map[string]string)
	for k, v := range headers {
		lowerHeaders[strings.ToLower(k)] = strings.ToLower(strings.Join(v, " "))
	}

	cookieNames := make([]string, 0, len(cookies))
	for _, c := range cookies {
		cookieNames = append(cookieNames, strings.ToLower(c.Name))
	}

	for _, sig := range defaultSignatures {
		detected := false

		// Check Headers
		for hk, hv := range sig.Headers {
			if val, exists := lowerHeaders[hk]; exists {
				if hv == "" || strings.Contains(val, hv) {
					detected = true
					break
				}
			}
		}

		// Check Cookies
		if !detected {
			for _, cn := range sig.Cookies {
				targetCookie := strings.ToLower(cn)
				for _, cName := range cookieNames {
					if strings.Contains(cName, targetCookie) {
						detected = true
						break
					}
				}
				if detected {
					break
				}
			}
		}

		// Check Body Matches
		if !detected && len(sig.BodyMatches) > 0 && len(lowerBody) > 0 {
			for _, bm := range sig.BodyMatches {
				if strings.Contains(lowerBody, strings.ToLower(bm)) {
					detected = true
					break
				}
			}
		}

		if detected {
			matched[sig.Name] = true
		}
	}

	result := make([]string, 0, len(matched))
	for tech := range matched {
		result = append(result, tech)
	}
	return result
}
