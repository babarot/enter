package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
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

// StringOrSlice accepts either a single string or a list of strings in YAML.
type StringOrSlice []string

// UnmarshalYAML implements goccy/go-yaml BytesUnmarshaler.
func (s *StringOrSlice) UnmarshalYAML(data []byte) error {
	var single string
	if err := yaml.Unmarshal(data, &single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	var multi []string
	if err := yaml.Unmarshal(data, &multi); err != nil {
		return err
	}
	*s = StringOrSlice(multi)
	return nil
}

// When defines conditions for when a module should be displayed.
type When struct {
	Dir StringOrSlice `yaml:"dir"`
}

// Match returns true if cwd matches any of the dir patterns.
// Returns true if When is nil or Dir is empty (no restriction).
func (w *When) Match(cwd string) bool {
	if w == nil || len(w.Dir) == 0 {
		return true
	}
	home, _ := os.UserHomeDir()
	for _, pattern := range w.Dir {
		if strings.HasPrefix(pattern, "~/") {
			pattern = filepath.Join(home, pattern[2:])
		}
		if matched, _ := doublestar.Match(pattern, cwd); matched {
			return true
		}
	}
	return false
}

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
	When    *When  `yaml:"when"`
}

type GitConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Indicator bool      `yaml:"indicator"`
	Fields    GitFields `yaml:"fields"`
	When      *When     `yaml:"when"`
}

type GitFields struct {
	Url     Field[GitUrlConfig]     `yaml:"url"`
	Cwd     Field[GitCwdConfig]     `yaml:"cwd"`
	Summary Field[GitSummaryConfig] `yaml:"summary"`
	Status  Field[GitStatusConfig]  `yaml:"status"`
}

type GitUrlConfig struct{}

type GitCwdConfig struct {
	Style string `yaml:"style"` // "breadcrumb" | "tree"
}

type GitSummaryConfig struct {
	Symbols GitSymbols `yaml:"symbols"`
}

type GitStatusConfig struct {
	Style string `yaml:"style"` // "short" | "long"
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
	Enabled bool       `yaml:"enabled"`
	Fields  KubeFields `yaml:"fields"`
	When    *When      `yaml:"when"`
}

type KubeFields struct {
	Context Field[KubeContextConfig] `yaml:"context"`
}

type KubeContextConfig struct {
	Clean bool `yaml:"clean"` // strip cloud provider prefixes (GKE/EKS/AKS)
}

type GcpConfig struct {
	Enabled bool      `yaml:"enabled"`
	Fields  GcpFields `yaml:"fields"`
	When    *When     `yaml:"when"`
}

type GcpFields struct{} // reserved for future per-field config

type ClaudeConfig struct {
	Enabled bool         `yaml:"enabled"`
	Mode    string       `yaml:"mode"` // "always" | "auto"
	Fields  ClaudeFields `yaml:"fields"`
	When    *When        `yaml:"when"`
}

type ClaudeFields struct {
	Usage  Field[ClaudeUsageConfig] `yaml:"usage"`
	Config Field[ClaudeConfigView]  `yaml:"config"`
}

type ClaudeUsageConfig struct {
	BarStyle  string `yaml:"bar_style"`  // "block" | "dot" | "fill"
	TimeStyle string `yaml:"time_style"` // "absolute" | "relative"
	CacheTTL  int    `yaml:"cache_ttl"`  // seconds
}

type ClaudeConfigView struct {
	Mode string `yaml:"mode"` // "always" | "auto"
}

type CodexConfig struct {
	Enabled bool        `yaml:"enabled"`
	Mode    string      `yaml:"mode"` // "always" | "auto"
	Fields  CodexFields `yaml:"fields"`
	When    *When       `yaml:"when"`
}

type CodexFields struct {
	Config Field[CodexConfigView] `yaml:"config"`
}

type CodexConfigView struct {
	Mode string `yaml:"mode"` // "always" | "auto"
}

// WhenFor returns the When condition for the named module, or nil if none is set.
func (mc *ModulesConfig) WhenFor(name string) *When {
	switch name {
	case ModuleCwd:
		return mc.Cwd.When
	case ModuleGit:
		return mc.Git.When
	case ModuleKube:
		return mc.Kube.When
	case ModuleGcp:
		return mc.Gcp.When
	case ModuleClaude:
		return mc.Claude.When
	case ModuleCodex:
		return mc.Codex.When
	default:
		return nil
	}
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
				Fields: GitFields{
					Url:     NewField(GitUrlConfig{}),
					Cwd:     NewField(GitCwdConfig{Style: GitCwdStyleTree}),
					Summary: NewField(GitSummaryConfig{Symbols: DefaultGitSymbols()}),
					Status:  NewField(GitStatusConfig{Style: GitStatusStyleShort}),
				},
			},
			Kube: KubeConfig{
				Enabled: false,
				Fields: KubeFields{
					Context: NewField(KubeContextConfig{Clean: true}),
				},
			},
			Gcp: GcpConfig{
				Enabled: false,
			},
			Claude: ClaudeConfig{
				Enabled: true,
				Mode:    ClaudeModeAuto,
				Fields: ClaudeFields{
					Usage: NewField(ClaudeUsageConfig{
						BarStyle:  BarStyleBlock,
						TimeStyle: TimeStyleAbsolute,
						CacheTTL:  120,
					}),
					Config: NewField(ClaudeConfigView{
						Mode: ClaudeModeAuto,
					}),
				},
			},
			Codex: CodexConfig{
				Enabled: true,
				Mode:    CodexModeAuto,
				Fields: CodexFields{
					Config: NewField(CodexConfigView{
						Mode: CodexModeAuto,
					}),
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

	// Reset all Fields to zero values before Unmarshal so that
	// Field[T].Present() accurately reflects YAML key presence.
	cfg.Modules.Git.Fields = GitFields{}
	cfg.Modules.Kube.Fields = KubeFields{}
	cfg.Modules.Claude.Fields = ClaudeFields{}
	cfg.Modules.Codex.Fields = CodexFields{}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return Default()
	}

	// Extract module order, sub-key order, and which modules have "fields:" in YAML
	var fieldsPresent map[string]bool
	cfg.ModuleOrder, cfg.SubKeyOrder, fieldsPresent = extractOrder(data)

	// Restore default fields for modules that don't have "fields:" in YAML
	defaults := Default()
	if !fieldsPresent[ModuleGit] {
		cfg.Modules.Git.Fields = defaults.Modules.Git.Fields
	} else {
		// Mark empty-key fields as present (YAML unmarshaler is not called for null values)
		ensureFieldsPresent(cfg.SubKeyOrder[ModuleGit], &cfg.Modules.Git.Fields)
	}
	if !fieldsPresent[ModuleKube] {
		cfg.Modules.Kube.Fields = defaults.Modules.Kube.Fields
	} else {
		ensureFieldsPresent(cfg.SubKeyOrder[ModuleKube], &cfg.Modules.Kube.Fields)
	}
	if !fieldsPresent[ModuleClaude] {
		cfg.Modules.Claude.Fields = defaults.Modules.Claude.Fields
	} else {
		ensureFieldsPresent(cfg.SubKeyOrder[ModuleClaude], &cfg.Modules.Claude.Fields)
	}
	if !fieldsPresent[ModuleCodex] {
		cfg.Modules.Codex.Fields = defaults.Modules.Codex.Fields
	} else {
		ensureFieldsPresent(cfg.SubKeyOrder[ModuleCodex], &cfg.Modules.Codex.Fields)
	}

	// Warn about empty fields and deprecated "enabled" key
	warnFieldsConfig(data, fieldsPresent, cfg.SubKeyOrder)

	// Validate and normalize config values
	cfg.validate()

	// Fill empty git symbols with defaults
	if cfg.Modules.Git.Fields.Summary.Present() {
		s := cfg.Modules.Git.Fields.Summary.Get()
		defSym := DefaultGitSymbols()
		if s.Symbols.Unstaged == "" {
			s.Symbols.Unstaged = defSym.Unstaged
		}
		if s.Symbols.Staged == "" {
			s.Symbols.Staged = defSym.Staged
		}
		if s.Symbols.Stash == "" {
			s.Symbols.Stash = defSym.Stash
		}
		if s.Symbols.Untracked == "" {
			s.Symbols.Untracked = defSym.Untracked
		}
		if s.Symbols.Ahead == "" {
			s.Symbols.Ahead = defSym.Ahead
		}
		if s.Symbols.Behind == "" {
			s.Symbols.Behind = defSym.Behind
		}
		cfg.Modules.Git.Fields.Summary.Set(s)
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

	// Git fields
	if c.Modules.Git.Fields.Cwd.Present() {
		cwd := c.Modules.Git.Fields.Cwd.Get()
		if cwd.Style != GitCwdStyleBreadcrumb && cwd.Style != GitCwdStyleTree {
			cwd.Style = d.Modules.Git.Fields.Cwd.Get().Style
		}
		c.Modules.Git.Fields.Cwd.Set(cwd)
	}
	if c.Modules.Git.Fields.Status.Present() {
		st := c.Modules.Git.Fields.Status.Get()
		if st.Style != GitStatusStyleShort && st.Style != GitStatusStyleLong {
			st.Style = d.Modules.Git.Fields.Status.Get().Style
		}
		c.Modules.Git.Fields.Status.Set(st)
	}

	// Claude
	cl := &c.Modules.Claude
	if cl.Mode != ClaudeModeAlways && cl.Mode != ClaudeModeAuto {
		cl.Mode = d.Modules.Claude.Mode
	}
	if cl.Fields.Usage.Present() {
		u := cl.Fields.Usage.Get()
		du := d.Modules.Claude.Fields.Usage.Get()
		if u.BarStyle != BarStyleBlock && u.BarStyle != BarStyleDot && u.BarStyle != BarStyleFill {
			u.BarStyle = du.BarStyle
		}
		if u.TimeStyle != TimeStyleAbsolute && u.TimeStyle != TimeStyleRelative {
			u.TimeStyle = du.TimeStyle
		}
		if u.CacheTTL <= 0 {
			u.CacheTTL = du.CacheTTL
		}
		cl.Fields.Usage.Set(u)
	}
	if cl.Fields.Config.Present() {
		cv := cl.Fields.Config.Get()
		if cv.Mode != ClaudeModeAlways && cv.Mode != ClaudeModeAuto {
			cv.Mode = d.Modules.Claude.Fields.Config.Get().Mode
		}
		cl.Fields.Config.Set(cv)
	}

	// Codex
	cx := &c.Modules.Codex
	if cx.Mode != CodexModeAlways && cx.Mode != CodexModeAuto {
		cx.Mode = d.Modules.Codex.Mode
	}
	if cx.Fields.Config.Present() {
		cxc := cx.Fields.Config.Get()
		if cxc.Mode != CodexModeAlways && cxc.Mode != CodexModeAuto {
			cxc.Mode = d.Modules.Codex.Fields.Config.Get().Mode
		}
		cx.Fields.Config.Set(cxc)
	}
}

// ensureFieldsPresent marks fields as present when the YAML key existed
// but had a null/empty value (where UnmarshalYAML is not called by goccy/go-yaml).
func ensureFieldsPresent(keys []string, fields interface{}) {
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	switch f := fields.(type) {
	case *GitFields:
		if keySet["url"] && !f.Url.Present() {
			f.Url.MarkPresent()
		}
		if keySet["cwd"] && !f.Cwd.Present() {
			f.Cwd.MarkPresent()
		}
		if keySet["summary"] && !f.Summary.Present() {
			f.Summary.MarkPresent()
		}
		if keySet["status"] && !f.Status.Present() {
			f.Status.MarkPresent()
		}
	case *KubeFields:
		if keySet["context"] && !f.Context.Present() {
			f.Context.MarkPresent()
		}
	case *ClaudeFields:
		if keySet["usage"] && !f.Usage.Present() {
			f.Usage.MarkPresent()
		}
		if keySet["config"] && !f.Config.Present() {
			f.Config.MarkPresent()
		}
	case *CodexFields:
		if keySet["config"] && !f.Config.Present() {
			f.Config.MarkPresent()
		}
	}
}

// warnFieldsConfig emits warnings to stderr for deprecated or empty fields config.
func warnFieldsConfig(data []byte, fieldsPresent map[string]bool, subKeyOrder map[string][]string) {
	// Warn about empty fields
	for mod, present := range fieldsPresent {
		if present {
			if keys, ok := subKeyOrder[mod]; !ok || len(keys) == 0 {
				fmt.Fprintf(os.Stderr, "warning: module %q has empty fields\n", mod)
			}
		}
	}

	// Warn about deprecated "enabled" key inside field configs
	var raw yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &raw, yaml.UseOrderedMap()); err != nil {
		return
	}
	for _, item := range raw {
		if key, ok := item.Key.(string); ok && key == "modules" {
			if modules, ok := item.Value.(yaml.MapSlice); ok {
				for _, m := range modules {
					modName, ok := m.Key.(string)
					if !ok {
						continue
					}
					modValue, ok := m.Value.(yaml.MapSlice)
					if !ok {
						continue
					}
					for _, field := range modValue {
						fieldName, ok := field.Key.(string)
						if !ok || fieldName != "fields" {
							continue
						}
						fieldsValue, ok := field.Value.(yaml.MapSlice)
						if !ok {
							continue
						}
						for _, f := range fieldsValue {
							fName, ok := f.Key.(string)
							if !ok {
								continue
							}
							if fValue, ok := f.Value.(yaml.MapSlice); ok {
								for _, kv := range fValue {
									if kvKey, ok := kv.Key.(string); ok && kvKey == "enabled" {
										fmt.Fprintf(os.Stderr, "warning: %q field in module %q uses deprecated 'enabled' key (ignored, presence of key controls visibility now)\n", fName, modName)
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

// extractOrder parses the YAML with goccy/go-yaml MapSlice
// to extract module order, sub-key order per module, and which modules
// have an explicit "fields" key in the YAML.
func extractOrder(data []byte) ([]string, map[string][]string, map[string]bool) {
	var raw yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &raw, yaml.UseOrderedMap()); err != nil {
		return DefaultModuleOrder, nil, nil
	}

	// Find "modules" key
	for _, item := range raw {
		if key, ok := item.Key.(string); ok && key == "modules" {
			if modules, ok := item.Value.(yaml.MapSlice); ok {
				var moduleOrder []string
				subKeyOrder := make(map[string][]string)
				fieldsPresent := make(map[string]bool)

				for _, m := range modules {
					name, ok := m.Key.(string)
					if !ok {
						continue
					}
					moduleOrder = append(moduleOrder, name)

					// Extract sub-key order from this module's "fields" key,
					// or fall back to top-level keys for single-key modules.
					if modValue, ok := m.Value.(yaml.MapSlice); ok {
						var subKeys []string
						foundFields := false
						for _, field := range modValue {
							fieldName, ok := field.Key.(string)
							if !ok {
								continue
							}
							if fieldName == "fields" {
								fieldsPresent[name] = true
								if fieldsValue, ok := field.Value.(yaml.MapSlice); ok {
									for _, f := range fieldsValue {
										if fk, ok := f.Key.(string); ok {
											subKeys = append(subKeys, fk)
										}
									}
								}
								foundFields = true
								break
							}
						}
						if !foundFields {
							for _, field := range modValue {
								if fieldName, ok := field.Key.(string); ok {
									subKeys = append(subKeys, fieldName)
								}
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
					return moduleOrder, subKeyOrder, fieldsPresent
				}
			}
		}
	}

	return DefaultModuleOrder, nil, nil
}

func GenerateDefault() string {
	return `theme: "default"
format: "table"             # table | inline
separator: " │ "
trigger: "always"           # always | on_cd
key_style: "tree"           # flat (git.summary) | tree (├── summary)

modules:
  # Each module supports "when" for conditional display:
  #   when:
  #     dir: "~/src/github.com/mycompany/**"   # single glob pattern
  # Multiple patterns (OR):
  #   when:
  #     dir:
  #       - "~/src/github.com/mycompany/**"
  #       - "~/work/infra/**"

  cwd:
    enabled: true
    style: "short"        # parent | full | short | basename

  git:
    enabled: true
    indicator: true         # show "not a git repo" outside repos
    fields:                 # list fields to display (order matters, omit to hide)
      url:
      cwd:
        style: "tree"       # breadcrumb | tree
      summary:
        symbols:
          unstaged: "*"
          staged: "+"
          stash: "$"
          untracked: "%"
          ahead: "↑"
          behind: "↓"
      status:
        style: "short"      # short | long

  kube:
    enabled: false
    fields:
      context:
        clean: true         # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false
    # when:
    #   dir: "~/src/github.com/mycompany/**"

  claude:
    enabled: true
    mode: "auto"            # always | auto
    fields:                 # list fields to display (order matters, omit to hide)
      usage:
        bar_style: "block"  # block (▰▱) | dot (●○) | fill (█░)
        time_style: "absolute" # absolute (3:00pm) | relative (22m left)
        cache_ttl: 120      # cache duration in seconds
      config:
        mode: "auto"        # always (show ✓/✗) | auto (show existing only)

  codex:
    enabled: true
    mode: "auto"            # always | auto
    fields:                 # list fields to display (order matters, omit to hide)
      config:
        mode: "auto"        # always (show ✓/✗) | auto (show existing only)
`
}
