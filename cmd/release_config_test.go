package main

import (
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

	if got := release["replace_existing_artifacts"]; got != false {
		t.Fatalf("expected release.replace_existing_artifacts=false, got %v", got)
	}

	if _, exists := config["npm"]; exists {
		t.Fatal("expected legacy npm section to be removed")
	}

	if _, exists := config["npms"]; exists {
		t.Fatal("expected npm publishing to be moved out of GoReleaser")
	}

	if _, exists := config["brews"]; exists {
		t.Fatal("expected Homebrew section to be removed from .goreleaser.yml")
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

	if version == "latest" {
		t.Fatal("expected GoReleaser action version to be pinned explicitly, not latest")
	}

	if strings.HasPrefix(version, "v1.") {
		t.Fatalf("expected GoReleaser v2-compatible action version, got %q", version)
	}

	if !strings.Contains(version, "v2") {
		t.Fatalf("expected GoReleaser action version to target v2, got %q", version)
	}

	if _, exists := goreleaserStep.Env["GITHUB_TOKEN"]; !exists {
		t.Fatal("expected GITHUB_TOKEN env var for GitHub release publishing")
	}

	if _, exists := goreleaserStep.Env["NPM_TOKEN"]; exists {
		t.Fatal("expected GoReleaser step to stop handling legacy package publishing")
	}

	if _, exists := goreleaserStep.Env["HOMEBREW_TAP_TOKEN"]; exists {
		t.Fatal("expected Homebrew token to be removed from release workflow")
	}

	if _, exists := workflow.Jobs["npm-publish"]; exists {
		t.Fatal("expected release workflow to stop defining the legacy package publish job")
	}
}

func TestRepositoryDoesNotKeepLegacyNodeTooling(t *testing.T) {
	for _, path := range []string{
		filepath.Join("..", "package.json"),
		filepath.Join("..", "package-lock.json"),
		filepath.Join("..", "commitlint.config.js"),
		filepath.Join("..", ".husky", "commit-msg"),
	} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("expected %s to be removed with the legacy Node tooling", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}

func TestGoReleaserLdflagsTargetInternalAppVariables(t *testing.T) {
	config := readYAMLMap(t, filepath.Join("..", ".goreleaser.yml"))

	builds, ok := config["builds"].([]any)
	if !ok || len(builds) == 0 {
		t.Fatal("expected builds section in .goreleaser.yml")
	}

	build, ok := builds[0].(map[string]any)
	if !ok {
		t.Fatal("expected first build entry to be an object")
	}

	ldflags, ok := toStringSlice(build["ldflags"])
	if !ok {
		t.Fatal("expected ldflags array in first build entry")
	}

	for _, want := range []string{
		"-X github.com/largeoliu/redmine-cli/internal/app.version={{.Version}}",
		"-X github.com/largeoliu/redmine-cli/internal/app.commit={{.Commit}}",
		"-X github.com/largeoliu/redmine-cli/internal/app.date={{.Date}}",
	} {
		if !slices.Contains(ldflags, want) {
			t.Fatalf("expected ldflags to contain %q, got %v", want, ldflags)
		}
	}
}

func TestGoReleaserDoesNotMutateReleaseInputs(t *testing.T) {
	config := readYAMLMap(t, filepath.Join("..", ".goreleaser.yml"))

	before, exists := config["before"]
	if !exists {
		return
	}

	beforeMap, ok := before.(map[string]any)
	if !ok {
		t.Fatal("expected before section to be an object")
	}

	hooks, ok := toStringSlice(beforeMap["hooks"])
	if !ok {
		t.Fatal("expected before.hooks to be a string array")
	}

	for _, forbidden := range []string{"go mod tidy", "go generate ./..."} {
		if slices.Contains(hooks, forbidden) {
			t.Fatalf("expected release flow to avoid mutating hook %q", forbidden)
		}
	}
}

func TestReleaseWorkflowHasPreflightGate(t *testing.T) {
	workflow := readWorkflow(t, filepath.Join("..", ".github", "workflows", "release.yml"))

	if _, exists := workflow.Jobs["preflight"]; !exists {
		t.Fatal("expected preflight job in release workflow")
	}

	goreleaserJob, ok := workflow.Jobs["goreleaser"]
	if !ok {
		t.Fatal("expected goreleaser job in release workflow")
	}

	goreleaserNeeds, ok := toStringSliceOrString(goreleaserJob.Needs)
	if !ok || !slices.Contains(goreleaserNeeds, "preflight") {
		t.Fatalf("expected goreleaser job to need preflight, got %v", goreleaserJob.Needs)
	}

	if _, exists := workflow.Jobs["npm-publish"]; exists {
		t.Fatal("expected release workflow to stop after goreleaser without the legacy package publish job")
	}
}

func TestReleaseWorkflowUsesRepositoryDefaultBranchForPreflight(t *testing.T) {
	workflow := readWorkflow(t, filepath.Join("..", ".github", "workflows", "release.yml"))

	preflightJob, ok := workflow.Jobs["preflight"]
	if !ok {
		t.Fatal("expected preflight job in release workflow")
	}

	resolveStep := findStepByName(t, preflightJob, "Resolve protected branch")
	if !strings.Contains(resolveStep.Run, "github.event.repository.default_branch") {
		t.Fatalf("expected Resolve protected branch step to use repository default branch, got %q", resolveStep.Run)
	}

	fetchStep := findStepByName(t, preflightJob, "Fetch protected branch")
	if strings.Contains(fetchStep.Run, "--depth=1") {
		t.Fatalf("expected Fetch protected branch step to avoid shallow fetch, got %q", fetchStep.Run)
	}
}

func TestReleaseWorkflowPreflightChecksTagLineageAndVersionMetadata(t *testing.T) {
	workflow := readWorkflow(t, filepath.Join("..", ".github", "workflows", "release.yml"))

	preflightJob, ok := workflow.Jobs["preflight"]
	if !ok {
		t.Fatal("expected preflight job in release workflow")
	}

	ancestryStep := findStepByName(t, preflightJob, "Ensure tag commit is on protected branch")
	if !strings.Contains(ancestryStep.Run, "git merge-base --is-ancestor") {
		t.Fatalf("expected ancestry step to use git merge-base --is-ancestor, got %q", ancestryStep.Run)
	}

	smokeStep := findStepByName(t, preflightJob, "Smoke test version metadata")
	if !strings.Contains(smokeStep.Run, "go build -ldflags") {
		t.Fatalf("expected smoke step to build with ldflags, got %q", smokeStep.Run)
	}
	if !strings.Contains(smokeStep.Run, "./bin/redmine info") {
		t.Fatalf("expected smoke step to run ./bin/redmine info, got %q", smokeStep.Run)
	}
}

func TestCIWorkflowVerifiesGeneratedFiles(t *testing.T) {
	workflow := readWorkflow(t, filepath.Join("..", ".github", "workflows", "ci.yml"))

	generatedJob, ok := workflow.Jobs["generated"]
	if !ok {
		t.Fatal("expected generated job in CI workflow")
	}

	verifyStep := findStepByName(t, generatedJob, "Verify generated files are committed")
	if !strings.Contains(verifyStep.Run, "go generate ./...") {
		t.Fatalf("expected generated job to run go generate ./..., got %q", verifyStep.Run)
	}
	if !strings.Contains(verifyStep.Run, "git diff --exit-code") {
		t.Fatalf("expected generated job to fail on uncommitted generated output, got %q", verifyStep.Run)
	}

	buildJob, ok := workflow.Jobs["build"]
	if !ok {
		t.Fatal("expected build job in CI workflow")
	}

	buildNeeds, ok := toStringSliceOrString(buildJob.Needs)
	if !ok || !slices.Contains(buildNeeds, "generated") {
		t.Fatalf("expected build job to need generated, got %v", buildJob.Needs)
	}

	ciPassedJob, ok := workflow.Jobs["ci-passed"]
	if !ok {
		t.Fatal("expected ci-passed job in CI workflow")
	}

	ciPassedNeeds, ok := toStringSliceOrString(ciPassedJob.Needs)
	if !ok || !slices.Contains(ciPassedNeeds, "generated") {
		t.Fatalf("expected ci-passed job to need generated, got %v", ciPassedJob.Needs)
	}
}

type workflowConfig struct {
	Jobs map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Steps []workflowStep `yaml:"steps"`
	Needs any            `yaml:"needs"`
}

type workflowStep struct {
	Name string         `yaml:"name"`
	Run  string         `yaml:"run"`
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

func findStepByName(t *testing.T, job workflowJob, name string) workflowStep {
	t.Helper()

	step, ok := findStep(job, func(step workflowStep) bool {
		return step.Name == name
	})
	if !ok {
		t.Fatalf("expected step with name %q", name)
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

func toStringSliceOrString(value any) ([]string, bool) {
	if value == nil {
		return nil, false
	}

	if single, ok := value.(string); ok {
		return []string{single}, true
	}

	return toStringSlice(value)
}
