package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imagetypes "github.com/docker/docker/api/types/image"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type SDKClient struct {
	cli *client.Client
}

func NewSDKClient() (*SDKClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return &SDKClient{cli: cli}, nil
}

func (c *SDKClient) Close() error {
	return c.cli.Close()
}

func (c *SDKClient) DiskUsage(ctx context.Context) (DiskUsage, error) {
	usage, err := c.cli.DiskUsage(ctx, types.DiskUsageOptions{})

	if err != nil {
		return DiskUsage{}, err
	}

	var out DiskUsage

	for _, img := range usage.Images {
		out.TotalSize += img.Size

		if img.Containers == 0 || img.SharedSize == -1 {
			out.ImagesReclaimable += img.Size
		}
	}

	for _, ctr := range usage.Containers {
		out.ContainersReclaimable += ctr.SizeRw
		out.TotalSize += ctr.SizeRw
	}

	for _, vol := range usage.Volumes {
		if vol.UsageData != nil {
			out.VolumesReclaimable += vol.UsageData.Size
			out.TotalSize += vol.UsageData.Size
		}
	}

	for _, cache := range usage.BuildCache {
		out.BuildCacheReclaimable += cache.Size
		out.TotalSize += cache.Size
	}

	out.ReclaimableSize = out.ImagesReclaimable + out.ContainersReclaimable + out.VolumesReclaimable + out.BuildCacheReclaimable

	return out, nil
}

func (c *SDKClient) ListImages(ctx context.Context) ([]Image, error) {
	items, err := c.cli.ImageList(ctx, imagetypes.ListOptions{All: true})

	if err != nil {
		return nil, err
	}

	out := make([]Image, 0, len(items))

	for _, item := range items {
		out = append(out, NormaliseImage(item)...)
	}

	return out, nil
}

func (c *SDKClient) ListContainers(ctx context.Context) ([]Container, error) {
	items, err := c.cli.ContainerList(ctx, containertypes.ListOptions{All: true, Size: true})

	if err != nil {
		return nil, err
	}

	out := make([]Container, 0, len(items))

	for _, item := range items {
		out = append(out, NormaliseContainer(item))
	}

	return out, nil
}

func (c *SDKClient) ListVolumes(ctx context.Context) ([]Volume, error) {
	resp, err := c.cli.VolumeList(ctx, volumetypes.ListOptions{})

	if err != nil {
		return nil, err
	}

	out := make([]Volume, 0, len(resp.Volumes))

	for _, item := range resp.Volumes {
		out = append(out, NormaliseVolume(item))
	}

	return out, nil
}

func (c *SDKClient) ListBuildCache(ctx context.Context) ([]BuildCache, error) {
	usage, err := c.cli.DiskUsage(ctx, types.DiskUsageOptions{})

	if err != nil {
		return nil, err
	}

	out := make([]BuildCache, 0, len(usage.BuildCache))

	for _, item := range usage.BuildCache {
		out = append(out, NormaliseBuildCache(item))
	}

	return out, nil
}

func (c *SDKClient) RemoveImages(ctx context.Context, ids []string) (ActionReport, error) {
	var report ActionReport

	for _, id := range ids {
		deleted, err := c.cli.ImageRemove(ctx, id, imagetypes.RemoveOptions{Force: false, PruneChildren: true})

		if err != nil {
			return report, err
		}

		for _, item := range deleted {
			if item.Deleted != "" {
				report.Deleted = append(report.Deleted, item.Deleted)
			}

			if item.Untagged != "" {
				report.Deleted = append(report.Deleted, item.Untagged)
			}
		}
	}

	return report, nil
}

func (c *SDKClient) RemoveContainers(ctx context.Context, ids []string) (ActionReport, error) {
	var report ActionReport

	for _, id := range ids {
		if err := c.cli.ContainerRemove(ctx, id, containertypes.RemoveOptions{RemoveVolumes: false, Force: false}); err != nil {
			return report, err
		}

		report.Deleted = append(report.Deleted, id)
	}

	return report, nil
}

func (c *SDKClient) RemoveVolumes(ctx context.Context, names []string) (ActionReport, error) {
	var report ActionReport

	for _, name := range names {
		if err := c.cli.VolumeRemove(ctx, name, false); err != nil {
			return report, err
		}

		report.Deleted = append(report.Deleted, name)
	}

	return report, nil
}

func (c *SDKClient) Prune(ctx context.Context, kind PruneType) (ActionReport, error) {
	var report ActionReport

	switch kind {
	case PruneDanglingImages:
		resp, err := c.cli.ImagesPrune(ctx, filters.NewArgs(filters.Arg("dangling", "true")))

		if err != nil {
			return report, err
		}

		for _, item := range resp.ImagesDeleted {
			if item.Deleted != "" {
				report.Deleted = append(report.Deleted, item.Deleted)
			}

			if item.Untagged != "" {
				report.Deleted = append(report.Deleted, item.Untagged)
			}
		}

		report.SpaceReclaimed = int64(resp.SpaceReclaimed)
	case PruneStoppedContainers:
		resp, err := c.cli.ContainersPrune(ctx, filters.Args{})

		if err != nil {
			return report, err
		}

		report.Deleted = resp.ContainersDeleted
		report.SpaceReclaimed = int64(resp.SpaceReclaimed)
	case PruneUnusedVolumes:
		resp, err := c.cli.VolumesPrune(ctx, filters.NewArgs(filters.Arg("all", "true")))

		if err != nil {
			return report, err
		}

		report.Deleted = resp.VolumesDeleted
		report.SpaceReclaimed = int64(resp.SpaceReclaimed)
	case PruneBuildCache:
		resp, err := c.cli.BuildCachePrune(ctx, build.CachePruneOptions{All: true})

		if err != nil {
			return report, err
		}

		report.Deleted = resp.CachesDeleted
		report.SpaceReclaimed = int64(resp.SpaceReclaimed)
	case PruneAllOrphans:
		for _, next := range []PruneType{PruneDanglingImages, PruneStoppedContainers, PruneBuildCache} {
			child, err := c.Prune(ctx, next)

			if err != nil {
				return report, err
			}

			report.Deleted = append(report.Deleted, child.Deleted...)
			report.Warnings = append(report.Warnings, child.Warnings...)
			report.SpaceReclaimed += child.SpaceReclaimed
		}
	default:
		report.Warnings = append(report.Warnings, "unknown prune type")
	}

	return report, nil
}
