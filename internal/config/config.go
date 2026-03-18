package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Theme        string        `yaml:"theme"`
	Format       string        `yaml:"format"`
	Separator    string        `yaml:"separator"`
	Trigger      string        `yaml:"trigger"`   // "always" | "on_cd"
	KeyStyle     string        `yaml:"key_style"` // "flat" | "tree"
	Modules      ModulesConfig `yaml:"modules"`

	// Derived from YAML key order (not a YAML field)
	ModuleOrder []string `yaml:"-"`
}

type ModulesConfig struct {
	Cwd    CwdConfig    `yaml:"cwd"`
	Git    GitConfig    `yaml:"git"`
	Kube   KubeConfig   `yaml:"kube"`
	Gcp    GcpConfig    `yaml:"gcp"`
	Claude ClaudeConfig `yaml:"claude"`
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
	Sign      GitSignConfig  `yaml:"sign"`
	Status    GitStatusConfig `yaml:"status"`
}

type GitUrlConfig struct {
	Enabled bool `yaml:"enabled"`
}

type GitCwdConfig struct {
	Enabled bool   `yaml:"enabled"`
	Style   string `yaml:"style"` // "breadcrumb" | "tree"
}

type GitSignConfig struct {
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
	Enabled      bool     `yaml:"enabled"`
	CleanContext bool     `yaml:"clean_context"` // strip cloud provider prefixes from context name
	Order        []string `yaml:"order"`
}

type GcpConfig struct {
	Enabled bool     `yaml:"enabled"`
	Order   []string `yaml:"order"`
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

var DefaultModuleOrder = []string{"cwd", "git", "kube", "gcp", "claude"}

func Default() *Config {
	return &Config{
		Theme:       "default",
		Format:      "table",
		Separator:   " │ ",
		Trigger:     "always",
		KeyStyle:    "tree",
		ModuleOrder: DefaultModuleOrder,
		Modules: ModulesConfig{
			Cwd: CwdConfig{
				Enabled: true,
				Style:   "short",
			},
			Git: GitConfig{
				Enabled:   true,
				Indicator: true,
				Url:       GitUrlConfig{Enabled: true},
				Cwd:       GitCwdConfig{Enabled: true, Style: "tree"},
				Sign:      GitSignConfig{Symbols: DefaultGitSymbols()},
				Status:    GitStatusConfig{Enabled: true, Style: "short"},
			},
			Kube: KubeConfig{
				Enabled:      false,
				CleanContext: true,
			},
			Gcp: GcpConfig{
				Enabled: false,
			},
			Claude: ClaudeConfig{
				Enabled: true,
				Mode:    "auto",
				Usage: ClaudeUsageConfig{
					BarStyle:  "block",
					TimeStyle: "absolute",
					CacheTTL:  120,
				},
				Config: ClaudeConfigView{
					Enabled: true,
					Mode:    "auto",
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

	// Extract module order from YAML key order using goccy/go-yaml
	cfg.ModuleOrder = extractModuleOrder(data)

	// Validate and normalize config values
	cfg.validate()

	// Fill empty symbols with defaults
	defaults := DefaultGitSymbols()
	sym := &cfg.Modules.Git.Sign.Symbols
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

	if c.Format != "table" && c.Format != "inline" {
		c.Format = d.Format
	}
	if c.Trigger != "always" && c.Trigger != "on_cd" {
		c.Trigger = d.Trigger
	}
	if c.KeyStyle != "flat" && c.KeyStyle != "tree" {
		c.KeyStyle = d.KeyStyle
	}

	gitCwd := &c.Modules.Git.Cwd
	if gitCwd.Style != "breadcrumb" && gitCwd.Style != "tree" {
		gitCwd.Style = d.Modules.Git.Cwd.Style
	}
	gitStatus := &c.Modules.Git.Status
	if gitStatus.Style != "short" && gitStatus.Style != "long" {
		gitStatus.Style = d.Modules.Git.Status.Style
	}

	cl := &c.Modules.Claude
	if cl.Mode != "always" && cl.Mode != "auto" {
		cl.Mode = d.Modules.Claude.Mode
	}
	usage := &cl.Usage
	if usage.BarStyle != "block" && usage.BarStyle != "dot" && usage.BarStyle != "fill" {
		usage.BarStyle = d.Modules.Claude.Usage.BarStyle
	}
	if usage.TimeStyle != "absolute" && usage.TimeStyle != "relative" {
		usage.TimeStyle = d.Modules.Claude.Usage.TimeStyle
	}
	if usage.CacheTTL <= 0 {
		usage.CacheTTL = d.Modules.Claude.Usage.CacheTTL
	}
}

// extractModuleOrder parses the YAML with goccy/go-yaml MapSlice
// to extract the key order of the "modules" section.
func extractModuleOrder(data []byte) []string {
	var raw yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &raw, yaml.UseOrderedMap()); err != nil {
		return DefaultModuleOrder
	}

	// Find "modules" key
	for _, item := range raw {
		if key, ok := item.Key.(string); ok && key == "modules" {
			if modules, ok := item.Value.(yaml.MapSlice); ok {
				var order []string
				for _, m := range modules {
					if name, ok := m.Key.(string); ok {
						order = append(order, name)
					}
				}
				if len(order) > 0 {
					// Append any default modules not in config
					seen := make(map[string]bool)
					for _, name := range order {
						seen[name] = true
					}
					for _, name := range DefaultModuleOrder {
						if !seen[name] {
							order = append(order, name)
						}
					}
					return order
				}
			}
		}
	}

	return DefaultModuleOrder
}

func GenerateDefault() string {
	return `theme: "default"
format: "table"             # table | inline
separator: " │ "
trigger: "always"           # always | on_cd
key_style: "tree"           # flat (git.sign) | tree (├── sign)

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
    sign:
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
    clean_context: true     # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false
    # order: [project, account, region, config]  # sub-key display order

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
`
}
