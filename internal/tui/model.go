package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/oullin/docktui/internal/cleanup"
	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

type tab int

type Model struct {
	ctx       context.Context
	client    dockerd.Client
	config    project.Config
	projects  project.Config
	inventory project.Inventory
	usage     dockerd.DiskUsage
	cache     []dockerd.BuildCache

	active tab
	table  table.Model
	width  int
	height int

	selectedImages     map[string]dockerd.Image
	selectedContainers map[string]dockerd.Container
	selectedVolumes    map[string]dockerd.Volume

	loading      bool
	status       string
	showHelp     bool
	pending      cleanup.Plan
	volumePrompt bool
	spinner      spinner.Model
}

type loadMsg struct {
	usage      dockerd.DiskUsage
	images     []dockerd.Image
	containers []dockerd.Container
	volumes    []dockerd.Volume
	cache      []dockerd.BuildCache
	err        error
}

type actionMsg struct {
	report dockerd.ActionReport
	err    error
}

const (
	tabImages tab = iota
	tabContainers
	tabVolumes
	tabBuildCache
	tabProjects
)

var tabNames = []string{"Images", "Containers", "Volumes", "Build Cache", "Projects"}

func New(ctx context.Context, client dockerd.Client, config project.Config) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	m := Model{
		ctx:                ctx,
		client:             client,
		config:             config,
		selectedImages:     map[string]dockerd.Image{},
		selectedContainers: map[string]dockerd.Container{},
		selectedVolumes:    map[string]dockerd.Volume{},
		loading:            true,
		spinner:            s,
	}
	m.table = newTable()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.load(), m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildRows()

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
		"",
		m.tabs(),
		"",
		body,
		"",
		m.footer(),
	}

	if m.showHelp {
		parts = append(parts, "", m.help())
	}

	view := strings.Join(parts, "\n")

	if m.pending.Title != "" {
		view += "\n\n" + m.confirmation()
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(view)
}

func (m Model) load() tea.Cmd {
	return func() tea.Msg {
		usage, err := m.client.DiskUsage(m.ctx)

		if err != nil {
			return loadMsg{err: err}
		}

		images, err := m.client.ListImages(m.ctx)

		if err != nil {
			return loadMsg{err: err}
		}

		containers, err := m.client.ListContainers(m.ctx)

		if err != nil {
			return loadMsg{err: err}
		}

		volumes, err := m.client.ListVolumes(m.ctx)

		if err != nil {
			return loadMsg{err: err}
		}

		cache, err := m.client.ListBuildCache(m.ctx)

		if err != nil {
			return loadMsg{err: err}
		}

		return loadMsg{usage: usage, images: images, containers: containers, volumes: volumes, cache: cache}
	}
}

func (m Model) execute(plan cleanup.Plan) tea.Cmd {
	return func() tea.Msg {
		var report dockerd.ActionReport

		var err error

		if len(plan.ContainerIDs) > 0 {
			var next dockerd.ActionReport

			next, err = m.client.RemoveContainers(m.ctx, plan.ContainerIDs)
			report = merge(report, next)

			if err != nil {
				return actionMsg{report: report, err: err}
			}
		}

		if len(plan.ImageIDs) > 0 {
			var next dockerd.ActionReport

			next, err = m.client.RemoveImages(m.ctx, plan.ImageIDs)
			report = merge(report, next)

			if err != nil {
				return actionMsg{report: report, err: err}
			}
		}

		if len(plan.VolumeNames) > 0 {
			var next dockerd.ActionReport

			next, err = m.client.RemoveVolumes(m.ctx, plan.VolumeNames)
			report = merge(report, next)

			if err != nil {
				return actionMsg{report: report, err: err}
			}
		}

		if plan.PruneType != "" {
			var next dockerd.ActionReport

			next, err = m.client.Prune(m.ctx, plan.PruneType)
			report = merge(report, next)
		}

		return actionMsg{report: report, err: err}
	}
}

func merge(a dockerd.ActionReport, b dockerd.ActionReport) dockerd.ActionReport {
	a.Deleted = append(a.Deleted, b.Deleted...)
	a.Warnings = append(a.Warnings, b.Warnings...)
	a.SpaceReclaimed += b.SpaceReclaimed

	return a
}

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
	occupied := 9

	if m.showHelp {
		occupied += 5
	}

	if m.pending.Title != "" {
		previewLen := len(m.pending.Preview)

		if previewLen > 8 {
			previewLen = 9
		}

		occupied += 8 + previewLen
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
		repoWidth := m.width - 5 - 16 - 12 - 16 - 12

		if repoWidth < 28 {
			repoWidth = 28
		}

		cols = columns("Sel", 5, "Repository", repoWidth, "Tag", 16, "Size", 12, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Images))

		for _, item := range sortedImages(m.inventory.Images) {
			rows = append(rows, table.Row{mark(m.selectedImages, item.ID), item.Repository, item.Tag, dockerd.FormatBytes(item.Size), item.ProjectName})
		}
	case tabContainers:
		imageWidth := m.width - 5 - 24 - 10 - 16 - 12

		if imageWidth < 28 {
			imageWidth = 28
		}

		cols = columns("Sel", 5, "Name", 24, "Image", imageWidth, "State", 10, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Containers))

		for _, item := range sortedContainers(m.inventory.Containers) {
			rows = append(rows, table.Row{mark(m.selectedContainers, item.ID), item.Name, item.Image, item.State, item.ProjectName})
		}
	case tabVolumes:
		nameWidth := m.width - 5 - 12 - 12 - 16 - 12

		if nameWidth < 34 {
			nameWidth = 34
		}

		cols = columns("Sel", 5, "Name", nameWidth, "Driver", 12, "Size", 12, "Project", 16)
		rows = make([]table.Row, 0, len(m.inventory.Volumes))

		for _, item := range sortedVolumes(m.inventory.Volumes) {
			rows = append(rows, table.Row{mark(m.selectedVolumes, item.Name), item.Name, item.Driver, dockerd.FormatBytes(item.Size), item.ProjectName})
		}
	case tabBuildCache:
		descWidth := m.width - 18 - 12 - 8 - 11

		if descWidth < 36 {
			descWidth = 36
		}

		cols = columns("ID", 18, "Description", descWidth, "Size", 12, "In Use", 8)
		rows = make([]table.Row, 0, len(m.cache))

		for _, item := range m.cache {
			rows = append(rows, table.Row{shortID(item.ID), item.Description, dockerd.FormatBytes(item.Size), fmt.Sprintf("%t", item.InUse)})
		}
	case tabProjects:
		nameWidth := m.width - 8 - 12 - 8 - 11

		if nameWidth < 24 {
			nameWidth = 24
		}

		cols = columns("Name", nameWidth, "Images", 8, "Containers", 12, "Volumes", 8)
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

func (m *Model) toggleSelected() {
	idx := m.table.Cursor()

	switch m.active {
	case tabImages:
		items := sortedImages(m.inventory.Images)

		if idx < len(items) {
			toggle(m.selectedImages, items[idx].ID, items[idx])
		}
	case tabContainers:
		items := sortedContainers(m.inventory.Containers)

		if idx < len(items) {
			toggle(m.selectedContainers, items[idx].ID, items[idx])
		}
	case tabVolumes:
		items := sortedVolumes(m.inventory.Volumes)

		if idx < len(items) {
			toggle(m.selectedVolumes, items[idx].Name, items[idx])
		}
	}
}

func toggle[T any](items map[string]T, key string, value T) {
	if _, ok := items[key]; ok {
		delete(items, key)

		return
	}

	items[key] = value
}

func mark[T any](items map[string]T, key string) string {
	if _, ok := items[key]; ok {
		return "x"
	}

	return ""
}

func (m Model) deletePlan() cleanup.Plan {
	switch m.active {
	case tabImages:
		return cleanup.ForImages(mapValues(m.selectedImages))
	case tabContainers:
		return cleanup.ForContainers(mapValues(m.selectedContainers))
	case tabVolumes:
		return cleanup.ForVolumes(mapValues(m.selectedVolumes))
	default:
		return cleanup.Plan{}
	}
}

func (m Model) projectPlan() cleanup.Plan {
	idx := m.table.Cursor()

	if idx >= len(m.projects.Projects) {
		return cleanup.Plan{}
	}

	def := m.projects.Projects[idx]

	return cleanup.ForProject(def, project.Match(def, m.inventory))
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

func (m *Model) clearSelections() {
	m.selectedImages = map[string]dockerd.Image{}
	m.selectedContainers = map[string]dockerd.Container{}
	m.selectedVolumes = map[string]dockerd.Volume{}
}

func pruneTypeForTab(active tab) dockerd.PruneType {
	switch active {
	case tabImages:
		return dockerd.PruneDanglingImages
	case tabContainers:
		return dockerd.PruneStoppedContainers
	case tabVolumes:
		return dockerd.PruneUnusedVolumes
	case tabBuildCache:
		return dockerd.PruneBuildCache
	default:
		return dockerd.PruneAllOrphans
	}
}

func mapValues[T any](items map[string]T) []T {
	out := make([]T, 0, len(items))

	for _, item := range items {
		out = append(out, item)
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
