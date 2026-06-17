package cleanup

import (
	"testing"

	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

func TestForProjectRequiresSecondConfirmationForVolumes(t *testing.T) {
	plan := ForProject(project.Definition{Name: "demo"}, project.Resources{
		Images:  []dockerd.Image{{ID: "img", Reference: "demo:local"}},
		Volumes: []dockerd.Volume{{Name: "demo-data"}},
	})
	if !plan.RequiresVolumes {
		t.Fatalf("expected volume confirmation")
	}
	if len(plan.ImageIDs) != 1 || len(plan.VolumeNames) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
}

func TestPrunePlan(t *testing.T) {
	plan := ForPrune(dockerd.PruneBuildCache)
	if plan.Action != Prune || plan.PruneType != dockerd.PruneBuildCache || plan.Empty() {
		t.Fatalf("unexpected prune plan: %#v", plan)
	}
}
