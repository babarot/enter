package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/babarot/enter/internal/config"
	"github.com/babarot/enter/internal/module"
	"github.com/babarot/enter/internal/modules"
	"github.com/babarot/enter/internal/render"
)

func main() {
	// Simple flag parsing (no external dependency needed for now)
	args := os.Args[1:]

	for i, arg := range args {
		switch arg {
		case "--init-shell":
			if i+1 < len(args) {
				printShellInit(args[i+1])
			} else {
				fmt.Fprintln(os.Stderr, "Usage: enter --init-shell <zsh|bash>")
				os.Exit(1)
			}
			return
		case "--init-config":
			generateConfig()
			return
		case "--help", "-h":
			fmt.Println("Usage: enter [options]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --init-shell <zsh|bash>  Print shell integration snippet")
			fmt.Println("  --init-config            Generate default config file")
			fmt.Println("  --config <path>          Path to config file")
			fmt.Println("  --help, -h               Show this help")
			fmt.Println("  --version, -v            Show version")
			return
		case "--version", "-v":
			fmt.Println("enter 0.1.0")
			return
		}
	}

	// Load config
	var configPath string
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			configPath = args[i+1]
		}
	}
	cfg := config.Load(configPath)

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
