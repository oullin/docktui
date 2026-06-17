package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oullin/docktui/internal/cleanup"
	dockerd "github.com/oullin/docktui/internal/docker"
)

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
