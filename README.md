# enter

Press Enter. See where you are.

![](https://assets.babarot.dev/files/2026/03/d1decd0fe84a91d1.png)

[![Tests](https://github.com/babarot/enter/actions/workflows/build.yaml/badge.svg)](https://github.com/babarot/enter/actions/workflows/build.yaml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/babarot/5f07873547f89c2d2b930848a96e9113/raw/enter-coverage.json)](https://github.com/babarot/enter/actions/workflows/build.yaml)

## Why

We deal with more context than ever. Git branch, Kubernetes cluster, GCP project, Claude Code usage, Codex CLI config — all things you need to be aware of, and all things that change depending on where you are.

The traditional answer is to cram this into your shell prompt or tmux statusline. But that approach doesn't scale: the prompt gets long, the statusline gets crowded, and neither adapts well to what actually matters in each directory.

**enter** takes a different approach. Instead of displaying context *all the time*, it shows it **on demand** — when you press Enter on an empty command line. It detects the current directory, figures out what's relevant (git repo? Claude Code project? Kubernetes context?), and displays a clean, structured summary. Nothing clutters your prompt. Nothing runs on every command. Just press Enter when you need to know.

- **Directory-aware**: Only shows what's relevant. Git info appears in repos, Claude usage appears in Claude projects, Codex config appears in Codex projects, kube context appears when configured.
- **Pluggable**: Each info source (git, kube, gcp, claude, codex) is an independent module. Enable what you need.
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
--edit-config                     Open config file in $EDITOR
--version, -v                     Show version
--help                            Show help
```

## Configuration

Config file location: `~/.config/enter/config.yaml`

Generate a default config:

```bash
enter --init-config
```

Edit the config in your editor:

```bash
enter --edit-config
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
    fields:                   # list fields to display (order matters, omit to hide)
      url:                    # repository HTTPS URL (parsed from remote)
      cwd:
        style: "tree"         # breadcrumb | tree
      summary:
        symbols:
          unstaged: "*"
          staged: "+"
          stash: "$"
          untracked: "%"
          ahead: "↑"
          behind: "↓"
      status:
        style: "short"        # short | long

  kube:
    enabled: false
    fields:
      context:
        clean: true           # strip cloud provider prefixes (GKE/EKS/AKS)

  gcp:
    enabled: false
    # when:                   # conditional display (see below)
    #   dir: "~/src/github.com/mycompany/**"

  claude:
    enabled: true
    mode: "auto"              # always | auto
    fields:                   # list fields to display (order matters, omit to hide)
      usage:
        bar_style: "block"    # block (▰▱) | dot (●○) | fill (█░)
        time_style: "absolute" # absolute (3:00pm) | relative (22m left)
        cache_ttl: 120        # cache duration in seconds
      config:
        mode: "auto"          # always (show ✓/✗) | auto (show existing only)

  codex:
    enabled: true
    mode: "auto"              # always | auto
    fields:                   # list fields to display (order matters, omit to hide)
      config:
        mode: "auto"          # always (show ✓/✗) | auto (show existing only)
```

### Table Row Keys

| Key | Source | Description |
|-----|--------|-------------|
| `cwd` | cwd module | Current working directory |
| `git.url` | `fields: url:` | Repository HTTPS URL |
| `git.cwd` | `fields: cwd:` | Position in repo (breadcrumb or tree) |
| `git.summary` | `fields: summary:` | Branch, flags, ahead/behind, operation |
| `git.status` | `fields: status:` | git status output (short or long) |
| `kube.context` | `fields: context:` | Kubernetes context (cleaned if `clean: true`) |
| `kube.namespace` | `fields: context:` | Kubernetes namespace (defaults to "default") |
| `kube.cluster` | `fields: context:` | Kubernetes cluster name |
| `gcp.project` | gcp module | GCP project name |
| `gcp.account` | gcp module | GCP account email |
| `gcp.region` | gcp module | Compute region |
| `gcp.config` | gcp module | Active gcloud config (hidden if "default") |
| `claude.usage.5h` | `fields: usage:` | 5-hour rolling window utilization |
| `claude.usage.7d` | `fields: usage:` | 7-day rolling window utilization |
| `claude.config` | `fields: config:` | Project config status (CLAUDE.md, rules, skills, etc.) |
| `codex.config` | `fields: config:` | Project config status (AGENTS.md, config.toml, etc.) |

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

Modules not listed in the config file are appended in default order (`cwd`, `git`, `kube`, `gcp`, `claude`, `codex`).

### Field Visibility

Each module's `fields:` section controls **which fields are displayed** and **in what order**. Visibility is determined by presence — there is no `enabled` flag on individual fields.

| Situation | Behavior |
|-----------|----------|
| `fields:` not specified | All fields shown in default order |
| `fields:` with listed keys | Only listed fields shown, in listed order |
| Empty key (e.g. `url:`) | Field is shown with default settings |
| Field omitted from `fields:` | Field is hidden |
| `fields:` with no keys | Warning emitted, no fields shown |

```yaml
modules:
  git:
    enabled: true
    fields:
      status:      # ← shown first
        style: "short"
      summary:     # ← shown second
        symbols: ...
      cwd:         # ← shown third
        style: "tree"
      # url is omitted → hidden
```

No separate `order` field is needed. Both module order and sub-key order are driven entirely by YAML key order.

### Conditional Display

Each module supports a `when` field that controls when the module is shown based on the current working directory. If `when` is not set, the module is always shown (when enabled).

```yaml
modules:
  gcp:
    enabled: true
    when:
      dir: "**/github.com/mycompany/**"

  kube:
    enabled: true
    when:
      dir:
        - "**/github.com/mycompany/**"
        - "**/k8s-*/**"
```

- `dir` accepts a single glob pattern or a list of patterns
- Multiple patterns are OR'd — the module is shown if **any** pattern matches the current directory
- Patterns use [doublestar](https://github.com/bmatcuk/doublestar) glob syntax (`**` matches any number of path segments)
- `~/` is expanded to the home directory

| Condition | Behavior |
|-----------|----------|
| `enabled: false` | Module is always hidden (condition is not evaluated) |
| `enabled: true`, no `when` | Module is always shown |
| `enabled: true`, `when.dir` set | Module is shown only when cwd matches |

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
