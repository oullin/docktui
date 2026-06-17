package tui

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

func newTable() table.Model {
	t := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(14),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true)
	styles.Selected = styles.Selected.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Bold(false)
	t.SetStyles(styles)

	return t
}

func (m Model) tableHeight() int {
	occupied := 6

	if m.showHelp {
		occupied += 5
	}

	if m.pending.Title != "" {
		previewLen := len(m.pending.Preview)

		if previewLen > 8 {
			previewLen = 9
		}

		occupied += 7 + previewLen
	}

	h := m.height - occupied

	if h < 6 {
		h = 6
	}

	return h
}

func (m *Model) resizeTable() {
	m.table.SetHeight(m.tableHeight())
}

func (m *Model) rebuildRows() {
	m.resizeTable()

	var cols []table.Column

	var rows []table.Row

	switch m.active {
	case tabImages:
		cols = columns("Sel", 5, "Repository", 28, "Tag", 16, "Size", 12, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Images))

		for _, item := range sortedImages(m.inventory.Images) {
			rows = append(rows, table.Row{mark(m.selectedImages, item.ID), item.Repository, item.Tag, dockerd.FormatBytes(item.Size), item.ProjectName})
		}
	case tabContainers:
		cols = columns("Sel", 5, "Name", 24, "Image", 28, "State", 10, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Containers))

		for _, item := range sortedContainers(m.inventory.Containers) {
			rows = append(rows, table.Row{mark(m.selectedContainers, item.ID), item.Name, item.Image, item.State, item.ProjectName})
		}
	case tabVolumes:
		cols = columns("Sel", 5, "Name", 34, "Driver", 12, "Size", 12, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Volumes))

		for _, item := range sortedVolumes(m.inventory.Volumes) {
			rows = append(rows, table.Row{mark(m.selectedVolumes, item.Name), item.Name, item.Driver, dockerd.FormatBytes(item.Size), item.ProjectName})
		}
	case tabBuildCache:
		cols = columns("ID", 18, "Description", 36, "Size", 12, "In Use", 8)
		rows = make([]table.Row, 0, len(m.cache))

		for _, item := range m.cache {
			rows = append(rows, table.Row{shortID(item.ID), item.Description, dockerd.FormatBytes(item.Size), fmt.Sprintf("%t", item.InUse)})
		}
	case tabProjects:
		cols = columns("Name", 24, "Images", 8, "Containers", 12, "Volumes", 8)
		rows = make([]table.Row, 0, len(m.projects.Projects))

		for _, def := range m.projects.Projects {
			res := project.Match(def, m.inventory)
			rows = append(rows, table.Row{def.Name, fmt.Sprint(len(res.Images)), fmt.Sprint(len(res.Containers)), fmt.Sprint(len(res.Volumes))})
		}
	}

	m.table.SetRows(nil)
	m.table.SetColumns(cols)
	m.table.SetRows(rows)
}

func columns(args ...any) []table.Column {
	out := make([]table.Column, 0, len(args)/2)

	for i := 0; i < len(args); i += 2 {
		out = append(out, table.Column{Title: args[i].(string), Width: args[i+1].(int)})
	}

	return out
}

func sortedImages(items []dockerd.Image) []dockerd.Image {
	out := append([]dockerd.Image(nil), items...)

	sort.Slice(out, func(i, j int) bool { return out[i].Reference < out[j].Reference })

	return out
}

func sortedContainers(items []dockerd.Container) []dockerd.Container {
	out := append([]dockerd.Container(nil), items...)

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

func sortedVolumes(items []dockerd.Volume) []dockerd.Volume {
	out := append([]dockerd.Volume(nil), items...)

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })

	return out
}

func shortID(value string) string {
	if len(value) <= 12 {
		return value
	}

	return value[:12]
}
