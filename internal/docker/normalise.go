package docker

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types/build"
	containertypes "github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	volumetypes "github.com/docker/docker/api/types/volume"
)

func NormaliseImage(item imagetypes.Summary) []Image {
	refs := item.RepoTags

	if len(refs) == 0 {
		refs = item.RepoDigests
	}

	if len(refs) == 0 {
		refs = []string{"<none>:<none>"}
	}

	out := make([]Image, 0, len(refs))

	for _, ref := range refs {
		repo, tag := splitRef(ref)
		out = append(out, Image{
			ID:          item.ID,
			Repository:  repo,
			Tag:         tag,
			Reference:   ref,
			Size:        item.Size,
			Containers:  item.Containers,
			CreatedAt:   unixTime(item.Created),
			Dangling:    repo == "<none>" || tag == "<none>",
			Labels:      item.Labels,
			ProjectName: projectName(item.Labels),
		})
	}

	return out
}

func NormaliseContainer(item containertypes.Summary) Container {
	return Container{
		ID:          item.ID,
		Name:        firstName(item.Names),
		Image:       item.Image,
		State:       item.State,
		Status:      item.Status,
		CreatedAt:   unixTime(item.Created),
		SizeRW:      item.SizeRw,
		Labels:      item.Labels,
		ProjectName: projectName(item.Labels),
	}
}

func NormaliseVolume(item *volumetypes.Volume) Volume {
	var size int64

	if item.UsageData != nil {
		size = item.UsageData.Size
	}

	return Volume{
		Name:        item.Name,
		Driver:      item.Driver,
		Mountpoint:  item.Mountpoint,
		Size:        size,
		Labels:      item.Labels,
		ProjectName: projectName(item.Labels),
	}
}

func NormaliseBuildCache(item *build.CacheRecord) BuildCache {
	var lastUsed time.Time

	if item.LastUsedAt != nil {
		lastUsed = *item.LastUsedAt
	}

	return BuildCache{
		ID:          item.ID,
		Description: item.Description,
		Size:        item.Size,
		InUse:       item.InUse,
		CreatedAt:   item.CreatedAt,
		LastUsedAt:  lastUsed,
	}
}

func splitRef(ref string) (string, string) {
	if strings.Contains(ref, "@") {
		parts := strings.SplitN(ref, "@", 2)

		return parts[0], parts[1]
	}

	idx := strings.LastIndex(ref, ":")

	if idx < 0 || strings.Contains(ref[idx+1:], "/") {
		return ref, ""
	}

	return ref[:idx], ref[idx+1:]
}

func projectName(labels map[string]string) string {
	return labels["com.docker.compose.project"]
}

func firstName(names []string) string {
	if len(names) == 0 {
		return ""
	}

	return strings.TrimPrefix(names[0], "/")
}

func unixTime(sec int64) time.Time {
	if sec <= 0 {
		return time.Time{}
	}

	return time.Unix(sec, 0)
}
