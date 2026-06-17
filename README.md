# docktui

`docktui` is a standalone, full-screen terminal UI for inspecting local Docker disk usage, exploring Docker resources, and safely reclaiming space.

It integrates with the Docker Go SDK to normalize Docker resources, summarizes real-time disk usage statistics, and enforces confirmation safety rails before executing destructive operations.

---

## Features

- **Resource Visualization**: View, paginate, and sort Images, Containers, Volumes, and Build Cache.
- **Disk Usage Stats**: The top header displays total, reclaimable, and resource-specific disk space usage in real time.
- **Project-Centric Management**: Groups resources into logical "Projects" (auto-detected via Docker Compose labels or customized via configuration files) to clean up all related resources at once.
- **Visual Loader**: Visual loading spinners are displayed during data fetching or deletion tasks.
- **Double Confirmation Safety**:
  - Pruning or deleting requires a confirmation prompt (`y` to proceed, `n`/`esc` to cancel).
  - Deleting named volumes requires a **second explicit confirmation** (`v` to confirm volume deletion) to prevent accidental data loss.

---

## Installation & Running

You can run `docktui` either locally using Go and Task, or inside a Docker container using Make (no local Go installation required).

### Option 1: Using Task (Local)

If you have Go and [Task](https://taskfile.dev) installed on your host system:

```sh
# Build the binary into dist/
task build

# Run docktui from source
task run
```

### Option 2: Using Make & Docker (Dockerized)

If you only have Docker and Make installed:

```sh
# Run docktui inside a container with the docker socket mounted
make run

# Build the binary inside a container
make build
```

---

## Command Line Flags

`docktui` supports the following flags:

- `-config <path>`: Specifies a custom YAML configuration file. When provided, default config paths are ignored.
- `-version`: Prints the current version and exits.

---

## Navigation & Key Bindings

| Key | Action |
|---|---|
| `tab` / `→` | Switch to the next tab |
| `shift+tab` / `←` | Switch to the previous tab |
| `↑` / `↓` | Navigate through table rows |
| `space` | Toggle row selection (on Images, Containers, and Volumes tabs) |
| `d` | Delete selected resources (opens confirmation preview) |
| `p` | Prune the active resource type (dangling/stopped resources) |
| `c` / `enter` | Clean the selected project (from the Projects tab) |
| `r` | Refresh Docker resources |
| `?` | Toggle the Help overlay |
| `q` / `ctrl+c` | Quit the application |

---

## Configuration

`docktui` loads and merges project rules from the following paths by default:
1. `./.docktui.yaml`
2. Local files matching `config/*.yaml` or `config/*.yml` (ignored by git except for `config/.gitkeep`)
3. `~/.config/docktui/projects.yaml`

If the `-config` flag is explicitly provided, only the specified configuration file is loaded.

### Wildcard / Glob Matching
When defining patterns for `images`, `labels`, or `volumes`, `docktui` supports standard shell globbing wildcards (e.g., `*`, `?`, `[a-z]`).

### Example Configuration (`.docktui.yaml`)

```yaml
projects:
  - name: my-go-app
    images:
      - "my-go-app-*:latest"
      - "go-formatter-*:local"
    labels:
      local.go-app.component: "backend"
      formatter-fingerprint: "*"
    volumes:
      - "go-cache-volume"
```

If a configured project name matches an auto-detected Docker Compose project, `docktui` merges your manual rules with the detected Compose rules, allowing you to clean up custom-labeled sidecars or cache volumes alongside standard Compose resources.

---

## Development

Canonical tools and tasks for local development:

### Local Tasks (using `task`)

```sh
# Format Go code
task fmt

# Format all Go files using go-fmt v0.4.1
task format-all

# Lint the codebase
task lint

# Run unit tests
task test

# Verify the build and check version output
task smoke
```

### Docker-based Tasks (using `make`)

All commands are executed within a consistent Dockerized Go environment (`golang:1.26.4-alpine`):

```sh
# Format all Go code in Docker
make format-all

# Lint codebase in Docker
make lint

# Run unit tests in Docker
make test

# Smoke-test binary in Docker
make smoke
```
