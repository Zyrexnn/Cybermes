package report

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type ReportArtifacts struct {
	Target       string `json:"target"`
	SummaryPath  string `json:"summary_path"`
	MetadataPath string `json:"metadata_path"`
	HTMLPath     string `json:"html_path"`
	PDFPath      string `json:"pdf_path"`
	PDFGenerated bool   `json:"pdf_generated"`
	BrowserUsed  string `json:"browser_used,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// BuildFileURL converts a local absolute filesystem path into a valid file:/// URI across OS platforms
func BuildFileURL(absPath string) string {
	cleanPath := filepath.ToSlash(absPath)
	if strings.HasPrefix(cleanPath, "/") {
		return "file://" + cleanPath
	}
	return "file:///" + cleanPath
}

// FindChromiumBrowser scans known filesystem paths and system PATH for a Chromium-based browser
func FindChromiumBrowser() string {
	// 1. Environment variables
	envVars := []string{"CHROME_BIN", "PUPPETEER_EXECUTABLE_PATH", "BROWSER_PATH", "CYBERMES_CHROME_PATH"}
	for _, ev := range envVars {
		if val := os.Getenv(ev); val != "" {
			if _, err := os.Stat(val); err == nil {
				return val
			}
		}
	}

	// 2. OS-specific default installation paths
	var candidatePaths []string
	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")

		candidatePaths = []string{
			filepath.Join(programFiles, `Microsoft\Edge\Application\msedge.exe`),
			filepath.Join(programFilesX86, `Microsoft\Edge\Application\msedge.exe`),
			filepath.Join(programFiles, `Google\Chrome\Application\chrome.exe`),
			filepath.Join(programFilesX86, `Google\Chrome\Application\chrome.exe`),
			filepath.Join(programFiles, `BraveSoftware\Brave-Browser\Application\brave.exe`),
			filepath.Join(localAppData, `Microsoft\Edge\Application\msedge.exe`),
			filepath.Join(localAppData, `Google\Chrome\Application\chrome.exe`),
		}
	case "darwin":
		candidatePaths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
		}
	default: // Linux, Docker, BSD
		candidatePaths = []string{
			"/usr/bin/google-chrome-stable",
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
			"/snap/bin/chromium",
			"/usr/local/bin/chrome",
			"/usr/local/bin/chromium",
		}
	}

	for _, p := range candidatePaths {
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// 3. System PATH lookup
	binaryNames := []string{"google-chrome-stable", "google-chrome", "chromium-browser", "chromium", "msedge", "chrome", "brave"}
	for _, bin := range binaryNames {
		if path, err := exec.LookPath(bin); err == nil {
			return path
		}
	}

	return ""
}

// RenderPDF converts a local HTML file to an A4 PDF document using chromedp (Chrome DevTools Protocol)
func RenderPDF(htmlPath string, pdfPath string) (string, error) {
	absHTML, err := filepath.Abs(htmlPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for HTML: %w", err)
	}

	if _, err := os.Stat(absHTML); os.IsNotExist(err) {
		return "", fmt.Errorf("source HTML file does not exist: %s", absHTML)
	}

	browserPath := FindChromiumBrowser()
	if browserPath == "" {
		return "", fmt.Errorf("no Chromium browser (Chrome, Edge, Chromium) found in environment")
	}

	// Prepare execution allocator options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.NoSandbox,
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("headless", "new"),
		chromedp.Flag("hide-scrollbars", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Timeout context for rendering
	ctx, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer timeoutCancel()

	fileURL := BuildFileURL(absHTML)
	var pdfBuffer []byte

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond), // allow font rendering and CSS animations to settle
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).  // A4 Width in inches
				WithPaperHeight(11.69). // A4 Height in inches
				WithMarginTop(0.4).
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)

	if err != nil {
		return browserPath, fmt.Errorf("chromedp PDF generation failed: %w", err)
	}

	if err := os.WriteFile(pdfPath, pdfBuffer, 0666); err != nil {
		return browserPath, fmt.Errorf("failed to save PDF file to %s: %w", pdfPath, err)
	}

	return browserPath, nil
}

// GenerateFullReport generates SUMMARY.md, metadata.json, report.html, and optionally REPORT.pdf
func GenerateFullReport(targetDir string, data *SummaryData, generatePDF bool) (*ReportArtifacts, error) {
	targetName := filepath.Base(targetDir)
	summaryPath := filepath.Join(targetDir, "SUMMARY.md")
	metaPath := filepath.Join(targetDir, "metadata.json")
	pdfPath := filepath.Join(targetDir, "REPORT.pdf")

	artifacts := &ReportArtifacts{
		Target:       targetName,
		SummaryPath:  summaryPath,
		MetadataPath: metaPath,
		PDFPath:      pdfPath,
	}

	// 1. Generate report.html
	htmlPath, err := GenerateHTMLDashboard(targetDir, data)
	if err != nil {
		return artifacts, fmt.Errorf("failed to generate report.html: %w", err)
	}
	artifacts.HTMLPath = htmlPath

	// 2. Generate REPORT.pdf if requested
	if generatePDF {
		browserUsed, err := RenderPDF(htmlPath, pdfPath)
		if err != nil {
			artifacts.PDFGenerated = false
			artifacts.ErrorMessage = err.Error()
		} else {
			artifacts.PDFGenerated = true
			artifacts.BrowserUsed = browserUsed
		}
	}

	return artifacts, nil
}
