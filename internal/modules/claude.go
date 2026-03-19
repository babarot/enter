package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
)

type ClaudeModule struct{}

func (m *ClaudeModule) Name() string { return config.ModuleClaude }

type usageData struct {
	FiveHour *usageWindow `json:"five_hour"`
	SevenDay *usageWindow `json:"seven_day"`
	FetchedAt int64       `json:"_fetched_at,omitempty"`
}

type usageWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

func (m *ClaudeModule) Run(ctx *module.Context) *module.Output {
	cfg := &ctx.Config.Modules.Claude
	if !cfg.Enabled {
		return nil
	}

	if cfg.Mode == config.ClaudeModeAuto && !detectClaudeProject(ctx.Cwd) {
		return nil
	}

	var segments []module.Segment
	var rows []module.Row

	// Usage (5h + 7d windows)
	usageSegs, usageRows := buildUsageOutput(cfg)
	segments = append(segments, usageSegs...)
	rows = append(rows, usageRows...)

	// Config view
	if cfg.Fields.Config.Enabled {
		configSegs, configRow := buildConfigOutput(ctx.Cwd, cfg.Fields.Config.Mode)
		if configRow != nil {
			rows = append(rows, *configRow)
			if len(segments) > 0 {
				segments = append(segments, module.Plain(" | "))
			}
			segments = append(segments, configSegs...)
		}
	}

	if len(segments) == 0 {
		return nil
	}

	return &module.Output{
		Name:     m.Name(),
		Segments: segments,
		Rows:     rows,
	}
}

func buildUsageOutput(cfg *config.ClaudeConfig) ([]module.Segment, []module.Row) {
	usage := fetchUsage(cfg.Fields.Usage.CacheTTL)
	if usage == nil {
		return nil, nil
	}

	barStyle := cfg.Fields.Usage.BarStyle
	timeStyle := cfg.Fields.Usage.TimeStyle
	barWidth := 10

	var segments []module.Segment
	var rows []module.Row

	if usage.FiveHour != nil {
		segs, row := buildWindowRow("claude.usage.5h", "current", "time",
			usage.FiveHour, barWidth, barStyle, timeStyle)
		segments = append(segments, segs...)
		rows = append(rows, row)
	}

	if usage.SevenDay != nil {
		if len(segments) > 0 {
			segments = append(segments, module.Plain(" | "))
		}
		segs, row := buildWindowRow("claude.usage.7d", "weekly", "datetime",
			usage.SevenDay, barWidth, barStyle, timeStyle)
		segments = append(segments, segs...)
		rows = append(rows, row)
	}

	// Combined row (claude.usage) when both windows exist
	if len(rows) >= 2 {
		var combinedSegs []module.Segment
		for i, row := range rows {
			label := "5h "
			if row.Key == "claude.usage.7d" {
				label = "7d "
			}
			if len(combinedSegs) > 0 {
				combinedSegs = append(combinedSegs, module.Plain("\n"))
			}
			combinedSegs = append(combinedSegs, module.NewSegment(label, module.Muted))
			combinedSegs = append(combinedSegs, rows[i].Segments...)
		}
		rows = append(rows, module.Row{
			Key:      "claude.usage",
			Segments: combinedSegs,
		})
	}

	return segments, rows
}

func buildWindowRow(key, inlineLabel, displayStyle string, window *usageWindow,
	barWidth int, barStyle, timeStyle string) ([]module.Segment, module.Row) {

	pct := int(window.Utilization)
	reset := formatReset(window.ResetsAt, displayStyle, timeStyle)
	bar := buildBar(pct, barWidth, barStyle)
	color := pctColor(pct)

	segments := []module.Segment{
		module.NewSegment(fmt.Sprintf("%s %s %d%% ⟳ %s", inlineLabel, bar, pct, reset), color),
	}

	row := module.Row{
		Key: key,
		Segments: []module.Segment{
			module.NewSegment(bar+" ", module.Default),
			module.NewSegment(fmt.Sprintf("%3d%%", pct), color),
			module.Plain(" ⟳ "),
			module.NewSegment(reset, module.Muted),
		},
	}

	return segments, row
}

func buildConfigOutput(cwd, mode string) ([]module.Segment, *module.Row) {
	segs := buildConfigView(cwd, mode)
	if len(segs) == 0 {
		return nil, nil
	}
	return segs, &module.Row{
		Key:      "claude.config",
		Segments: segs,
	}
}

// detectClaudeProject checks if cwd (or git root) contains .claude/ or CLAUDE.md
func detectClaudeProject(cwd string) bool {
	if hasClaudeFiles(cwd) {
		return true
	}

	if root, ok := execGit(cwd, "rev-parse", "--show-toplevel"); ok {
		if root != cwd && hasClaudeFiles(root) {
			return true
		}
	}

	return false
}

func hasClaudeFiles(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		return true
	}
	return false
}

func buildBar(pct, width int, style string) string {
	filled, empty := "▰", "▱" // block (default)
	switch style {
	case config.BarStyleDot:
		filled, empty = "●", "○"
	case config.BarStyleFill:
		filled, empty = "█", "░"
	}

	n := pct * width / 100
	if n > width {
		n = width
	}
	if n < 0 {
		n = 0
	}

	var b strings.Builder
	for i := 0; i < width; i++ {
		if i < n {
			b.WriteString(filled)
		} else {
			b.WriteString(empty)
		}
	}
	return b.String()
}

// formatReset formats a reset time.
// displayStyle: "time" (for current) or "datetime" (for weekly)
// timeStyle: "absolute" or "relative"
func formatReset(isoStr, displayStyle, timeStyle string) string {
	if isoStr == "" {
		return "?"
	}
	t, err := time.Parse(time.RFC3339, isoStr)
	if err != nil {
		return "?"
	}

	if timeStyle == config.TimeStyleRelative {
		return formatRelativeTime(t)
	}

	t = t.Local()

	switch displayStyle {
	case "time":
		return strings.ToLower(t.Format("3:04pm"))
	case "datetime":
		return t.Format("Jan 2, ") + strings.ToLower(t.Format("3:04pm"))
	default:
		return strings.ToLower(t.Format("3:04pm"))
	}
}

func formatRelativeTime(target time.Time) string {
	diff := time.Until(target)
	if diff <= 0 {
		return "now"
	}

	totalMin := int(diff.Minutes())
	days := totalMin / 1440
	hours := (totalMin % 1440) / 60
	mins := totalMin % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if mins > 0 && days == 0 {
		parts = append(parts, fmt.Sprintf("%dm", mins))
	}
	if len(parts) == 0 {
		return "now"
	}
	return strings.Join(parts, " ") + " left"
}

func pctColor(pct int) module.SemanticColor {
	switch {
	case pct >= 80:
		return module.Danger
	case pct >= 60:
		return module.Warning
	default:
		return module.Success
	}
}

// --- claude.config view ---

type configItem struct {
	label  string
	exists bool
	count  int // -1 means not a directory (just a file check)
}

func buildConfigView(cwd, mode string) []module.Segment {
	// Find project root (git root or cwd)
	root := cwd
	if toplevel, ok := execGit(cwd, "rev-parse", "--show-toplevel"); ok {
		root = toplevel
	}

	items := []configItem{
		checkFile(root, "CLAUDE.md"),
		checkFile(root, ".claude/settings.json"),
		checkFile(root, ".claude/settings.local.json"),
		checkDir(root, ".claude/rules"),
		checkDir(root, ".claude/skills"),
		checkDir(root, ".claude/agents"),
		checkDir(root, ".claude/commands"),
		checkFile(root, ".mcp.json"),
	}

	var segments []module.Segment
	first := true
	for _, item := range items {
		if mode == config.ClaudeModeAuto && !item.exists {
			continue
		}

		if !first {
			segments = append(segments, module.Plain("\n"))
		}
		first = false

		if item.exists {
			segments = append(segments, module.NewSegment("✓ ", module.Success))
		} else {
			segments = append(segments, module.NewSegment("✗ ", module.Muted))
		}

		label := item.label
		if item.count >= 0 && item.exists {
			label = fmt.Sprintf("%s (%d)", item.label, item.count)
		}

		color := module.Secondary
		if !item.exists {
			color = module.Muted
		}
		segments = append(segments, module.NewSegment(label, color))
	}

	return segments
}

func checkFile(root, rel string) configItem {
	path := filepath.Join(root, rel)
	_, err := os.Stat(path)
	return configItem{
		label:  rel,
		exists: err == nil,
		count:  -1,
	}
}

func checkDir(root, rel string) configItem {
	path := filepath.Join(root, rel)
	entries, err := os.ReadDir(path)
	if err != nil {
		return configItem{label: rel, exists: false, count: 0}
	}
	// Count only files (not subdirs starting with .)
	count := 0
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			count++
		}
	}
	return configItem{label: rel, exists: true, count: count}
}

// --- OAuth token resolution ---

type claudeCredentials struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

func extractToken(data []byte) string {
	var creds claudeCredentials
	if json.Unmarshal(data, &creds) == nil {
		token := creds.ClaudeAiOauth.AccessToken
		if token != "" && token != "null" {
			return token
		}
	}
	return ""
}

func getOAuthToken() string {
	// 1. Environment variable
	if token := os.Getenv("CLAUDE_CODE_OAUTH_TOKEN"); token != "" {
		return token
	}

	// 2. macOS Keychain
	if out, err := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w").Output(); err == nil {
		if token := extractToken(out); token != "" {
			return token
		}
	}

	// 3. Credentials file
	if home, _ := os.UserHomeDir(); home != "" {
		if data, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json")); err == nil {
			if token := extractToken(data); token != "" {
				return token
			}
		}
	}

	return ""
}

// --- Usage API with cache ---

var (
	cacheDir  = filepath.Join(os.TempDir(), "claude")
	cacheFile = filepath.Join(os.TempDir(), "claude", "enter-usage-cache.json")
)

func fetchUsage(cacheTTL int) *usageData {
	// Check cache
	if info, err := os.Stat(cacheFile); err == nil {
		age := time.Since(info.ModTime()).Seconds()
		if age < float64(cacheTTL) {
			if data, err := os.ReadFile(cacheFile); err == nil {
				var usage usageData
				if json.Unmarshal(data, &usage) == nil {
					return &usage
				}
			}
		}
	}

	// Fetch from API
	token := getOAuthToken()
	if token == "" {
		return readStaleCache()
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return readStaleCache()
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("User-Agent", "enter/0.1.0")

	resp, err := client.Do(req)
	if err != nil {
		return readStaleCache()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readStaleCache()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return readStaleCache()
	}

	var usage usageData
	if json.Unmarshal(body, &usage) != nil || usage.FiveHour == nil {
		return readStaleCache()
	}

	usage.FetchedAt = time.Now().UnixMilli()

	// Write cache
	os.MkdirAll(cacheDir, 0o755)
	if data, err := json.Marshal(&usage); err == nil {
		os.WriteFile(cacheFile, data, 0o644)
	}

	return &usage
}

func readStaleCache() *usageData {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}
	var usage usageData
	if json.Unmarshal(data, &usage) != nil {
		return nil
	}
	return &usage
}
