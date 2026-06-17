package cleanup

import (
	"fmt"

	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
)

type Action string

type Plan struct {
	Action          Action
	Title           string
	Preview         []string
	ImageIDs        []string
	ContainerIDs    []string
	VolumeNames     []string
	PruneType       dockerd.PruneType
	RequiresVolumes bool
}

const (
	DeleteImages     Action = "delete-images"
	DeleteContainers Action = "delete-containers"
	DeleteVolumes    Action = "delete-volumes"
	CleanProject     Action = "clean-project"
	Prune            Action = "prune"
)

func ForProject(def project.Definition, resources project.Resources) Plan {
	plan := Plan{Action: CleanProject, Title: fmt.Sprintf("Clean project %q", def.Name)}

	for _, item := range resources.Images {
		plan.ImageIDs = append(plan.ImageIDs, item.ID)
		plan.Preview = append(plan.Preview, "image "+item.Reference)
	}

	for _, item := range resources.Containers {
		plan.ContainerIDs = append(plan.ContainerIDs, item.ID)
		plan.Preview = append(plan.Preview, "container "+item.Name)
	}

	for _, item := range resources.Volumes {
		plan.VolumeNames = append(plan.VolumeNames, item.Name)
		plan.Preview = append(plan.Preview, "volume "+item.Name)
		plan.RequiresVolumes = true
	}

	return plan
}

func ForImages(items []dockerd.Image) Plan {
	plan := Plan{Action: DeleteImages, Title: "Delete selected images"}

	for _, item := range items {
		plan.ImageIDs = append(plan.ImageIDs, item.ID)
		plan.Preview = append(plan.Preview, "image "+item.Reference)
	}

	return plan
}

func ForContainers(items []dockerd.Container) Plan {
	plan := Plan{Action: DeleteContainers, Title: "Delete selected containers"}

	for _, item := range items {
		plan.ContainerIDs = append(plan.ContainerIDs, item.ID)
		plan.Preview = append(plan.Preview, "container "+item.Name)
	}

	return plan
}

func ForVolumes(items []dockerd.Volume) Plan {
	plan := Plan{Action: DeleteVolumes, Title: "Delete selected volumes", RequiresVolumes: len(items) > 0}

	for _, item := range items {
		plan.VolumeNames = append(plan.VolumeNames, item.Name)
		plan.Preview = append(plan.Preview, "volume "+item.Name)
	}

	return plan
}

func ForPrune(kind dockerd.PruneType) Plan {
	return Plan{
		Action:    Prune,
		Title:     "Prune " + string(kind),
		PruneType: kind,
		Preview:   []string{"Docker will decide which resources are unused for this prune type."},
	}
}

func (p Plan) Empty() bool {
	return len(p.ImageIDs) == 0 && len(p.ContainerIDs) == 0 && len(p.VolumeNames) == 0 && p.PruneType == ""
}
