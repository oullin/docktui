package tui

import (
	"github.com/oullin/docktui/internal/cleanup"
	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

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
