# enter

Show contextual info every time you press Enter in your shell.

```
╭────────────────┬──────────────────────────────────────────╮
│pwd             │ ~/s/g/babarot/project                    │
│git.url         │ https://github.com/babarot/project       │
│git.sign        │ (main *%)                                │
│git.cwd         │ /                                        │
│git.status      │ M  internal/modules/git.go               │
│claude.usage.5h │ ▰▱▱▱▱▱▱▱▱▱  14% ⟳ 3:00pm               │
│claude.usage.7d │ ▰▱▱▱▱▱▱▱▱▱  14% ⟳ Mar 19, 2:00pm       │
│claude.config   │ ✓ CLAUDE.md  ✓ rules (3)  ✓ skills (2)  │
╰────────────────┴──────────────────────────────────────────╯
```

## Install

```bash
go install github.com/babarot/enter/cmd/enter@latest
```

Or build from source:

```bash
git clone https://github.com/babarot/enter.git
cd enter
go build -o enter ./cmd/enter/
```

## Shell Integration

Add to your shell config:

**zsh** (`~/.zshrc`):

```zsh
eval "$(enter --init-shell zsh)"
```

**bash** (`~/.bashrc`):

```bash
eval "$(enter --init-shell bash)"
```

Only fires on **empty Enter** (pressing Enter with no command). Running commands like `ls` or `git status` will not trigger the display.

How it works:
- **zsh**: Uses `bindkey '^m'` with a custom widget + `precmd` hook. The widget sets a flag on empty input, and `precmd` runs `enter` only when the flag is set. This avoids overriding `accept-line` directly, preventing conflicts with other plugins (fzf-tab, etc.).
- **bash**: Uses `DEBUG` trap + `PROMPT_COMMAND` to detect whether a command was entered.

## CLI Flags

```
--format <table|inline>           Display format (overrides config)
--theme <name>                    Color theme (overrides config)
--config <path>                   Path to config file
--last-pwd <path>                 Previous directory (for trigger: on_cd)
--init-shell <zsh|bash>           Print shell integration snippet
--init-config                     Generate default config file
--version, -v                     Show version
--help                            Show help
```

## Configuration

Config file location: `~/.config/enter/config.yaml`

Generate a default config:

```bash
enter --init-config
```

### Full Config Reference

```yaml
# Color theme: default | tokyo-night | catppuccin | dracula | nord | gruvbox
theme: "default"

# Display format:
#   table  - bordered table with labeled rows (default)
#   inline - single line with separator
format: "table"

# Separator between segments in inline format
separator: " │ "

# When to show output on empty Enter:
#   always - every empty Enter (default)
#   on_cd  - only when directory changed
trigger: "always"

# Module display order is determined by the key order in this file.
# Reorder the module sections below to change the display order.
# Modules not listed here are appended in default order.
modules:
  # ── pwd ──────────────────────────────────────────
  pwd:
    enabled: true

    # Path display style:
    #   short    - "~/s/g/babarot/enter" (abbreviated, keep last 2)
    #   parent   - "babarot/enter" (parent + basename)
    #   full     - "~/src/github.com/babarot/enter"
    #   basename - "enter"
    style: "short"

  # ── git ──────────────────────────────────────────
  git:
    enabled: true

    # Show the repository HTTPS URL (parsed from remote origin)
    # Supports: git@, ssh://, https:// remote formats
    # Table key: git.url
    show_repo: true

    # Show "not a git repo" when outside a git repository
    show_indicator: true

    # Show current position within the repository
    # Table key: git.cwd
    # At repo root: "/"
    # In subdirectory: depends on tree_style
    show_tree: true

    # Show git status output
    # Table key: git.status
    # Color-coded: M=yellow, A=green, D=red, ??=muted (short)
    #              Section-based coloring (long)
    show_status: true

    # How to display the repo position:
    #   tree       - tree visualization with └── branches
    #   breadcrumb - "/enter → cmd → enter"
    tree_style: "tree"

    # Git status output format:
    #   short - "M  file.go" (git status --short)
    #   long  - full git status output with sections
    status_style: "short"

    # Customize git status symbols
    symbols:
      unstaged: "*"     # Unstaged changes
      staged: "+"       # Staged changes
      stash: "$"        # Stash entries exist
      untracked: "%"    # Untracked files
      ahead: "↑"        # Commits ahead of upstream
      behind: "↓"       # Commits behind upstream

  # ── kube ─────────────────────────────────────────
  kube:
    enabled: false

    # Symbol displayed before the context name
    symbol: "⎈"

    # Reads from $KUBECONFIG (colon-separated, multiple files)
    # Falls back to ~/.kube/config

  # ── gcp ──────────────────────────────────────────
  gcp:
    enabled: false

    # Symbol displayed before the project name
    symbol: "☁"

    # Resolution order:
    # 1. $CLOUDSDK_CORE_PROJECT
    # 2. ~/.config/gcloud/properties [core] project
    # 3. ~/.config/gcloud/active_config → config_{name}

  # ── claude ───────────────────────────────────────
  claude:
    enabled: true

    # When to show Claude Code usage:
    #   auto   - show when .claude/ or CLAUDE.md exists (in cwd or git root)
    #   always - always show
    mode: "auto"

    # Progress bar style:
    #   block - ▰▱ (default)
    #   dot   - ●○
    #   fill  - █░
    bar_style: "block"

    # Reset time display:
    #   absolute - "3:00pm", "Mar 19, 2:00pm"
    #   relative - "22m left", "3h 15m left"
    time_style: "absolute"

    # API response cache duration in seconds
    cache_ttl: 120

    # OAuth token resolution order:
    # 1. $CLAUDE_CODE_OAUTH_TOKEN
    # 2. macOS Keychain (Claude Code-credentials)
    # 3. ~/.claude/.credentials.json

    # Show Claude Code project configuration status
    config_view:
      enabled: true

      # Display mode:
      #   auto   - show only existing items with ✓
      #   always - show all items with ✓/✗
      mode: "auto"

    # Checked items:
    # CLAUDE.md, .claude/settings.json, .claude/settings.local.json,
    # .claude/rules/, .claude/skills/, .claude/agents/,
    # .claude/commands/, .mcp.json
```

### Display Formats

**table** (default):

```
╭────────────────┬──────────────────────────────────────────╮
│pwd             │ ~/s/g/babarot/project                    │
│git.url         │ https://github.com/babarot/project       │
│git.sign        │ (main *%)                                │
│git.cwd         │ /                                        │
│git.status      │ ╭──────────────────────────╮             │
│                │ │ M  internal/modules/git.go│             │
│                │ ╰──────────────────────────╯             │
│claude.usage.5h │ ▰▱▱▱▱▱▱▱▱▱  14% ⟳ 3:00pm               │
│claude.usage.7d │ ▰▱▱▱▱▱▱▱▱▱  14% ⟳ Mar 19, 2:00pm       │
│claude.config   │ ╭───────────────────────────╮            │
│                │ │ ✓ CLAUDE.md               │            │
│                │ │ ✓ .claude/settings.json   │            │
│                │ │ ✓ .claude/rules (3)       │            │
│                │ │ ✓ .claude/skills (2)      │            │
│                │ ╰───────────────────────────╯            │
╰────────────────┴──────────────────────────────────────────╯
```

Multiline values (git.status, git.cwd tree, claude.config) are automatically wrapped in nested boxes.

**inline**:

```
~/s/g/babarot/enter │ (main *%) │ current ▰▱▱▱▱▱▱▱▱▱ 14% ⟳ 3:00pm | weekly ▰▱▱▱▱▱▱▱▱▱ 14% ⟳ Mar 19, 2:00pm
```

### Table Row Keys

| Key | Source | Description |
|-----|--------|-------------|
| `pwd` | pwd module | Current directory |
| `git.url` | `show_repo: true` | Repository HTTPS URL |
| `git.sign` | always (when in repo) | Branch, flags, ahead/behind, operation |
| `git.cwd` | `show_tree: true` | Position in repo (breadcrumb or tree) |
| `git.status` | `show_status: true` | git status output (short or long) |
| `kube` | kube module | Kubernetes current-context |
| `gcp` | gcp module | GCP project |
| `claude.usage.5h` | claude module | 5-hour rolling window utilization |
| `claude.usage.7d` | claude module | 7-day rolling window utilization |
| `claude.config` | `config_view.enabled: true` | Project config status (CLAUDE.md, rules, skills, etc.) |

### Themes

Six built-in themes:

| Theme | Description |
|-------|-------------|
| `default` | Balanced palette for dark terminals |
| `tokyo-night` | Based on Tokyo Night color scheme |
| `catppuccin` | Based on Catppuccin Mocha |
| `dracula` | Based on Dracula theme |
| `nord` | Based on Nord color scheme |
| `gruvbox` | Based on Gruvbox dark |

### Module Order

The display order of modules is determined by the key order in the config file. Simply reorder the module sections to change the display order:

```yaml
modules:
  claude:    # ← shown first
    enabled: true
  git:       # ← shown second
    enabled: true
  pwd:       # ← shown third
    enabled: true
```

Modules not listed in the config file are appended in default order (`pwd`, `git`, `kube`, `gcp`, `claude`). No separate `order` field is needed.

### Git Symbols

The `symbols` map in the git config customizes the status indicators shown in `git.sign`:

```
(main *+$% ↑1↓2|REBASE 3/5)
       ││││ │ │  └── in-progress operation
       ││││ │ └── commits behind upstream
       ││││ └── commits ahead of upstream
       │││└── untracked files
       ││└── stash entries
       │└── staged changes
       └── unstaged changes
```

## Architecture

```
cmd/enter/main.go        CLI, parallel execution (goroutine + WaitGroup)
internal/
├── config/config.go     YAML config loading + defaults
├── module/module.go     Module interface, Segment, SemanticColor
├── render/
│   ├── render.go        Output formatting (table/inline)
│   └── theme.go         Color themes
└── modules/
    ├── pwd.go           Current directory
    ├── git.go           Git status, repo URL, tree, status output
    ├── kube.go          Kubernetes context
    ├── gcp.go           GCP project
    └── claude.go        Claude Code API usage + project config status
```

All modules run in parallel using goroutines. Each module returns `nil` if disabled or has nothing to display.

## License

MIT
