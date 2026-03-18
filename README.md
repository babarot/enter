# enter

Press Enter. See where you are.

![](https://assets.babarot.dev/files/2026/03/ed1c76d931bd9f89.png)

## Why

We deal with more context than ever. Git branch, Kubernetes cluster, GCP project, Claude Code usage — all things you need to be aware of, and all things that change depending on where you are.

The traditional answer is to cram this into your shell prompt or tmux statusline. But that approach doesn't scale: the prompt gets long, the statusline gets crowded, and neither adapts well to what actually matters in each directory.

**enter** takes a different approach. Instead of displaying context *all the time*, it shows it **on demand** — when you press Enter on an empty command line. It detects the current directory, figures out what's relevant (git repo? Claude Code project? Kubernetes context?), and displays a clean, structured summary. Nothing clutters your prompt. Nothing runs on every command. Just press Enter when you need to know.

- **Directory-aware**: Only shows what's relevant. Git info appears in repos, Claude usage appears in Claude projects, kube context appears when configured.
- **Pluggable**: Each info source (git, kube, gcp, claude) is an independent module. Enable what you need.
- **Configurable display**: Table or inline format. Tree-style keys. Themes. Symbol customization. All via a single YAML file.
- **Fast**: All modules run in parallel. The display order follows your config file — no extra `order` fields needed.

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
--last-dir <path>                 Previous directory (for trigger: on_cd)
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
  # ── cwd ──────────────────────────────────────────
  cwd:
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

    context:
      # Strip cloud provider prefixes from context name
      # GKE: gke_project_region_cluster → project/cluster
      # EKS: arn:aws:eks:region:account:cluster/name → name
      # AKS: aks_project_region_cluster → project/cluster
      clean: true

    # Reads from $KUBECONFIG (colon-separated, multiple files)
    # Falls back to ~/.kube/config
    # Empty namespace defaults to "default"

  # ── gcp ──────────────────────────────────────────
  gcp:
    enabled: false

    # Sub-keys: gcp.project, gcp.account, gcp.region, gcp.config
    # gcp.config only shown when active config is not "default"
    #
    # Resolution order (per field):
    # 1. Environment variable ($CLOUDSDK_CORE_PROJECT, $CLOUDSDK_CORE_ACCOUNT, $CLOUDSDK_COMPUTE_REGION)
    # 2. Active gcloud configuration (~/.config/gcloud/configurations/config_{name})
    # 3. Global properties (~/.config/gcloud/properties)

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
│cwd             │ ~/s/g/babarot/project                    │
│git.url         │ https://github.com/babarot/project       │
│git.cwd         │ /                                        │
│git.sign        │ (main *%)                                │
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
| `cwd` | cwd module | Current working directory |
| `git.url` | `show_repo: true` | Repository HTTPS URL |
| `git.cwd` | `show_tree: true` | Position in repo (breadcrumb or tree) |
| `git.sign` | always (when in repo) | Branch, flags, ahead/behind, operation |
| `git.status` | `show_status: true` | git status output (short or long) |
| `kube.context` | kube module | Kubernetes context (cleaned if `context.clean: true`) |
| `kube.namespace` | kube module | Kubernetes namespace (defaults to "default") |
| `kube.cluster` | kube module | Kubernetes cluster name |
| `gcp.project` | gcp module | GCP project name |
| `gcp.account` | gcp module | GCP account email |
| `gcp.region` | gcp module | Compute region |
| `gcp.config` | gcp module | Active gcloud config (hidden if "default") |
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
  cwd:       # ← shown third
    enabled: true
```

Modules not listed in the config file are appended in default order (`cwd`, `git`, `kube`, `gcp`, `claude`).

Sub-keys within a module are also ordered by their position in the config file. Simply reorder the sections:

```yaml
modules:
  git:
    status:      # ← shown first
      enabled: true
    sign:        # ← shown second
      symbols: ...
    cwd:         # ← shown third
      enabled: true
    url:         # ← shown fourth
      enabled: true
```

No separate `order` field is needed. Both module order and sub-key order are driven entirely by YAML key order.

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

## License

MIT
