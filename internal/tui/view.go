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
	titleText := "DOCKTUI"

	if m.loading {
		titleText += " " + m.spinner.View()
	}

	titleBoxStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	titleBlock := titleBoxStyle.Render(" " + titleText + " ")

	statsText := fmt.Sprintf(
		"Total: %s  Reclaimable: %s (Images: %s, Containers: %s, Volumes: %s, Cache: %s) ",
		dockerd.FormatBytes(m.usage.TotalSize),
		dockerd.FormatBytes(m.usage.ReclaimableSize),
		dockerd.FormatBytes(m.usage.ImagesReclaimable),
		dockerd.FormatBytes(m.usage.ContainersReclaimable),
		dockerd.FormatBytes(m.usage.VolumesReclaimable),
		dockerd.FormatBytes(m.usage.BuildCacheReclaimable),
	)

	statsStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("246"))
	statsBlock := statsStyle.Render(statsText)

	paddingWidth := m.width - lipgloss.Width(titleBlock) - lipgloss.Width(statsBlock)

	if paddingWidth < 0 {
		paddingWidth = 0
	}

	paddingStyle := lipgloss.NewStyle().Background(lipgloss.Color("235"))

	return titleBlock + paddingStyle.Render(strings.Repeat(" ", paddingWidth)) + statsBlock
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

func keyHint(key, desc string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	return keyStyle.Render(key) + " " + descStyle.Render(desc)
}

func (m Model) footer() string {
	hints := []string{
		keyHint("r", "refresh"),
		keyHint("tab", "switch tab"),
		keyHint("space", "select"),
		keyHint("d", "delete"),
		keyHint("p", "prune"),
		keyHint("c", "clean project"),
		keyHint("?", "help"),
		keyHint("q", "quit"),
	}
	keysBar := strings.Join(hints, "  ")

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)
	selStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	statusText := m.status

	if statusText == "" {
		statusText = fmt.Sprintf("Ready  Images: %d • Containers: %d • Volumes: %d",
			len(m.inventory.Images), len(m.inventory.Containers), len(m.inventory.Volumes))
	}

	statusBlock := statusStyle.Render(statusText)

	selText := fmt.Sprintf("Selected: %d images, %d containers, %d volumes ",
		len(m.selectedImages), len(m.selectedContainers), len(m.selectedVolumes))
	selBlock := selStyle.Render(selText)

	paddingWidth := m.width - lipgloss.Width(statusBlock) - lipgloss.Width(selBlock)

	if paddingWidth < 0 {
		paddingWidth = 0
	}

	paddingStyle := lipgloss.NewStyle().Background(lipgloss.Color("236"))
	statusBar := statusBlock + paddingStyle.Render(strings.Repeat(" ", paddingWidth)) + selBlock

	return keysBar + "\n" + statusBar
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

func (m Model) loaderView(height int) string {
	tabName := tabNames[m.active]
	text := fmt.Sprintf("%s Loading %s...", m.spinner.View(), tabName)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(text)
}

func (m Model) View() string {
	if m.width == 0 {
		return "docktui"
	}

	var body string
	h := m.tableHeight()

	if m.loading {
		body = m.loaderView(h + 2)
	} else {
		body = m.table.View()
	}

	parts := []string{
		m.header(),
		m.tabs(),
		body,
		m.footer(),
	}

	if m.showHelp {
		parts = append(parts, m.help())
	}

	view := strings.Join(parts, "\n")

	if m.pending.Title != "" {
		view += "\n" + m.confirmation()
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(view)
}
