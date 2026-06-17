# docktui

`docktui` is a standalone terminal UI for inspecting local Docker disk usage and safely reclaiming space.

It uses the Docker Go SDK behind an internal interface, normalizes Docker resources before they reach the UI, and asks for confirmation before every destructive action. Deleting named volumes requires a second explicit confirmation.

## Usage

```sh
task build
task run
```

Key bindings:

- `r` refresh Docker resources
- `tab`, `shift+tab`, arrows switch tabs or move rows
- `space` select a row on Images, Containers, or Volumes
- `d` delete selected rows
- `p` prune the active resource type
- `c` clean the selected project from the Projects tab
- `?` toggle help
- `q` quit

## Config

`docktui` loads project rules from:

- `./.docktui.yaml`
- `~/.config/docktui/projects.yaml`

Example:

```yaml
projects:
  - name: go-fmt
    images:
      - "go-fmt-*:local"
    labels:
      local.go-fmt.formatter-fingerprint: "*"
    volumes:
      - "go-fmt-cache"
```

Docker Compose projects are auto-detected from the `com.docker.compose.project` label. If a config project has the same name as a detected Compose project, its rules extend the detected project.

## Safety

Bulk `p` pruning on Images, Containers, and Build Cache targets dangling images, stopped containers, and build cache. The Projects tab applies Compose labels plus configured image, label, and volume rules.

`docktui` does not include unused volume pruning in the all-orphans shortcut because named volumes require separate confirmation.

## Development

Task is the canonical command runner:

```sh
task fmt
task lint
task test
task build
task smoke
```
