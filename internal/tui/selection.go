package tui

import (
	dockerd "github.com/oullin/docktui/internal/docker"
)

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

func (m *Model) clearSelections() {
	m.selectedImages = map[string]dockerd.Image{}
	m.selectedContainers = map[string]dockerd.Container{}
	m.selectedVolumes = map[string]dockerd.Volume{}
}

func mapValues[T any](items map[string]T) []T {
	out := make([]T, 0, len(items))

	for _, item := range items {
		out = append(out, item)
	}

	return out
}
