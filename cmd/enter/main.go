package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
	"github.com/babarot/enter/internal/modules"
	"github.com/babarot/enter/internal/render"
)

var version = "0.1.0"

func main() {
	var (
		initShell  string
		initConfig bool
		configPath string
		format     string
		theme      string
		showVer    bool
	)

	flag.StringVar(&initShell, "init-shell", "", "Print shell integration snippet (zsh|bash)")
	flag.BoolVar(&initConfig, "init-config", false, "Generate default config file")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&format, "format", "", "Display format (inline|table|compact)")
	flag.StringVar(&theme, "theme", "", "Color theme")
	flag.BoolVar(&showVer, "version", false, "Show version")
	flag.BoolVar(&showVer, "v", false, "Show version")
	flag.Parse()

	if showVer {
		fmt.Printf("enter %s\n", version)
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

	ctx := &module.Context{
		Cwd:    cwd,
		Config: cfg,
	}

	// All modules in display order
	allModules := []module.Module{
		&modules.PwdModule{},
		&modules.GitModule{},
		&modules.KubeModule{},
		&modules.GcpModule{},
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
		fmt.Print(`__enter_precmd() {
  enter
}
autoload -Uz add-zsh-hook
add-zsh-hook precmd __enter_precmd
`)
	case "bash":
		fmt.Print(`__enter_precmd() {
  enter
}
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
