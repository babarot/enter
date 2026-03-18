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
	Enabled       bool       `yaml:"enabled"`
	ShowRepo      bool       `yaml:"show_repo"`
	ShowIndicator bool       `yaml:"show_indicator"`
	ShowTree      bool       `yaml:"show_tree"`
	ShowStatus    bool       `yaml:"show_status"`
	TreeStyle     string     `yaml:"tree_style"`   // "breadcrumb" | "tree"
	StatusStyle   string     `yaml:"status_style"`  // "short" | "long"
	Order         []string   `yaml:"order"`
	Symbols       GitSymbols `yaml:"symbols"`
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
	Symbol       string   `yaml:"symbol"`
	CleanContext bool     `yaml:"clean_context"` // strip cloud provider prefixes from context name
	Order        []string `yaml:"order"`
}

type GcpConfig struct {
	Enabled bool   `yaml:"enabled"`
	Symbol  string `yaml:"symbol"`
}

type ClaudeConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Mode       string            `yaml:"mode"`         // "always" | "auto"
	BarStyle   string            `yaml:"bar_style"`    // "block" | "dot" | "fill"
	TimeStyle  string            `yaml:"time_style"`   // "absolute" | "relative"
	CacheTTL   int               `yaml:"cache_ttl"`    // seconds
	Order      []string          `yaml:"order"`
	ConfigView ClaudeConfigView  `yaml:"config_view"`
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
				Enabled:       true,
				ShowRepo:      true,
				ShowIndicator: true,
				ShowTree:      true,
				ShowStatus:    true,
				TreeStyle:     "tree",
				StatusStyle:   "short",
				Symbols:       DefaultGitSymbols(),
			},
			Kube: KubeConfig{
				Enabled:      false,
				Symbol:       "⎈",
				CleanContext: true,
			},
			Gcp: GcpConfig{
				Enabled: false,
				Symbol:  "☁",
			},
			Claude: ClaudeConfig{
				Enabled:   true,
				Mode:      "auto",
				BarStyle:  "block",
				TimeStyle: "absolute",
				CacheTTL:  120,
				ConfigView: ClaudeConfigView{
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
	if cfg.Modules.Git.Symbols.Unstaged == "" {
		cfg.Modules.Git.Symbols.Unstaged = defaults.Unstaged
	}
	if cfg.Modules.Git.Symbols.Staged == "" {
		cfg.Modules.Git.Symbols.Staged = defaults.Staged
	}
	if cfg.Modules.Git.Symbols.Stash == "" {
		cfg.Modules.Git.Symbols.Stash = defaults.Stash
	}
	if cfg.Modules.Git.Symbols.Untracked == "" {
		cfg.Modules.Git.Symbols.Untracked = defaults.Untracked
	}
	if cfg.Modules.Git.Symbols.Ahead == "" {
		cfg.Modules.Git.Symbols.Ahead = defaults.Ahead
	}
	if cfg.Modules.Git.Symbols.Behind == "" {
		cfg.Modules.Git.Symbols.Behind = defaults.Behind
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

	git := &c.Modules.Git
	if git.TreeStyle != "breadcrumb" && git.TreeStyle != "tree" {
		git.TreeStyle = d.Modules.Git.TreeStyle
	}
	if git.StatusStyle != "short" && git.StatusStyle != "long" {
		git.StatusStyle = d.Modules.Git.StatusStyle
	}

	cl := &c.Modules.Claude
	if cl.Mode != "always" && cl.Mode != "auto" {
		cl.Mode = d.Modules.Claude.Mode
	}
	if cl.BarStyle != "block" && cl.BarStyle != "dot" && cl.BarStyle != "fill" {
		cl.BarStyle = d.Modules.Claude.BarStyle
	}
	if cl.TimeStyle != "absolute" && cl.TimeStyle != "relative" {
		cl.TimeStyle = d.Modules.Claude.TimeStyle
	}
	if cl.CacheTTL <= 0 {
		cl.CacheTTL = d.Modules.Claude.CacheTTL
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
    # order: [url, cwd, sign, status]  # sub-key display order
    show_repo: true         # show repository URL
    show_indicator: true    # show whether in a git repo
    show_tree: true         # show current position in repo
    show_status: true       # show git status output
    tree_style: "tree"      # breadcrumb | tree
    status_style: "short"   # short | long
    symbols:
      unstaged: "*"
      staged: "+"
      stash: "$"
      untracked: "%"
      ahead: "↑"
      behind: "↓"

  kube:
    enabled: false
    # symbol: "⎈"
    clean_context: true     # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false
    # symbol: "☁"

  claude:
    enabled: true
    # order: [usage, config]  # sub-key display order
    mode: "auto"            # always | auto
    bar_style: "block"      # block (▰▱) | dot (●○) | fill (█░)
    time_style: "absolute"  # absolute (3:00pm) | relative (22m left)
    cache_ttl: 120          # cache duration in seconds
    config_view:
      enabled: true
      mode: "auto"          # always (show ✓/✗) | auto (show existing only)
`
}
