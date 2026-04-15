package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGoReleaserConfigUsesSupportedReleaseTargets(t *testing.T) {
	config := readYAMLMap(t, filepath.Join("..", ".goreleaser.yml"))

	if got := config["version"]; got != 2 {
		t.Fatalf("expected .goreleaser.yml version to be 2, got %v", got)
	}

	release, ok := config["release"].(map[string]any)
	if !ok {
		t.Fatal("expected release section in .goreleaser.yml")
	}

	if _, exists := release["overwrite"]; exists {
		t.Fatal("expected release.overwrite to be removed for GoReleaser v2")
	}

	if got := release["replace_existing_artifacts"]; got != true {
		t.Fatalf("expected release.replace_existing_artifacts=true, got %v", got)
	}

	if _, exists := config["npm"]; exists {
		t.Fatal("expected legacy npm section to be removed")
	}

	if _, exists := config["npms"]; exists {
		t.Fatal("expected npm publishing to be moved out of GoReleaser")
	}

	if _, exists := config["brews"]; exists {
		t.Fatal("expected Homebrew publishing to be removed from .goreleaser.yml")
	}

	archives, ok := config["archives"].([]any)
	if !ok || len(archives) == 0 {
		t.Fatal("expected archives section in .goreleaser.yml")
	}

	archive, ok := archives[0].(map[string]any)
	if !ok {
		t.Fatal("expected first archive entry to be an object")
	}

	if _, exists := archive["format"]; exists {
		t.Fatal("expected archives.format to be replaced with archives.formats")
	}

	if _, exists := archive["formats"]; !exists {
		t.Fatal("expected archives.formats in .goreleaser.yml")
	}

	nameTemplate, ok := archive["name_template"].(string)
	if !ok || nameTemplate == "" {
		t.Fatal("expected archives.name_template in .goreleaser.yml")
	}

	if !strings.HasPrefix(nameTemplate, "redmine-cli_") {
		t.Fatalf("expected archives.name_template to match installer asset prefix, got %q", nameTemplate)
	}

	formatOverrides, ok := archive["format_overrides"].([]any)
	if !ok || len(formatOverrides) == 0 {
		t.Fatal("expected archives.format_overrides in .goreleaser.yml")
	}

	override, ok := formatOverrides[0].(map[string]any)
	if !ok {
		t.Fatal("expected first format_overrides entry to be an object")
	}

	if _, exists := override["format"]; exists {
		t.Fatal("expected archives.format_overrides.format to be replaced with formats")
	}

	if _, exists := override["formats"]; !exists {
		t.Fatal("expected archives.format_overrides.formats in .goreleaser.yml")
	}
}

func TestReleaseWorkflowMatchesConfiguredPublishTargets(t *testing.T) {
	workflow := readWorkflow(t, filepath.Join("..", ".github", "workflows", "release.yml"))
	goreleaserStep := findStepByUses(t, workflow.Jobs["goreleaser"], "goreleaser/goreleaser-action@")

	version, ok := goreleaserStep.With["version"].(string)
	if !ok || version == "" {
		t.Fatal("expected GoReleaser action version to be set")
	}

	if strings.HasPrefix(version, "v1.") {
		t.Fatalf("expected GoReleaser v2-compatible action version, got %q", version)
	}

	if _, exists := goreleaserStep.Env["GITHUB_TOKEN"]; !exists {
		t.Fatal("expected GITHUB_TOKEN env var for GitHub release publishing")
	}

	if _, exists := goreleaserStep.Env["NPM_TOKEN"]; exists {
		t.Fatal("expected GoReleaser step to stop handling npm publishing")
	}

	if _, exists := goreleaserStep.Env["HOMEBREW_TAP_TOKEN"]; exists {
		t.Fatal("expected Homebrew token to be removed from release workflow")
	}

	npmJob, ok := workflow.Jobs["npm-publish"]
	if !ok {
		t.Fatal("expected separate npm-publish job in release workflow")
	}

	setupNode := findStepByUses(t, npmJob, "actions/setup-node@")
	registryURL, ok := setupNode.With["registry-url"].(string)
	if !ok || registryURL == "" {
		t.Fatal("expected npm-publish job to configure registry-url")
	}

	publishStep := findStepByUses(t, npmJob, "")
	if _, exists := publishStep.Env["NODE_AUTH_TOKEN"]; !exists {
		t.Fatal("expected npm publish step to use NODE_AUTH_TOKEN")
	}
}

func TestNPMPackageUsesStandaloneInstaller(t *testing.T) {
	pkg := readJSONMap(t, filepath.Join("..", "package.json"))

	scripts, ok := pkg["scripts"].(map[string]any)
	if !ok {
		t.Fatal("expected scripts object in package.json")
	}

	if got := scripts["postinstall"]; got != "node scripts/download.js" {
		t.Fatalf("expected package.json postinstall to run download.js, got %v", got)
	}

	deps, ok := pkg["dependencies"].(map[string]any)
	if !ok {
		t.Fatal("expected dependencies object in package.json")
	}

	for _, dep := range []string{"tar", "unzipper"} {
		if _, exists := deps[dep]; !exists {
			t.Fatalf("expected dependency %q for npm installer", dep)
		}
	}

	files, ok := toStringSlice(pkg["files"])
	if !ok {
		t.Fatal("expected files array in package.json")
	}

	if !slices.Contains(files, "scripts/download.js") {
		t.Fatal("expected npm package files to include only the installer script")
	}

	if slices.Contains(files, "scripts") {
		t.Fatal("expected npm package files to avoid publishing unrelated scripts")
	}
}

type workflowConfig struct {
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Steps []workflowStep `yaml:"steps"`
}

type workflowStep struct {
	Uses string         `yaml:"uses"`
	With map[string]any `yaml:"with"`
	Env  map[string]any `yaml:"env"`
}

func readYAMLMap(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var cfg map[string]any
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}

	return cfg
}

func readWorkflow(t *testing.T, path string) workflowConfig {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var workflow workflowConfig
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}

	return workflow
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}

	return cfg
}

func findGoReleaserStep(t *testing.T, workflow workflowConfig) workflowStep {
	t.Helper()

	for _, job := range workflow.Jobs {
		if step, ok := findStep(job, func(step workflowStep) bool {
			return strings.HasPrefix(step.Uses, "goreleaser/goreleaser-action@")
		}); ok {
			return step
		}
	}

	t.Fatal("expected goreleaser action step in release workflow")
	return workflowStep{}
}

func findStepByUses(t *testing.T, job workflowJob, prefix string) workflowStep {
	t.Helper()

	step, ok := findStep(job, func(step workflowStep) bool {
		if prefix == "" {
			return len(step.Env) > 0 && step.Uses == ""
		}
		return strings.HasPrefix(step.Uses, prefix)
	})
	if !ok {
		t.Fatalf("expected step with uses prefix %q", prefix)
	}

	return step
}

func findStep(job workflowJob, match func(workflowStep) bool) (workflowStep, bool) {
	for _, step := range job.Steps {
		if match(step) {
			return step, true
		}
	}

	return workflowStep{}, false
}

func toStringSlice(value any) ([]string, bool) {
	raw, ok := value.([]any)
	if !ok {
		return nil, false
	}

	result := make([]string, 0, len(raw))
	for _, item := range raw {
		str, ok := item.(string)
		if !ok {
			return nil, false
		}
		result = append(result, str)
	}

	return result, true
}
