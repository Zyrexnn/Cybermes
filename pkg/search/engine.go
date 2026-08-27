package search

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

var (
	headingRegex = regexp.MustCompile(`^#{1,4}\s+(.+)$`)
	codeBlockRe  = regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\n.*?```")
)

var kbMapping = map[string]string{
	"payloads":   "PayloadsAllTheThings",
	"hacktricks": "hacktricks",
	"claude":     "Claude-BugHunter",
	"strix":      "strix-skills",
	"hack":       "hack-skills",
}

var synonymMap = map[string][]string{
	"idor": {"insecure direct object", "bola", "broken object"},
	"bola": {"insecure direct object", "idor", "broken object level"},
	"bpla": {"broken property level authorization", "mass assignment"},
	"sqli": {"sql injection", "sql-injection"},
	"xss":  {"cross-site scripting", "cross site scripting"},
	"csrf": {"cross-site request forgery", "xsrf"},
	"xsrf": {"cross-site request forgery", "csrf"},
	"rce":  {"remote code execution", "command injection"},
	"ssrf": {"server side request forgery", "server-side request forgery"},
	"xxe":  {"xml external entity"},
	"jwt":  {"json web token"},
	"lfi":  {"local file inclusion", "path traversal", "directory traversal"},
	"rfi":  {"remote file inclusion"},
	"cors": {"cross-origin resource sharing"},
	"crlf": {"http response splitting", "carriage return"},
}

func expandKeywords(terms []string) []string {
	seen := make(map[string]struct{}, len(terms)*2)
	var result []string

	for _, term := range terms {
		t := strings.TrimSpace(strings.ToLower(term))
		if len(t) <= 1 {
			continue
		}
		if _, exists := seen[t]; !exists {
			seen[t] = struct{}{}
			result = append(result, t)
		}
		if synonyms, ok := synonymMap[t]; ok {
			for _, syn := range synonyms {
				if _, exists := seen[syn]; !exists {
					seen[syn] = struct{}{}
					result = append(result, syn)
				}
			}
		}
	}
	return result
}

type Snippet struct {
	Heading   string `json:"heading"`
	StartLine int    `json:"start_line"`
	Score     int    `json:"score"`
	Content   string `json:"content"`
	File      string `json:"file"`
	SourceKB  string `json:"source_kb"`
}

type Searcher struct {
	BaseDir string
	RootDir string
}

func NewSearcher(baseDir, rootDir string) *Searcher {
	return &Searcher{
		BaseDir: baseDir,
		RootDir: rootDir,
	}
}

type candidateFile struct {
	path  string
	score int
}

func (s *Searcher) Search(query string, source string, limit int, maxChars int) ([]Snippet, error) {
	searchPath := s.BaseDir
	if sub, ok := kbMapping[source]; ok {
		targetSub := filepath.Join(s.BaseDir, sub)
		if _, err := os.Stat(targetSub); err == nil {
			searchPath = targetSub
		}
	}

	rawTerms := strings.Fields(strings.ToLower(query))
	keywords := expandKeywords(rawTerms)
	if len(keywords) == 0 {
		return nil, nil
	}

	candidates, err := s.findCandidateFiles(searchPath, keywords)
	if err != nil {
		return nil, err
	}

	var allSnippets []Snippet
	maxCandidates := 25
	if len(candidates) < maxCandidates {
		maxCandidates = len(candidates)
	}

	for i := 0; i < maxCandidates; i++ {
		fileSnippets := s.extractSnippets(candidates[i].path, keywords, maxChars)
		allSnippets = append(allSnippets, fileSnippets...)
	}

	sort.SliceStable(allSnippets, func(i, j int) bool {
		return allSnippets[i].Score > allSnippets[j].Score
	})

	if limit > 0 && len(allSnippets) > limit {
		allSnippets = allSnippets[:limit]
	}

	return allSnippets, nil
}

func (s *Searcher) findCandidateFiles(searchPath string, keywords []string) ([]candidateFile, error) {
	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return nil, nil
	}
	var fileList []string

	err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "site" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		name := strings.ToLower(d.Name())
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".txt") {
			if name != "summary.md" && name != "_sidebar.md" && name != "toc.md" {
				fileList = append(fileList, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	type fileJob struct {
		path string
	}

	jobs := make(chan fileJob, len(fileList))
	for _, f := range fileList {
		jobs <- fileJob{path: f}
	}
	close(jobs)

	numWorkers := 8
	var mu sync.Mutex
	var scored []candidateFile
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				data, err := os.ReadFile(job.path)
				if err != nil {
					continue
				}
				if bytes.IndexByte(data, 0) != -1 || !utf8.Valid(data) {
					continue
				}
				contentLower := strings.ToLower(string(data))
				totalHits := 0
				uniqueHits := 0
				for _, kw := range keywords {
					count := strings.Count(contentLower, kw)
					if count > 0 {
						totalHits += count
						uniqueHits++
					}
				}
				if totalHits > 0 {
					score := totalHits + (uniqueHits * 15)
					mu.Lock()
					scored = append(scored, candidateFile{path: job.path, score: score})
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	return scored, nil
}

type section struct {
	heading   string
	startLine int
	text      string
}

func (s *Searcher) detectKB(filePath string) string {
	rel, err := filepath.Rel(s.BaseDir, filePath)
	if err != nil {
		return "knowledge"
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) > 0 && parts[0] != "." {
		return parts[0]
	}
	return "knowledge"
}

func (s *Searcher) extractSnippets(filePath string, keywords []string, maxChars int) []Snippet {
	data, err := os.ReadFile(filePath)
	if err != nil || bytes.IndexByte(data, 0) != -1 || !utf8.Valid(data) {
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var sections []section
	currHeading := "General"
	var currLines []string
	startLine := 1
	lineNo := 1

	for scanner.Scan() {
		line := scanner.Text()
		if headingRegex.MatchString(line) {
			if len(currLines) > 0 {
				sections = append(sections, section{
					heading:   currHeading,
					startLine: startLine,
					text:      strings.Join(currLines, "\n"),
				})
			}
			currHeading = strings.TrimSpace(line)
			currLines = []string{line}
			startLine = lineNo
		} else {
			currLines = append(currLines, line)
		}
		lineNo++
	}
	if len(currLines) > 0 {
		sections = append(sections, section{
			heading:   currHeading,
			startLine: startLine,
			text:      strings.Join(currLines, "\n"),
		})
	}

	kbSource := s.detectKB(filePath)
	relPath, err := filepath.Rel(s.RootDir, filePath)
	if err != nil {
		relPath = filePath
	}

	fileNameLower := strings.ToLower(filepath.Base(filePath))
	var snippets []Snippet

	for _, sec := range sections {
		lowerText := strings.ToLower(sec.text)
		lowerHeading := strings.ToLower(sec.heading)

		hitCount := 0
		for _, kw := range keywords {
			hitCount += strings.Count(lowerText, kw)
		}
		if hitCount == 0 {
			continue
		}

		score := hitCount * 5
		for _, kw := range keywords {
			if strings.Contains(lowerHeading, kw) {
				score += 40
				break
			}
		}
		if strings.Contains(sec.text, "```") {
			score += 35
		}
		for _, term := range []string{"payload", "bypass", "exploit", "poc", "syntax", "example"} {
			if strings.Contains(lowerText, term) {
				score += 25
				break
			}
		}
		if fileNameLower == "summary.md" || fileNameLower == "_sidebar.md" {
			score -= 60
		}

		trimmedText := sec.text
		if len(sec.text) > maxChars {
			matches := codeBlockRe.FindAllString(sec.text, -1)
			if len(matches) > 0 && len(matches[0]) < maxChars {
				trimmedText = sec.heading + "\n\n" + matches[0] + "\n\n*(Truncated for context efficiency)*"
			} else {
				trimmedText = strings.TrimRight(sec.text[:maxChars], " \t\r\n") + "\n\n*(Truncated for context efficiency)*"
			}
		}

		snippets = append(snippets, Snippet{
			Heading:   sec.heading,
			StartLine: sec.startLine,
			Score:     score,
			Content:   trimmedText,
			File:      relPath,
			SourceKB:  kbSource,
		})
	}

	sort.SliceStable(snippets, func(i, j int) bool {
		return snippets[i].Score > snippets[j].Score
	})

	if len(snippets) > 2 {
		snippets = snippets[:2]
	}
	return snippets
}
