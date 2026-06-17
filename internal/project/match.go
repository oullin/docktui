package project

import (
	"path/filepath"
	"sort"

	dockerd "github.com/oullin/docktui/internal/docker"
)

type Inventory struct {
	Images     []dockerd.Image
	Containers []dockerd.Container
	Volumes    []dockerd.Volume
}

type Resources struct {
	Images     []dockerd.Image
	Containers []dockerd.Container
	Volumes    []dockerd.Volume
}

func Discover(config Config, inv Inventory) Config {
	detected := Config{}
	names := map[string]bool{}

	for _, item := range inv.Images {
		if item.ProjectName != "" {
			names[item.ProjectName] = true
		}
	}

	for _, item := range inv.Containers {
		if item.ProjectName != "" {
			names[item.ProjectName] = true
		}
	}

	for _, item := range inv.Volumes {
		if item.ProjectName != "" {
			names[item.ProjectName] = true
		}
	}

	list := make([]string, 0, len(names))

	for name := range names {
		list = append(list, name)
	}

	sort.Strings(list)

	for _, name := range list {
		detected.Projects = append(detected.Projects, Definition{Name: name})
	}

	return Merge(detected, config)
}

func Match(def Definition, inv Inventory) Resources {
	var out Resources

	for _, item := range inv.Images {
		if item.ProjectName == def.Name || matchesAny(def.Images, item.Reference) || labelsMatch(def.Labels, item.Labels) {
			out.Images = append(out.Images, item)
		}
	}

	for _, item := range inv.Containers {
		if item.ProjectName == def.Name || labelsMatch(def.Labels, item.Labels) {
			out.Containers = append(out.Containers, item)
		}
	}

	for _, item := range inv.Volumes {
		if item.ProjectName == def.Name || matchesAny(def.Volumes, item.Name) || labelsMatch(def.Labels, item.Labels) {
			out.Volumes = append(out.Volumes, item)
		}
	}

	return out
}

func matchesAny(patterns []string, value string) bool {
	for _, pattern := range patterns {
		ok, err := filepath.Match(pattern, value)

		if err == nil && ok {
			return true
		}
	}

	return false
}

func labelsMatch(want, got map[string]string) bool {
	if len(want) == 0 || len(got) == 0 {
		return false
	}

	for key, pattern := range want {
		value, ok := got[key]

		if !ok {
			return false
		}

		matched, err := filepath.Match(pattern, value)

		if err != nil || !matched {
			return false
		}
	}

	return true
}
