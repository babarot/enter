# enter

Show contextual info every time you press Enter in your shell.

```
babarot/enter │ (main *%) │ ⎈ prod-cluster │ ☁ my-project
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

## CLI Flags

```
--format <inline|table|compact>   Display format (overrides config)
--theme <name>                    Color theme (overrides config)
--config <path>                   Path to config file
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
#   inline  - single line with separator (default)
#   table   - bordered table with labeled rows
#   compact - multi-line with labels, no border
format: "inline"

# Separator between segments in inline format
separator: " │ "

modules:
  # ── pwd ──────────────────────────────────────────
  pwd:
    enabled: true

    # Path display style:
    #   parent   - "babarot/enter" (parent + basename)
    #   full     - "~/src/github.com/babarot/enter"
    #   short    - "~/s/g/b/enter" (abbreviated)
    #   basename - "enter"
    style: "parent"

  # ── git ──────────────────────────────────────────
  git:
    enabled: true

    # Show the repository HTTPS URL (parsed from remote origin)
    # Supports: git@, ssh://, https:// remote formats
    # Table/compact key: git.url
    show_repo: false

    # Show "not a git repo" when outside a git repository
    show_indicator: false

    # Show current position within the repository
    # Table/compact key: git.cwd
    # At repo root: "/"
    # In subdirectory: depends on tree_style
    show_tree: false

    # Show git status output
    # Table/compact key: git.status
    # Color-coded: M=yellow, A=green, D=red, ??=muted (short)
    #              Section-based coloring (long)
    show_status: false

    # How to display the repo position:
    #   breadcrumb - "/enter → cmd → enter"
    #   tree       - tree visualization with └── branches
    tree_style: "breadcrumb"

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
```

### Display Formats

**inline** (default):

```
babarot/enter │ (main *%)
```

**table**:

```
╭───────────┬─────────────────────────────────╮
│pwd        │ babarot/enter                   │
│git.url    │ https://github.com/babarot/enter│
│git.sign   │ (main *%)                       │
│git.cwd    │ /                               │
│git.status │ M  internal/modules/git.go      │
╰───────────┴─────────────────────────────────╯
```

**compact**:

```
pwd        babarot/enter
git.url    https://github.com/babarot/enter
git.sign   (main *%)
git.cwd    /
git.status M  internal/modules/git.go
```

### Table/Compact Row Keys

| Key | Source | Description |
|-----|--------|-------------|
| `pwd` | pwd module | Current directory |
| `git.url` | `show_repo: true` | Repository HTTPS URL |
| `git.sign` | always (when in repo) | Branch, flags, ahead/behind, operation |
| `git.cwd` | `show_tree: true` | Position in repo (breadcrumb or tree) |
| `git.status` | `show_status: true` | git status output (short or long) |
| `kube` | kube module | Kubernetes current-context |
| `gcp` | gcp module | GCP project |

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
│   ├── render.go        Output formatting (inline/table/compact)
│   └── theme.go         Color themes
└── modules/
    ├── pwd.go           Current directory
    ├── git.go           Git status, repo URL, tree, status output
    ├── kube.go          Kubernetes context
    └── gcp.go           GCP project
```

All modules run in parallel using goroutines. Each module returns `nil` if disabled or has nothing to display.

## License

MIT
