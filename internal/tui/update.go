package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oullin/docktui/internal/cleanup"
	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.load(), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeTable()

		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd

		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd
	case loadMsg:
		m.loading = false

		if msg.err != nil {
			m.status = msg.err.Error()

			return m, nil
		}

		m.usage = msg.usage
		m.inventory = project.Inventory{Images: msg.images, Containers: msg.containers, Volumes: msg.volumes}
		m.cache = msg.cache
		m.projects = project.Discover(m.config, m.inventory)
		m.rebuildRows()
		m.status = fmt.Sprintf("Loaded %d images, %d containers, %d volumes", len(msg.images), len(msg.containers), len(msg.volumes))

		return m, nil
	case actionMsg:
		m.pending = cleanup.Plan{}
		m.volumePrompt = false

		if msg.err != nil {
			m.status = msg.err.Error()

			return m, nil
		}

		m.clearSelections()
		m.status = fmt.Sprintf("Deleted %d items, reclaimed %s", len(msg.report.Deleted), dockerd.FormatBytes(msg.report.SpaceReclaimed))
		m.loading = true

		return m, tea.Batch(m.load(), m.spinner.Tick)
	case tea.KeyMsg:
		if m.pending.Title != "" {
			return m.updateConfirmation(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			_ = m.client.Close()

			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp

			return m, nil
		case "r":
			m.loading = true
			m.status = "Refreshing Docker resources..."

			return m, tea.Batch(m.load(), m.spinner.Tick)
		case "tab", "right":
			m.active = (m.active + 1) % tab(len(tabNames))
			m.rebuildRows()

			return m, nil
		case "shift+tab", "left":
			m.active = (m.active + tab(len(tabNames)) - 1) % tab(len(tabNames))
			m.rebuildRows()

			return m, nil
		case " ":
			m.toggleSelected()
			m.rebuildRows()

			return m, nil
		case "d":
			m.pending = m.deletePlan()

			if m.pending.Empty() {
				m.pending = cleanup.Plan{}
				m.status = "Select one or more rows first."
			}

			return m, nil
		case "p":
			m.pending = cleanup.ForPrune(pruneTypeForTab(m.active))

			return m, nil
		case "c", "enter":
			if m.active == tabProjects {
				m.pending = m.projectPlan()

				if m.pending.Empty() {
					m.pending = cleanup.Plan{}
					m.status = "Project has no matching cleanup targets."
				}
			}

			return m, nil
		}
	}

	var cmd tea.Cmd

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m Model) updateConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "esc", "n":
		m.pending = cleanup.Plan{}
		m.volumePrompt = false
		m.status = "Canceled."

		return m, nil
	case "y":
		if m.pending.RequiresVolumes && !m.volumePrompt {
			m.volumePrompt = true

			return m, nil
		}

		m.loading = true

		return m, tea.Batch(m.execute(m.pending), m.spinner.Tick)
	case "v":
		if m.volumePrompt {
			m.loading = true

			return m, tea.Batch(m.execute(m.pending), m.spinner.Tick)
		}
	}

	return m, nil
}
