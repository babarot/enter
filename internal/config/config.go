package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Theme     string        `yaml:"theme"`
	Format    string        `yaml:"format"`
	Separator string        `yaml:"separator"`
	Modules   ModulesConfig `yaml:"modules"`
}

type ModulesConfig struct {
	Pwd  PwdConfig  `yaml:"pwd"`
	Git  GitConfig  `yaml:"git"`
	Kube KubeConfig `yaml:"kube"`
	Gcp  GcpConfig  `yaml:"gcp"`
}

type PwdConfig struct {
	Enabled bool   `yaml:"enabled"`
	Style   string `yaml:"style"`
}

type GitConfig struct {
	Enabled bool       `yaml:"enabled"`
	Symbols GitSymbols `yaml:"symbols"`
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
	Enabled bool   `yaml:"enabled"`
	Symbol  string `yaml:"symbol"`
}

type GcpConfig struct {
	Enabled bool   `yaml:"enabled"`
	Symbol  string `yaml:"symbol"`
}

func Default() *Config {
	return &Config{
		Theme:     "default",
		Format:    "inline",
		Separator: " │ ",
		Modules: ModulesConfig{
			Pwd: PwdConfig{
				Enabled: true,
				Style:   "parent",
			},
			Git: GitConfig{
				Enabled: true,
				Symbols: DefaultGitSymbols(),
			},
			Kube: KubeConfig{
				Enabled: false,
				Symbol:  "⎈",
			},
			Gcp: GcpConfig{
				Enabled: false,
				Symbol:  "☁",
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
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "enter", "config.yaml")
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

func GenerateDefault() string {
	return `theme: "default"
format: "inline"            # inline | table | compact
separator: " │ "

modules:
  pwd:
    enabled: true
    style: "parent"       # parent | full | short | basename

  git:
    enabled: true
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

  gcp:
    enabled: false
    # symbol: "☁"
`
}
