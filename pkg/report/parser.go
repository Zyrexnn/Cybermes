package report

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	frontmatterRe  = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	fmTitleRe      = regexp.MustCompile(`(?m)^title:\s*['"]?(.+?)['"]?$`)
	fmSeverityRe   = regexp.MustCompile(`(?mi)^severity:\s*['"]?([A-Za-z]+)['"]?$`)
	headingTitleRe = regexp.MustCompile(`(?m)^#\s+(?:(?:Vulnerability Report|Finding|Vuln):\s*)?(?:\[[A-Z]+\]\s*[-:]?\s*)?(?:[0-9]+\.\s*)?(.+)$`)

	tableSevRe = regexp.MustCompile(`(?i)\|\s*\*{0,2}(?:Severity|Severity Rating|Risk Level)\*{0,2}\s*\|\s*[\x60*]?([A-Za-z]+)`)
	kvSevRe    = regexp.MustCompile(`(?i)(?:Severity|Severity Rating|Risk Level)\s*[:=]\s*[\x60*]?([A-Za-z]+)`)

	tableCvssRe = regexp.MustCompile(`(?i)\|\s*\*{0,2}CVSS(?:\s*v?3(?:\.1)?)?(?:\s*Score)?\*{0,2}\s*\|\s*[\x60*]?([0-9\.]+(?:\s*\([^\)\|\n]+\))?)`)
	kvCvssRe    = regexp.MustCompile(`(?i)CVSS(?:\s*v?3(?:\.1)?)?(?:\s*Score)?\s*[:=]\s*[\x60*]?([0-9\.]+(?:\s*\([^\)\|\n]+\))?)`)

	tableCweRe = regexp.MustCompile(`(?i)\|\s*\*{0,2}CWE\*{0,2}\s*\|\s*[\x60*]?((?:CWE-)?\d+[^|*\n\x60]*)`)
	kvCweRe    = regexp.MustCompile(`(?i)CWE\s*[:=]\s*[\x60*]?((?:CWE-)?\d+[^|*\n\x60]*)`)

	tableEpRe = regexp.MustCompile(`(?i)\|\s*\*{0,2}(?:Affected Endpoint|Affected Asset|Target|Endpoint|URL/Host|URL)\*{0,2}\s*\|\s*[\x60*]?([^|*\n\x60]+)`)
	kvEpRe    = regexp.MustCompile(`(?i)(?:Affected Endpoint|Affected Asset|Target|Endpoint|URL/Host|URL)\s*[:=]\s*[\x60*]?([^|*\n\x60]+)`)

	prefixSevRe = regexp.MustCompile(`^(?:\[)?(CRITICAL|HIGH|MEDIUM|LOW|INFO|INFORMATIONAL)(?:\])?[-_]`)
	cleanValRe  = regexp.MustCompile("[\x60*_]")
)

var severityWeights = map[string]int{
	"CRITICAL":      1,
	"HIGH":          2,
	"MEDIUM":        3,
	"LOW":           4,
	"INFORMATIONAL": 5,
	"UNKNOWN":       6,
}

type FindingMeta struct {
	FileName     string `json:"file_name"`
	RelativePath string `json:"relative_path"`
	Title        string `json:"title"`
	Severity     string `json:"severity"`
	CVSS         string `json:"cvss"`
	CWE          string `json:"cwe"`
	Endpoint     string `json:"endpoint"`
	LastModified string `json:"last_modified"`
}

func CleanExtractedValue(val string) string {
	return strings.TrimSpace(cleanValRe.ReplaceAllString(val, ""))
}

func ParseFindingFile(filePath string) (*FindingMeta, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	content := string(data)
	filename := filepath.Base(filePath)

	var title string
	fmM := frontmatterRe.FindStringSubmatch(content)
	if len(fmM) > 1 {
		fmTitleM := fmTitleRe.FindStringSubmatch(fmM[1])
		if len(fmTitleM) > 1 {
			title = strings.TrimSpace(fmTitleM[1])
		}
	}
	if title == "" {
		hTitleM := headingTitleRe.FindStringSubmatch(content)
		if len(hTitleM) > 1 {
			title = strings.TrimSpace(hTitleM[1])
		} else {
			title = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSuffix(filename, ".md"), "_", " "), "-", " ")
		}
	}

	var severity string
	if len(fmM) > 1 {
		fmSevM := fmSeverityRe.FindStringSubmatch(fmM[1])
		if len(fmSevM) > 1 {
			severity = strings.ToUpper(strings.TrimSpace(fmSevM[1]))
		}
	}
	if severity == "" {
		if m := tableSevRe.FindStringSubmatch(content); len(m) > 1 {
			severity = strings.ToUpper(strings.TrimSpace(m[1]))
		} else if m := kvSevRe.FindStringSubmatch(content); len(m) > 1 {
			severity = strings.ToUpper(strings.TrimSpace(m[1]))
		} else if m := prefixSevRe.FindStringSubmatch(filename); len(m) > 1 {
			severity = strings.ToUpper(m[1])
		} else {
			severity = "UNKNOWN"
		}
	}

	if severity == "INFO" || severity == "NOTE" {
		severity = "INFORMATIONAL"
	}
	if _, ok := severityWeights[severity]; !ok {
		severity = "UNKNOWN"
	}

	var cvss string
	if m := tableCvssRe.FindStringSubmatch(content); len(m) > 1 {
		cvss = CleanExtractedValue(m[1])
	} else if m := kvCvssRe.FindStringSubmatch(content); len(m) > 1 {
		cvss = CleanExtractedValue(m[1])
	} else {
		cvss = "N/A"
	}

	var cwe string
	if m := tableCweRe.FindStringSubmatch(content); len(m) > 1 {
		raw := CleanExtractedValue(m[1])
		if strings.HasPrefix(strings.ToUpper(raw), "CWE-") {
			cwe = raw
		} else {
			cwe = "CWE-" + raw
		}
	} else if m := kvCweRe.FindStringSubmatch(content); len(m) > 1 {
		raw := CleanExtractedValue(m[1])
		if strings.HasPrefix(strings.ToUpper(raw), "CWE-") {
			cwe = raw
		} else {
			cwe = "CWE-" + raw
		}
	} else {
		cwe = "N/A"
	}

	var endpoint string
	if m := tableEpRe.FindStringSubmatch(content); len(m) > 1 {
		endpoint = CleanExtractedValue(m[1])
	} else if m := kvEpRe.FindStringSubmatch(content); len(m) > 1 {
		endpoint = CleanExtractedValue(m[1])
	} else {
		endpoint = "N/A"
	}

	info, _ := os.Stat(filePath)
	modTime := time.Now().Format("2006-01-02 15:04:05")
	if info != nil {
		modTime = info.ModTime().Format("2006-01-02 15:04:05")
	}

	return &FindingMeta{
		FileName:     filename,
		RelativePath: "findings/" + filename,
		Title:        title,
		Severity:     severity,
		CVSS:         cvss,
		CWE:          cwe,
		Endpoint:     endpoint,
		LastModified: modTime,
	}, nil
}

func ExtractCustomSections(existingContent string) string {
	if existingContent == "" {
		return ""
	}
	customPattern := regexp.MustCompile(`(?si)(##\s+(?:Verified Working Controls|Executive Narrative|Recommendations|Priority Action Items|Attack Path Narrative).*)$`)
	m := customPattern.FindStringSubmatch(existingContent)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
