# enter

[![Tests](https://github.com/babarot/enter/actions/workflows/build.yaml/badge.svg)](https://github.com/babarot/enter/actions/workflows/build.yaml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/babarot/5f07873547f89c2d2b930848a96e9113/raw/enter-coverage.json)](https://github.com/babarot/enter/actions/workflows/build.yaml)

Press Enter. See where you are.

![](https://assets.babarot.dev/files/2026/03/ed1c76d931bd9f89.png)

[![Tests](https://github.com/babarot/enter/actions/workflows/build.yaml/badge.svg)](https://github.com/babarot/enter/actions/workflows/build.yaml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/babarot/5f07873547f89c2d2b930848a96e9113/raw/enter-coverage.json)](https://github.com/babarot/enter/actions/workflows/build.yaml)

## Why

We deal with more context than ever. Git branch, Kubernetes cluster, GCP project, Claude Code usage — all things you need to be aware of, and all things that change depending on where you are.

The traditional answer is to cram this into your shell prompt or tmux statusline. But that approach doesn't scale: the prompt gets long, the statusline gets crowded, and neither adapts well to what actually matters in each directory.

**enter** takes a different approach. Instead of displaying context *all the time*, it shows it **on demand** — when you press Enter on an empty command line. It detects the current directory, figures out what's relevant (git repo? Claude Code project? Kubernetes context?), and displays a clean, structured summary. Nothing clutters your prompt. Nothing runs on every command. Just press Enter when you need to know.

- **Directory-aware**: Only shows what's relevant. Git info appears in repos, Claude usage appears in Claude projects, kube context appears when configured.
- **Pluggable**: Each info source (git, kube, gcp, claude) is an independent module. Enable what you need.
- **Configurable display**: Table or inline format. Tree-style keys. Themes. Symbol customization. All via a single YAML file.
- **Fast**: All modules run in parallel. The display order follows your config file — no extra `order` fields needed.

## Install

**Homebrew**:

```bash
brew install babarot/tap/enter
```

**go install**:

```bash
go install github.com/babarot/enter/cmd/enter@latest
```

**Build from source**:

```bash
git clone https://github.com/babarot/enter.git
cd enter
make build
```

## Setup

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

## Options

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

# Key display style in table format:
#   flat - "git.summary"
#   tree - "├── summary"
key_style: "tree"

# Module display order follows YAML key order.
# Reorder sections below to change display order.
# Unlisted modules are appended in default order.
modules:
  cwd:
    enabled: true
    style: "short"            # short | parent | full | basename

  git:
    enabled: true
    indicator: true           # show "not a git repo" outside repos
    url:
      enabled: true           # repository HTTPS URL (parsed from remote)
    cwd:
      enabled: true
      style: "tree"           # breadcrumb | tree
    summary:
      symbols:
        unstaged: "*"
        staged: "+"
        stash: "$"
        untracked: "%"
        ahead: "↑"
        behind: "↓"
    status:
      enabled: true
      style: "short"          # short | long

  kube:
    enabled: false
    context:
      clean: true             # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false

  claude:
    enabled: true
    mode: "auto"              # always | auto
    usage:
      bar_style: "block"      # block (▰▱) | dot (●○) | fill (█░)
      time_style: "absolute"  # absolute (3:00pm) | relative (22m left)
      cache_ttl: 120          # cache duration in seconds
    config:
      enabled: true
      mode: "auto"            # always (show ✓/✗) | auto (show existing only)
```

### Table Row Keys

| Key | Source | Description |
|-----|--------|-------------|
| `cwd` | cwd module | Current working directory |
| `git.url` | `url.enabled: true` | Repository HTTPS URL |
| `git.cwd` | `cwd.enabled: true` | Position in repo (breadcrumb or tree) |
| `git.summary` | always (when in repo) | Branch, flags, ahead/behind, operation |
| `git.status` | `status.enabled: true` | git status output (short or long) |
| `kube.context` | kube module | Kubernetes context (cleaned if `context.clean: true`) |
| `kube.namespace` | kube module | Kubernetes namespace (defaults to "default") |
| `kube.cluster` | kube module | Kubernetes cluster name |
| `gcp.project` | gcp module | GCP project name |
| `gcp.account` | gcp module | GCP account email |
| `gcp.region` | gcp module | Compute region |
| `gcp.config` | gcp module | Active gcloud config (hidden if "default") |
| `claude.usage.5h` | claude module | 5-hour rolling window utilization |
| `claude.usage.7d` | claude module | 7-day rolling window utilization |
| `claude.config` | `config.enabled: true` | Project config status (CLAUDE.md, rules, skills, etc.) |

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
    summary:        # ← shown second
      symbols: ...
    cwd:         # ← shown third
      enabled: true
    url:         # ← shown fourth
      enabled: true
```

No separate `order` field is needed. Both module order and sub-key order are driven entirely by YAML key order.

Sub-keys within a module (e.g. `git.url`, `git.sign`) can be reordered with the `order` field:

```yaml
modules:
  git:
    order: [sign, cwd, url, status]  # default: [url, sign, cwd, status]
  claude:
    order: [config, usage]           # default: [usage, config]
```

Sub-keys not listed in `order` are appended at the end. Omit `order` to use the default.

### Git Symbols

The `symbols` map in the git config customizes the status indicators shown in `git.summary`:

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
