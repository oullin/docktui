package project

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Projects []Definition `yaml:"projects"`
}

type Definition struct {
	Name    string            `yaml:"name"`
	Images  []string          `yaml:"images"`
	Labels  map[string]string `yaml:"labels"`
	Volumes []string          `yaml:"volumes"`
}

func LoadConfig(explicit string) (Config, error) {
	paths := []string{}

	if explicit != "" {
		paths = append(paths, explicit)
	} else {
		paths = append(paths, ".docktui.yaml")

		var localPaths []string

		if matches, err := filepath.Glob(filepath.Join("config", "*.yaml")); err == nil {
			localPaths = append(localPaths, matches...)
		}

		if matches, err := filepath.Glob(filepath.Join("config", "*.yml")); err == nil {
			localPaths = append(localPaths, matches...)
		}

		sort.Strings(localPaths)

		paths = append(paths, localPaths...)

		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths, filepath.Join(home, ".config", "docktui", "projects.yaml"))
		}
	}

	var merged Config

	for _, path := range paths {
		cfg, err := readConfig(path)

		if err != nil {
			if explicit != "" || !errors.Is(err, os.ErrNotExist) {
				return Config{}, err
			}

			continue
		}

		merged = Merge(merged, cfg)
	}

	return merged, nil
}

func readConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return Config{}, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Merge(configs ...Config) Config {
	byName := map[string]Definition{}

	for _, cfg := range configs {
		for _, def := range cfg.Projects {
			if def.Name == "" {
				continue
			}

			existing := byName[def.Name]
			existing.Name = def.Name
			existing.Images = appendUnique(existing.Images, def.Images...)
			existing.Volumes = appendUnique(existing.Volumes, def.Volumes...)

			if existing.Labels == nil {
				existing.Labels = map[string]string{}
			}

			for key, value := range def.Labels {
				existing.Labels[key] = value
			}

			byName[def.Name] = existing
		}
	}

	names := make([]string, 0, len(byName))

	for name := range byName {
		names = append(names, name)
	}

	sort.Strings(names)

	out := Config{Projects: make([]Definition, 0, len(names))}

	for _, name := range names {
		out.Projects = append(out.Projects, byName[name])
	}

	return out
}

func appendUnique(values []string, next ...string) []string {
	seen := map[string]bool{}

	for _, value := range values {
		seen[value] = true
	}

	for _, value := range next {
		if value == "" || seen[value] {
			continue
		}

		values = append(values, value)
		seen[value] = true
	}

	return values
}
