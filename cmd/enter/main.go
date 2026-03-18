package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
	"github.com/babarot/enter/internal/modules"
	"github.com/babarot/enter/internal/render"
)

var (
	version  = "dev"
	revision = "HEAD"
)

func main() {
	var (
		initShell  string
		initConfig bool
		editConfig bool
		configPath string
		format     string
		theme      string
		lastDir    string
		showVer    bool
	)

	flag.StringVar(&initShell, "init-shell", "", "Print shell integration snippet (zsh|bash)")
	flag.BoolVar(&initConfig, "init-config", false, "Generate default config file")
	flag.BoolVar(&editConfig, "edit-config", false, "Open config file in $EDITOR")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&format, "format", "", "Display format (table|inline)")
	flag.StringVar(&theme, "theme", "", "Color theme")
	flag.StringVar(&lastDir, "last-dir", "", "Previous working directory (for trigger: on_cd)")
	flag.BoolVar(&showVer, "version", false, "Show version")
	flag.BoolVar(&showVer, "v", false, "Show version")
	flag.Parse()

	if showVer {
		fmt.Printf("enter %s (%s)\n", version, revision)
		return
	}

	if initShell != "" {
		printShellInit(initShell)
		return
	}

	if initConfig {
		generateConfig()
		return
	}

	if editConfig {
		openConfigInEditor()
		return
	}

	// Load config
	cfg := config.Load(configPath)

	// CLI flags override config
	if format != "" {
		cfg.Format = format
	}
	if theme != "" {
		cfg.Theme = theme
	}

	cwd, _ := os.Getwd()

	// trigger: on_cd — skip if directory hasn't changed
	if cfg.Trigger == config.TriggerOnCd && lastDir != "" && lastDir == cwd {
		return
	}

	ctx := &module.Context{
		Cwd:    cwd,
		Config: cfg,
	}

	// Module registry
	moduleMap := map[string]module.Module{
		config.ModuleCwd:    &modules.CwdModule{},
		config.ModuleGit:    &modules.GitModule{},
		config.ModuleKube:   &modules.KubeModule{},
		config.ModuleGcp:    &modules.GcpModule{},
		config.ModuleClaude: &modules.ClaudeModule{},
		config.ModuleCodex:  &modules.CodexModule{},
	}

	// Order modules based on config
	var allModules []module.Module
	for _, name := range cfg.ModuleOrder {
		if m, ok := moduleMap[name]; ok {
			allModules = append(allModules, m)
		}
	}

	// Run all modules in parallel
	results := make([]*module.Output, len(allModules))
	var wg sync.WaitGroup
	for i, m := range allModules {
		wg.Add(1)
		go func(idx int, mod module.Module) {
			defer wg.Done()
			results[idx] = mod.Run(ctx)
		}(i, m)
	}
	wg.Wait()

	// Filter nil results
	var outputs []*module.Output
	for _, r := range results {
		if r != nil {
			outputs = append(outputs, r)
		}
	}

	line := render.Render(outputs, cfg)
	if line != "" {
		fmt.Println(line)
	}
}

func printShellInit(shell string) {
	switch shell {
	case "zsh":
		fmt.Print(`__enter_flag=false
__enter_last_dir="$PWD"
__enter_widget() {
  if [[ -z "$BUFFER" ]]; then
    __enter_flag=true
  fi
  zle accept-line
}
__enter_precmd() {
  if $__enter_flag; then
    __enter_flag=false
    enter --last-dir="$__enter_last_dir"
    __enter_last_dir="$PWD"
  fi
}
zle -N __enter_widget
bindkey '^m' __enter_widget
autoload -Uz add-zsh-hook
add-zsh-hook precmd __enter_precmd
`)
	case "bash":
		fmt.Print(`__enter_prev_cmd=""
__enter_last_dir="$PWD"
__enter_preexec() { __enter_prev_cmd="$1"; }
__enter_precmd() {
  if [[ -z "$__enter_prev_cmd" ]]; then
    enter --last-dir="$__enter_last_dir"
    __enter_last_dir="$PWD"
  fi
  __enter_prev_cmd=""
}
trap '__enter_preexec "$BASH_COMMAND"' DEBUG
PROMPT_COMMAND="__enter_precmd;${PROMPT_COMMAND}"
`)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s. Supported: zsh, bash\n", shell)
		os.Exit(1)
	}
}

func generateConfig() {
	path := config.ConfigPath()

	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "Config already exists: %s\n", path)
		os.Exit(1)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, []byte(config.GenerateDefault()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created %s\n", path)
}

func openConfigInEditor() {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Fprintln(os.Stderr, "$EDITOR is not set")
		os.Exit(1)
	}

	path := config.ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, []byte(config.GenerateDefault()), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Created %s\n", path)
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open editor: %v\n", err)
		os.Exit(1)
	}
}
