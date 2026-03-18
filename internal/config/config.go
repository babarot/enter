package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Format
const (
	FormatTable  = "table"
	FormatInline = "inline"
)

// Trigger
const (
	TriggerAlways = "always"
	TriggerOnCd   = "on_cd"
)

// KeyStyle
const (
	KeyStyleTree = "tree"
	KeyStyleFlat = "flat"
)

// CwdStyle
const (
	CwdStyleShort    = "short"
	CwdStyleFull     = "full"
	CwdStyleBasename = "basename"
	CwdStyleParent   = "parent"
)

// GitCwdStyle
const (
	GitCwdStyleTree       = "tree"
	GitCwdStyleBreadcrumb = "breadcrumb"
)

// GitStatusStyle
const (
	GitStatusStyleShort = "short"
	GitStatusStyleLong  = "long"
)

// ClaudeMode
const (
	ClaudeModeAuto   = "auto"
	ClaudeModeAlways = "always"
)

// CodexMode
const (
	CodexModeAuto   = "auto"
	CodexModeAlways = "always"
)

// BarStyle
const (
	BarStyleBlock = "block"
	BarStyleDot   = "dot"
	BarStyleFill  = "fill"
)

// TimeStyle
const (
	TimeStyleAbsolute = "absolute"
	TimeStyleRelative = "relative"
)

// Module names
const (
	ModuleCwd    = "cwd"
	ModuleGit    = "git"
	ModuleKube   = "kube"
	ModuleGcp    = "gcp"
	ModuleClaude = "claude"
	ModuleCodex  = "codex"
)

// Theme
const (
	ThemeDefault = "default"
)

type Config struct {
	Theme        string        `yaml:"theme"`
	Format       string        `yaml:"format"`
	Separator    string        `yaml:"separator"`
	Trigger      string        `yaml:"trigger"`   // "always" | "on_cd"
	KeyStyle     string        `yaml:"key_style"` // "flat" | "tree"
	Modules      ModulesConfig `yaml:"modules"`

	// Derived from YAML key order (not a YAML field)
	ModuleOrder    []string            `yaml:"-"`
	SubKeyOrder    map[string][]string `yaml:"-"` // module name → sub-key order
}

type ModulesConfig struct {
	Cwd    CwdConfig    `yaml:"cwd"`
	Git    GitConfig    `yaml:"git"`
	Kube   KubeConfig   `yaml:"kube"`
	Gcp    GcpConfig    `yaml:"gcp"`
	Claude ClaudeConfig `yaml:"claude"`
	Codex  CodexConfig  `yaml:"codex"`
}

type CwdConfig struct {
	Enabled bool   `yaml:"enabled"`
	Style   string `yaml:"style"`
}

type GitConfig struct {
	Enabled   bool           `yaml:"enabled"`
	Indicator bool           `yaml:"indicator"`
	Url       GitUrlConfig   `yaml:"url"`
	Cwd       GitCwdConfig   `yaml:"cwd"`
	Summary   GitSummaryConfig `yaml:"summary"`
	Status    GitStatusConfig `yaml:"status"`
}

type GitUrlConfig struct {
	Enabled bool `yaml:"enabled"`
}

type GitCwdConfig struct {
	Enabled bool   `yaml:"enabled"`
	Style   string `yaml:"style"` // "breadcrumb" | "tree"
}

type GitSummaryConfig struct {
	Symbols GitSymbols `yaml:"symbols"`
}

type GitStatusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Style   string `yaml:"style"` // "short" | "long"
}

type GitSymbols struct {
	Unstaged  string `yaml:"unstaged"`
	Staged    string `yaml:"staged"`
	Stash     string `yaml:"stash"`
	Untracked string `yaml:"untracked"`
	Ahead     string `yaml:"ahead"`
	Behind    string `yaml:"behind"`
}

type KubeConfig struct {
	Enabled bool              `yaml:"enabled"`
	Context KubeContextConfig `yaml:"context"`
}

type KubeContextConfig struct {
	Clean bool `yaml:"clean"` // strip cloud provider prefixes (GKE/EKS/AKS)
}

type GcpConfig struct {
	Enabled bool `yaml:"enabled"`
}

type ClaudeConfig struct {
	Enabled bool               `yaml:"enabled"`
	Mode    string             `yaml:"mode"`   // "always" | "auto"
	Usage   ClaudeUsageConfig  `yaml:"usage"`
	Config  ClaudeConfigView   `yaml:"config"`
}

type ClaudeUsageConfig struct {
	BarStyle  string `yaml:"bar_style"`  // "block" | "dot" | "fill"
	TimeStyle string `yaml:"time_style"` // "absolute" | "relative"
	CacheTTL  int    `yaml:"cache_ttl"`  // seconds
}

type ClaudeConfigView struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"` // "always" | "auto"
}

type CodexConfig struct {
	Enabled bool            `yaml:"enabled"`
	Mode    string          `yaml:"mode"` // "always" | "auto"
	Config  CodexConfigView `yaml:"config"`
}

type CodexConfigView struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"` // "always" | "auto"
}

var DefaultModuleOrder = []string{ModuleCwd, ModuleGit, ModuleKube, ModuleGcp, ModuleClaude, ModuleCodex}

func Default() *Config {
	return &Config{
		Theme:       ThemeDefault,
		Format:      FormatTable,
		Separator:   " │ ",
		Trigger:     TriggerAlways,
		KeyStyle:    KeyStyleTree,
		ModuleOrder: DefaultModuleOrder,
		Modules: ModulesConfig{
			Cwd: CwdConfig{
				Enabled: true,
				Style:   CwdStyleShort,
			},
			Git: GitConfig{
				Enabled:   true,
				Indicator: true,
				Url:       GitUrlConfig{Enabled: true},
				Cwd:       GitCwdConfig{Enabled: true, Style: GitCwdStyleTree},
				Summary:   GitSummaryConfig{Symbols: DefaultGitSymbols()},
				Status:    GitStatusConfig{Enabled: true, Style: GitStatusStyleShort},
			},
			Kube: KubeConfig{
				Enabled: false,
				Context: KubeContextConfig{Clean: true},
			},
			Gcp: GcpConfig{
				Enabled: false,
			},
			Claude: ClaudeConfig{
				Enabled: true,
				Mode:    ClaudeModeAuto,
				Usage: ClaudeUsageConfig{
					BarStyle:  BarStyleBlock,
					TimeStyle: TimeStyleAbsolute,
					CacheTTL:  120,
				},
				Config: ClaudeConfigView{
					Enabled: true,
					Mode:    ClaudeModeAuto,
				},
			},
			Codex: CodexConfig{
				Enabled: true,
				Mode:    CodexModeAuto,
				Config: CodexConfigView{
					Enabled: true,
					Mode:    CodexModeAuto,
				},
			},
		},
	}
}

func DefaultGitSymbols() GitSymbols {
	return GitSymbols{
		Unstaged:  "*",
		Staged:    "+",
		Stash:     "$",
		Untracked: "%",
		Ahead:     "↑",
		Behind:    "↓",
	}
}

func ConfigPath() string {
	// Prefer XDG_CONFIG_HOME, fall back to ~/.config (not ~/Library/Application Support on macOS)
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "enter", "config.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "enter", "config.yaml")
}

func Load(path string) *Config {
	if path == "" {
		path = ConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Default()
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return Default()
	}

	// Extract module and sub-key order from YAML key order
	cfg.ModuleOrder, cfg.SubKeyOrder = extractOrder(data)

	// Validate and normalize config values
	cfg.validate()

	// Fill empty symbols with defaults
	defaults := DefaultGitSymbols()
	sym := &cfg.Modules.Git.Summary.Symbols
	if sym.Unstaged == "" {
		sym.Unstaged = defaults.Unstaged
	}
	if sym.Staged == "" {
		sym.Staged = defaults.Staged
	}
	if sym.Stash == "" {
		sym.Stash = defaults.Stash
	}
	if sym.Untracked == "" {
		sym.Untracked = defaults.Untracked
	}
	if sym.Ahead == "" {
		sym.Ahead = defaults.Ahead
	}
	if sym.Behind == "" {
		sym.Behind = defaults.Behind
	}

	return cfg
}

func (c *Config) validate() {
	d := Default()

	if c.Format != FormatTable && c.Format != FormatInline {
		c.Format = d.Format
	}
	if c.Trigger != TriggerAlways && c.Trigger != TriggerOnCd {
		c.Trigger = d.Trigger
	}
	if c.KeyStyle != KeyStyleFlat && c.KeyStyle != KeyStyleTree {
		c.KeyStyle = d.KeyStyle
	}

	gitCwd := &c.Modules.Git.Cwd
	if gitCwd.Style != GitCwdStyleBreadcrumb && gitCwd.Style != GitCwdStyleTree {
		gitCwd.Style = d.Modules.Git.Cwd.Style
	}
	gitStatus := &c.Modules.Git.Status
	if gitStatus.Style != GitStatusStyleShort && gitStatus.Style != GitStatusStyleLong {
		gitStatus.Style = d.Modules.Git.Status.Style
	}

	cl := &c.Modules.Claude
	if cl.Mode != ClaudeModeAlways && cl.Mode != ClaudeModeAuto {
		cl.Mode = d.Modules.Claude.Mode
	}
	usage := &cl.Usage
	if usage.BarStyle != BarStyleBlock && usage.BarStyle != BarStyleDot && usage.BarStyle != BarStyleFill {
		usage.BarStyle = d.Modules.Claude.Usage.BarStyle
	}
	if usage.TimeStyle != TimeStyleAbsolute && usage.TimeStyle != TimeStyleRelative {
		usage.TimeStyle = d.Modules.Claude.Usage.TimeStyle
	}
	if usage.CacheTTL <= 0 {
		usage.CacheTTL = d.Modules.Claude.Usage.CacheTTL
	}

	cx := &c.Modules.Codex
	if cx.Mode != CodexModeAlways && cx.Mode != CodexModeAuto {
		cx.Mode = d.Modules.Codex.Mode
	}
}

// extractOrder parses the YAML with goccy/go-yaml MapSlice
// to extract both module order and sub-key order per module.
func extractOrder(data []byte) ([]string, map[string][]string) {
	var raw yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &raw, yaml.UseOrderedMap()); err != nil {
		return DefaultModuleOrder, nil
	}

	// Find "modules" key
	for _, item := range raw {
		if key, ok := item.Key.(string); ok && key == "modules" {
			if modules, ok := item.Value.(yaml.MapSlice); ok {
				var moduleOrder []string
				subKeyOrder := make(map[string][]string)

				for _, m := range modules {
					name, ok := m.Key.(string)
					if !ok {
						continue
					}
					moduleOrder = append(moduleOrder, name)

					// Extract sub-key order from this module's keys
					if modValue, ok := m.Value.(yaml.MapSlice); ok {
						var subKeys []string
						for _, field := range modValue {
							if fieldName, ok := field.Key.(string); ok {
								subKeys = append(subKeys, fieldName)
							}
						}
						if len(subKeys) > 0 {
							subKeyOrder[name] = subKeys
						}
					}
				}

				if len(moduleOrder) > 0 {
					// Append any default modules not in config
					seen := make(map[string]bool)
					for _, name := range moduleOrder {
						seen[name] = true
					}
					for _, name := range DefaultModuleOrder {
						if !seen[name] {
							moduleOrder = append(moduleOrder, name)
						}
					}
					return moduleOrder, subKeyOrder
				}
			}
		}
	}

	return DefaultModuleOrder, nil
}

func GenerateDefault() string {
	return `theme: "default"
format: "table"             # table | inline
separator: " │ "
trigger: "always"           # always | on_cd
key_style: "tree"           # flat (git.summary) | tree (├── summary)

modules:
  cwd:
    enabled: true
    style: "short"        # parent | full | short | basename

  git:
    enabled: true
    indicator: true         # show "not a git repo" outside repos
    url:
      enabled: true
    cwd:
      enabled: true
      style: "tree"         # breadcrumb | tree
    summary:
      symbols:
        unstaged: "*"
        staged: "+"
        stash: "$"
        untracked: "%"
        ahead: "↑"
        behind: "↓"
    status:
      enabled: true
      style: "short"        # short | long

  kube:
    enabled: false
    context:
      clean: true           # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false

  claude:
    enabled: true
    mode: "auto"            # always | auto
    usage:
      bar_style: "block"    # block (▰▱) | dot (●○) | fill (█░)
      time_style: "absolute" # absolute (3:00pm) | relative (22m left)
      cache_ttl: 120        # cache duration in seconds
    config:
      enabled: true
      mode: "auto"          # always (show ✓/✗) | auto (show existing only)

  codex:
    enabled: true
    mode: "auto"            # always | auto
    config:
      enabled: true
      mode: "auto"          # always (show ✓/✗) | auto (show existing only)
`
}
