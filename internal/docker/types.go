package docker

import (
	"context"
	"time"
)

type Client interface {
	DiskUsage(context.Context) (DiskUsage, error)
	ListImages(context.Context) ([]Image, error)
	ListContainers(context.Context) ([]Container, error)
	ListVolumes(context.Context) ([]Volume, error)
	ListBuildCache(context.Context) ([]BuildCache, error)
	RemoveImages(context.Context, []string) (ActionReport, error)
	RemoveContainers(context.Context, []string) (ActionReport, error)
	RemoveVolumes(context.Context, []string) (ActionReport, error)
	Prune(context.Context, PruneType) (ActionReport, error)
	Close() error
}

type PruneType string

type DiskUsage struct {
	ImagesReclaimable     int64
	ContainersReclaimable int64
	VolumesReclaimable    int64
	BuildCacheReclaimable int64
	TotalSize             int64
	ReclaimableSize       int64
}

type Image struct {
	ID          string
	Repository  string
	Tag         string
	Reference   string
	Size        int64
	Containers  int64
	CreatedAt   time.Time
	Dangling    bool
	Labels      map[string]string
	ProjectName string
}

type Container struct {
	ID          string
	Name        string
	Image       string
	State       string
	Status      string
	CreatedAt   time.Time
	SizeRW      int64
	Labels      map[string]string
	ProjectName string
}

type Volume struct {
	Name        string
	Driver      string
	Mountpoint  string
	Size        int64
	Labels      map[string]string
	ProjectName string
}

type BuildCache struct {
	ID          string
	Description string
	Size        int64
	InUse       bool
	CreatedAt   time.Time
	LastUsedAt  time.Time
}

type ActionReport struct {
	Deleted        []string
	SpaceReclaimed int64
	Warnings       []string
}

const (
	PruneDanglingImages    PruneType = "dangling-images"
	PruneStoppedContainers PruneType = "stopped-containers"
	PruneUnusedVolumes     PruneType = "unused-volumes"
	PruneBuildCache        PruneType = "build-cache"
	PruneAllOrphans        PruneType = "all-orphans"
)
