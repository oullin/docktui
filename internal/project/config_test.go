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
  - name: project-a
    images:
      - "img-a"
`)

	if err := os.WriteFile(filepath.Join("config", "a.yaml"), aConfig, 0644); err != nil {
		t.Fatalf("failed to write config/a.yaml: %v", err)
	}

	bConfig := []byte(`
projects:
  - name: project-b
    images:
      - "img-b"
`)

	if err := os.WriteFile(filepath.Join("config", "b.yml"), bConfig, 0644); err != nil {
		t.Fatalf("failed to write config/b.yml: %v", err)
	}

	cfg, err := LoadConfig("")

	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	expectedProjects := map[string][]string{
		"project-dot": {"img-dot"},
		"project-a":   {"img-a"},
		"project-b":   {"img-b"},
	}

	if len(cfg.Projects) != len(expectedProjects) {
		t.Errorf("expected %d projects, got %d", len(expectedProjects), len(cfg.Projects))
	}

	for _, p := range cfg.Projects {
		expectedImages, ok := expectedProjects[p.Name]

		if !ok {
			t.Errorf("unexpected project loaded: %s", p.Name)

			continue
		}

		if len(p.Images) != len(expectedImages) || p.Images[0] != expectedImages[0] {
			t.Errorf("expected images %v for project %s, got %v", expectedImages, p.Name, p.Images)
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
