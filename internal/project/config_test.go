package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_LocalConfigFolder(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	defer func() {
		_ = os.Chdir(cwd)
	}()

	tempDir, err := os.MkdirTemp("", "docktui-config-test")

	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configDir := filepath.Join(tempDir, "config")

	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	dotConfig := []byte(`
projects:
  - name: project-dot
    images:
      - "img-dot"
`)

	if err := os.WriteFile(".docktui.yaml", dotConfig, 0644); err != nil {
		t.Fatalf("failed to write .docktui.yaml: %v", err)
	}

	aConfig := []byte(`
projects:
  - name: my-project
    labels:
      colour: red
`)

	if err := os.WriteFile(filepath.Join("config", "a.yml"), aConfig, 0644); err != nil {
		t.Fatalf("failed to write config/a.yml: %v", err)
	}

	zConfig := []byte(`
projects:
  - name: my-project
    labels:
      colour: blue
`)

	if err := os.WriteFile(filepath.Join("config", "z.yaml"), zConfig, 0644); err != nil {
		t.Fatalf("failed to write config/z.yaml: %v", err)
	}

	cfg, err := LoadConfig("")

	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	expectedProjects := map[string]struct {
		Labels map[string]string
	}{
		"project-dot": {Labels: nil},
		"my-project":  {Labels: map[string]string{"colour": "blue"}},
	}

	if len(cfg.Projects) != len(expectedProjects) {
		t.Errorf("expected %d projects, got %d", len(expectedProjects), len(cfg.Projects))
	}

	for _, p := range cfg.Projects {
		expected, ok := expectedProjects[p.Name]

		if !ok {
			t.Errorf("unexpected project loaded: %s", p.Name)

			continue
		}

		if expected.Labels != nil {
			if len(p.Labels) != len(expected.Labels) || p.Labels["colour"] != expected.Labels["colour"] {
				t.Errorf("expected labels %v for project %s, got %v", expected.Labels, p.Name, p.Labels)
			}
		}
	}
}

func TestLoadConfig_Explicit(t *testing.T) {
	cwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	defer func() {
		_ = os.Chdir(cwd)
	}()

	tempDir, err := os.MkdirTemp("", "docktui-config-test")

	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configDir := filepath.Join(tempDir, "config")

	if err := os.Mkdir(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	aConfig := []byte(`
projects:
  - name: project-a
    images:
      - "img-a"
`)

	if err := os.WriteFile(filepath.Join("config", "a.yaml"), aConfig, 0644); err != nil {
		t.Fatalf("failed to write config/a.yaml: %v", err)
	}

	explicitFile := filepath.Join(tempDir, "explicit.yaml")
	explicitConfig := []byte(`
projects:
  - name: project-explicit
    images:
      - "img-explicit"
`)

	if err := os.WriteFile(explicitFile, explicitConfig, 0644); err != nil {
		t.Fatalf("failed to write explicit.yaml: %v", err)
	}

	cfg, err := LoadConfig(explicitFile)

	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("expected exactly 1 project, got %d", len(cfg.Projects))
	}

	if cfg.Projects[0].Name != "project-explicit" {
		t.Errorf("expected project-explicit, got %s", cfg.Projects[0].Name)
	}
}
