package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
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
