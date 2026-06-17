package project

import (
	"testing"

	dockerd "github.com/oullin/docktui/internal/docker"
)

func TestDiscoverMergesComposeAndConfigProjects(t *testing.T) {
	cfg := Config{Projects: []Definition{{
		Name:    "demo",
		Images:  []string{"demo-*:local"},
		Volumes: []string{"demo-cache"},
	}}}
	inv := Inventory{Containers: []dockerd.Container{{Name: "api", ProjectName: "demo"}}, Volumes: []dockerd.Volume{{Name: "other", ProjectName: "extra"}}}

	got := Discover(cfg, inv)
	if len(got.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %#v", got.Projects)
	}
	if got.Projects[0].Name != "demo" || len(got.Projects[0].Images) != 1 {
		t.Fatalf("config rules should extend detected project: %#v", got.Projects[0])
	}
}

func TestMatchProjectRules(t *testing.T) {
	def := Definition{
		Name:    "go-fmt",
		Images:  []string{"go-fmt-*:local"},
		Labels:  map[string]string{"local.go-fmt.formatter-fingerprint": "*"},
		Volumes: []string{"go-fmt-cache"},
	}
	inv := Inventory{
		Images: []dockerd.Image{
			{Reference: "go-fmt-abc:local"},
			{Reference: "other:local", Labels: map[string]string{"local.go-fmt.formatter-fingerprint": "123"}},
		},
		Containers: []dockerd.Container{
			{Name: "formatter", Labels: map[string]string{"local.go-fmt.formatter-fingerprint": "123"}},
		},
		Volumes: []dockerd.Volume{
			{Name: "go-fmt-cache"},
			{Name: "ignored"},
		},
	}

	got := Match(def, inv)
	if len(got.Images) != 2 || len(got.Containers) != 1 || len(got.Volumes) != 1 {
		t.Fatalf("unexpected matches: %#v", got)
	}
}
