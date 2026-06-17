package tui

import (
	"context"
	"testing"

	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

func TestRebuildRowsSwitchesBetweenDifferentColumnCounts(t *testing.T) {
	m := New(context.Background(), nil, project.Config{})
	m.width = 100
	m.height = 30
	m.inventory = project.Inventory{
		Images: []dockerd.Image{
			{ID: "img-1", Repository: "demo", Tag: "latest", Size: 1024, ProjectName: "demo"},
		},
	}
	m.projects = project.Config{
		Projects: []project.Definition{{Name: "demo", Images: []string{"demo:latest"}}},
	}

	m.active = tabImages
	m.rebuildRows()

	m.active = tabProjects
	m.rebuildRows()
}
