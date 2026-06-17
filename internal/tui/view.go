package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	dockerd "github.com/oullin/docktui/internal/docker"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	activeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Padding(0, 1)
	tabStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Padding(0, 1)
	warnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

func (m Model) header() string {
	line := fmt.Sprintf(
		"%s  total %s  reclaimable %s  images %s  containers %s  volumes %s  cache %s",
		titleStyle.Render("docktui"),
		dockerd.FormatBytes(m.usage.TotalSize),
		dockerd.FormatBytes(m.usage.ReclaimableSize),
		dockerd.FormatBytes(m.usage.ImagesReclaimable),
		dockerd.FormatBytes(m.usage.ContainersReclaimable),
		dockerd.FormatBytes(m.usage.VolumesReclaimable),
		dockerd.FormatBytes(m.usage.BuildCacheReclaimable),
	)
	if m.loading {
		line += "  " + mutedStyle.Render("loading...")
	}
	return line
}

func (m Model) tabs() string {
	parts := make([]string, 0, len(tabNames))
	for i, name := range tabNames {
		if tab(i) == m.active {
			parts = append(parts, activeStyle.Render(name))
			continue
		}
		parts = append(parts, tabStyle.Render(name))
	}
	return strings.Join(parts, " ")
}

func (m Model) footer() string {
	selected := fmt.Sprintf("%d images, %d containers, %d volumes selected", len(m.selectedImages), len(m.selectedContainers), len(m.selectedVolumes))
	keys := "r refresh  tab switch  space select  d delete  p prune  c clean project  ? help  q quit"
	status := m.status
	if status == "" {
		status = selected
	} else {
		status += "  " + mutedStyle.Render(selected)
	}
	return mutedStyle.Render(keys) + "\n" + status
}

func (m Model) help() string {
	return strings.Join([]string{
		titleStyle.Render("Help"),
		"Use arrow keys to move through rows and tabs. Space toggles row selection on Images, Containers, and Volumes.",
		"Delete and prune actions open a preview first. Volume deletion needs a second explicit confirmation.",
		"The Projects tab merges Docker Compose labels with ./.docktui.yaml and ~/.config/docktui/projects.yaml.",
	}, "\n")
}

func (m Model) confirmation() string {
	lines := []string{warnStyle.Render(m.pending.Title)}
	preview := m.pending.Preview
	if len(preview) > 8 {
		preview = append(preview[:8], fmt.Sprintf("...and %d more", len(m.pending.Preview)-8))
	}
	for _, item := range preview {
		lines = append(lines, "  "+item)
	}
	if m.volumePrompt {
		lines = append(lines, warnStyle.Render("Named volumes are included. Press v to permanently delete volumes, or n to cancel."))
	} else {
		lines = append(lines, "Press y to continue, n or esc to cancel.")
	}
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("203")).Padding(1, 2).Render(strings.Join(lines, "\n"))
}
