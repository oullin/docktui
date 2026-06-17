package docker

import (
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	containertypes "github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	volumetypes "github.com/docker/docker/api/types/volume"
)

func TestNormaliseImageExpandsReferencesAndDetectsDangling(t *testing.T) {
	items := NormaliseImage(imagetypes.Summary{
		ID:         "sha256:abc",
		RepoTags:   []string{"repo/app:latest", "<none>:<none>"},
		Size:       1024,
		Containers: 0,
		Labels:     map[string]string{"com.docker.compose.project": "demo"},
	})

	if len(items) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(items))
	}

	if items[0].Repository != "repo/app" || items[0].Tag != "latest" {
		t.Fatalf("unexpected parsed ref: %#v", items[0])
	}

	if !items[1].Dangling {
		t.Fatalf("expected dangling image")
	}

	if items[0].ProjectName != "demo" {
		t.Fatalf("expected compose project")
	}
}

func TestNormaliseContainerAndVolumeProjectLabels(t *testing.T) {
	container := NormaliseContainer(containertypes.Summary{
		ID:      "abcdef",
		Names:   []string{"/demo-api"},
		Image:   "demo:local",
		State:   "exited",
		Created: 10,
		Labels:  map[string]string{"com.docker.compose.project": "demo"},
	})

	if container.Name != "demo-api" || container.ProjectName != "demo" || !container.CreatedAt.Equal(time.Unix(10, 0)) {
		t.Fatalf("unexpected container: %#v", container)
	}

	volume := NormaliseVolume(&volumetypes.Volume{
		Name:   "demo-cache",
		Driver: "local",
		Labels: map[string]string{"com.docker.compose.project": "demo"},
		UsageData: &volumetypes.UsageData{
			Size: 2048,
		},
	})

	if volume.Size != 2048 || volume.ProjectName != "demo" {
		t.Fatalf("unexpected volume: %#v", volume)
	}
}

func TestNormaliseBuildCache(t *testing.T) {
	created := time.Unix(20, 0)
	item := NormaliseBuildCache(&build.CacheRecord{ID: "cache", Description: "layer", Size: 4096, InUse: true, CreatedAt: created})

	if item.ID != "cache" || item.Size != 4096 || !item.InUse || !item.CreatedAt.Equal(created) {
		t.Fatalf("unexpected build cache: %#v", item)
	}
}
