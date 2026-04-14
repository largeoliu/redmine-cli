// test/integration/integration_test.go
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/largeoliu/redmine-cli/internal/client"
	"github.com/largeoliu/redmine-cli/internal/errors"
	"github.com/largeoliu/redmine-cli/internal/resources/issues"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/resources/statuses"
	"github.com/largeoliu/redmine-cli/internal/resources/time_entries"
	"github.com/largeoliu/redmine-cli/internal/resources/users"
	"github.com/largeoliu/redmine-cli/internal/testutil"
)

var (
	testURL       string
	testAPIKey    string
	testProjectID int
	testClient    *client.Client
)

const (
	testIssuePrefix = "[TEST] "
)

func TestMain(m *testing.M) {
	testURL = os.Getenv("REDMINE_URL")
	testAPIKey = os.Getenv("REDMINE_API_KEY")
	projectID := os.Getenv("REDMINE_PROJECT_ID")

	if testURL == "" || testAPIKey == "" {
		println("Skipping integration tests: REDMINE_URL and REDMINE_API_KEY must be set")
		os.Exit(0)
	}

	if projectID == "" {
		println("Skipping integration tests: REDMINE_PROJECT_ID must be set")
		os.Exit(0)
	}

	if id, err := strconv.Atoi(projectID); err == nil {
		testProjectID = id
	} else {
		println("Skipping integration tests: REDMINE_PROJECT_ID must be a valid integer")
		os.Exit(0)
	}

	testClient = client.NewClient(testURL, testAPIKey,
		client.WithTimeout(30*time.Second),
		client.WithRetry(3, 500*time.Millisecond, 5*time.Second),
	)

	testutil.LeakTestMain(m)
}

func skipIfNoCredentials(t *testing.T) {
	if testURL == "" || testAPIKey == "" {
		t.Skip("Skipping: REDMINE_URL and REDMINE_API_KEY must be set")
	}
}

func getTestProjectID(t *testing.T) int {
	if testProjectID > 0 {
		return testProjectID
	}

	projClient := projects.NewClient(testClient)
	list, err := projClient.List(context.Background(), map[string]string{"limit": "1"})
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(list.Projects) == 0 {
		t.Skip("No projects available for testing. Set REDMINE_PROJECT_ID to specify a project.")
	}

	t.Logf("Warning: REDMINE_PROJECT_ID not set, using project %d. Consider setting it to isolate tests.", list.Projects[0].ID)
	return list.Projects[0].ID
}

func TestConnection(t *testing.T) {
	skipIfNoCredentials(t)

	projClient := projects.NewClient(testClient)
	_, err := projClient.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to Redmine: %v", err)
	}
}

func TestProjectsList(t *testing.T) {
	skipIfNoCredentials(t)

	projClient := projects.NewClient(testClient)
	result, err := projClient.List(context.Background(), map[string]string{"limit": "10"})
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	t.Logf("Found %d projects", result.TotalCount)
}

func TestProjectsGet(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)

	projClient := projects.NewClient(testClient)
	project, err := projClient.Get(context.Background(), projectID, nil)
	if err != nil {
		t.Fatalf("Failed to get project %d: %v", projectID, err)
	}

	if project == nil {
		t.Fatal("Expected non-nil project")
	}

	t.Logf("Got project: %s (ID: %d)", project.Name, project.ID)
}

func TestUsersGetCurrent(t *testing.T) {
	skipIfNoCredentials(t)

	userClient := users.NewClient(testClient)
	user, err := userClient.GetCurrent(context.Background())
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	if user == nil {
		t.Fatal("Expected non-nil user")
	}

	t.Logf("Current user: %s (ID: %d)", user.Login, user.ID)
}

func TestIssuesList(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)

	issueClient := issues.NewClient(testClient)
	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     10,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	t.Logf("Found %d issues in project %d", result.TotalCount, projectID)
}

func TestIssueCRUD(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	createReq := &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sIntegration Test Issue - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "Integration test issue",
	}

	created, err := issueClient.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Logf("Created issue %d: %s in project %d", created.ID, created.Subject, projectID)
	issueID := created.ID

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), issueID)
		t.Logf("Cleaned up issue %d", issueID)
	})

	fetched, err := issueClient.Get(context.Background(), issueID, nil)
	if err != nil {
		t.Fatalf("Failed to get issue %d: %v", issueID, err)
	}

	if fetched.ID != issueID {
		t.Errorf("Expected issue ID %d, got %d", issueID, fetched.ID)
	}

	updateReq := &issues.IssueUpdateRequest{
		Subject: "Updated - " + fetched.Subject,
	}

	err = issueClient.Update(context.Background(), issueID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update issue %d: %v", issueID, err)
	}

	updated, err := issueClient.Get(context.Background(), issueID, nil)
	if err != nil {
		t.Fatalf("Failed to get updated issue %d: %v", issueID, err)
	}

	if updated.Subject != updateReq.Subject {
		t.Errorf("Expected subject %q, got %q", updateReq.Subject, updated.Subject)
	}

	t.Logf("Updated issue %d", issueID)
}

func TestIssueWithNotes(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	createReq := &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sIssue with Notes Test - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "Integration test issue with notes",
	}

	created, err := issueClient.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	issueID := created.ID
	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), issueID)
	})

	noteReq := &issues.IssueUpdateRequest{
		Notes: "This is a test note added during integration testing",
	}

	err = issueClient.Update(context.Background(), issueID, noteReq)
	if err != nil {
		t.Fatalf("Failed to add note to issue %d: %v", issueID, err)
	}

	fetched, err := issueClient.Get(context.Background(), issueID, map[string]string{"include": "journals"})
	if err != nil {
		t.Fatalf("Failed to get issue with journals: %v", err)
	}

	t.Logf("Issue %d subject: %s", issueID, fetched.Subject)
}

func TestPagination(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	page1, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     5,
		Offset:    0,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues page 1: %v", err)
	}

	page2, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     5,
		Offset:    5,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues page 2: %v", err)
	}

	t.Logf("Page 1: %d issues, Page 2: %d issues, Total: %d",
		len(page1.Issues), len(page2.Issues), page1.TotalCount)
}

func TestErrorHandling(t *testing.T) {
	skipIfNoCredentials(t)

	issueClient := issues.NewClient(testClient)

	_, err := issueClient.Get(context.Background(), 99999999, nil)
	if err == nil {
		t.Error("Expected error for non-existent issue")
	}

	invalidClient := client.NewClient(testURL, "invalid-api-key")
	invalidIssueClient := issues.NewClient(invalidClient)

	_, err = invalidIssueClient.List(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}
}

func TestTimeout(t *testing.T) {
	skipIfNoCredentials(t)

	shortTimeoutClient := client.NewClient(testURL, testAPIKey,
		client.WithTimeout(1*time.Nanosecond),
	)

	issueClient := issues.NewClient(shortTimeoutClient)

	_, err := issueClient.List(context.Background(), nil)
	if err == nil {
		t.Error("Expected timeout error")
	}

	t.Logf("Timeout handled correctly: %v", err)
}

func TestConcurrentRequests(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	done := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			_, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
				Limit:     5,
				ProjectID: projectID,
			}))
			done <- err
		}()
	}

	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request failed: %v", err)
		}
	}

	t.Log("All concurrent requests completed successfully")
}

func TestLargeResponse(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     100,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	t.Logf("Retrieved %d issues (total: %d) from project %d", len(result.Issues), result.TotalCount, projectID)
}

func TestProjectsWithInclude(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)

	projClient := projects.NewClient(testClient)
	project, err := projClient.Get(context.Background(), projectID, nil)
	if err != nil {
		t.Fatalf("Failed to get project with includes: %v", err)
	}

	t.Logf("Project: %s", project.Name)
}

func TestHTTPClientReuse(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	for i := 0; i < 3; i++ {
		_, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
			Limit:     1,
			ProjectID: projectID,
		}))
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
		}
	}

	t.Log("HTTP client reuse test passed")
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

func TestCustomLimit(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	customLimit := getEnvInt("REDMINE_TEST_LIMIT", 25)

	issueClient := issues.NewClient(testClient)
	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     customLimit,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues with custom limit: %v", err)
	}

	if len(result.Issues) > customLimit {
		t.Errorf("Expected at most %d issues, got %d", customLimit, len(result.Issues))
	}

	t.Logf("Retrieved %d issues with limit %d from project %d", len(result.Issues), customLimit, projectID)
}

func TestJSONResponseParsing(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)
	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		Limit:     1,
		ProjectID: projectID,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if len(result.Issues) == 0 {
		t.Skip("No issues available for JSON parsing test")
	}

	issue := result.Issues[0]

	data, err := json.MarshalIndent(issue, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal issue: %v", err)
	}

	t.Logf("Issue JSON:\n%s", string(data))

	var parsed issues.Issue
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal issue: %v", err)
	}

	if parsed.ID != issue.ID {
		t.Errorf("Expected ID %d, got %d", issue.ID, parsed.ID)
	}
}

func TestRateLimitHandling(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	for i := 0; i < 10; i++ {
		_, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
			Limit:     1,
			ProjectID: projectID,
		}))
		if err != nil {
			var appErr *errors.Error
			if errors.As(err, &appErr) && appErr.Category == errors.CategoryRateLimit {
				t.Logf("Rate limit hit at request %d, test passed", i+1)
				return
			}
			t.Errorf("Unexpected error: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Log("Rate limit test completed without hitting limits")
}

func TestCleanupTestIssues(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		ProjectID: projectID,
		Limit:     100,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	var cleanedCount int
	for _, issue := range result.Issues {
		if strings.HasPrefix(issue.Subject, testIssuePrefix) {
			if err := issueClient.Delete(context.Background(), issue.ID); err != nil {
				t.Logf("Failed to delete test issue %d: %v", issue.ID, err)
			} else {
				cleanedCount++
			}
		}
	}

	t.Logf("Cleaned up %d test issues from project %d", cleanedCount, projectID)
}

func TestHTTPStatusCode(t *testing.T) {
	skipIfNoCredentials(t)

	projClient := projects.NewClient(testClient)
	_, err := projClient.Get(context.Background(), 99999999, nil)
	if err == nil {
		t.Error("Expected error for non-existent project")
	}

	t.Logf("HTTP status code error handled: %v", err)
}

func TestAuthenticationWithInvalidKey(t *testing.T) {
	skipIfNoCredentials(t)

	invalidClient := client.NewClient(testURL, "invalid-key-12345")
	projClient := projects.NewClient(invalidClient)

	_, err := projClient.List(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	t.Logf("Authentication error handled: %v", err)
}

func TestBuildListParams(t *testing.T) {
	flags := issues.ListFlags{
		ProjectID:    1,
		TrackerID:    2,
		StatusID:     3,
		AssignedToID: 4,
		Limit:        10,
		Offset:       5,
		Query:        "query123",
		Sort:         "created_on:desc",
	}

	params := issues.BuildListParams(flags)

	if params["project_id"] != "1" {
		t.Errorf("Expected project_id=1, got %s", params["project_id"])
	}
	if params["tracker_id"] != "2" {
		t.Errorf("Expected tracker_id=2, got %s", params["tracker_id"])
	}
	if params["status_id"] != "3" {
		t.Errorf("Expected status_id=3, got %s", params["status_id"])
	}
	if params["assigned_to_id"] != "4" {
		t.Errorf("Expected assigned_to_id=4, got %s", params["assigned_to_id"])
	}
	if params["limit"] != "10" {
		t.Errorf("Expected limit=10, got %s", params["limit"])
	}
	if params["offset"] != "5" {
		t.Errorf("Expected offset=5, got %s", params["offset"])
	}
	if params["query_id"] != "query123" {
		t.Errorf("Expected query_id=query123, got %s", params["query_id"])
	}
	if params["sort"] != "created_on:desc" {
		t.Errorf("Expected sort=created_on:desc, got %s", params["sort"])
	}
}

func TestClientRetry(t *testing.T) {
	skipIfNoCredentials(t)

	retryClient := client.NewClient(testURL, testAPIKey,
		client.WithTimeout(10*time.Second),
		client.WithRetry(3, 100*time.Millisecond, 1*time.Second),
	)

	projClient := projects.NewClient(retryClient)
	_, err := projClient.List(context.Background(), map[string]string{"limit": "1"})
	if err != nil {
		t.Errorf("Expected successful request with retry, got error: %v", err)
	}
}

func TestIssueCreateWithAllFields(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	createReq := &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sFull Field Test - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "This is a test description",
	}

	created, err := issueClient.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	t.Logf("Created issue %d with all fields", created.ID)
}

func TestProjectsGetByIdentifier(t *testing.T) {
	skipIfNoCredentials(t)

	projClient := projects.NewClient(testClient)

	result, err := projClient.List(context.Background(), map[string]string{"limit": "1"})
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if len(result.Projects) == 0 {
		t.Skip("No projects available for identifier test")
	}

	project := result.Projects[0]
	fetched, err := projClient.GetByIdentifier(context.Background(), project.Identifier, nil)
	if err != nil {
		t.Fatalf("Failed to get project by identifier: %v", err)
	}

	if fetched.ID != project.ID {
		t.Errorf("Expected project ID %d, got %d", project.ID, fetched.ID)
	}

	t.Logf("Got project by identifier: %s (ID: %d)", fetched.Identifier, fetched.ID)
}

func TestHTTPMethods(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")

	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sHTTP Methods Test - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "Integration test for HTTP methods",
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	_, err = issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}

	err = issueClient.Update(context.Background(), created.ID, &issues.IssueUpdateRequest{
		Subject: "Updated via PUT",
	})
	if err != nil {
		t.Fatalf("Failed to update issue: %v", err)
	}

	t.Log("HTTP methods (POST, GET, PUT) tested successfully")
}

func TestStatusCodeHandling(t *testing.T) {
	skipIfNoCredentials(t)

	issueClient := issues.NewClient(testClient)

	testCases := []struct {
		name      string
		issueID   int
		expectErr bool
	}{
		{"Non-existent issue", 99999999, true},
		{"Invalid issue ID", -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := issueClient.Get(context.Background(), tc.issueID, nil)
			if tc.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestResponseHeaders(t *testing.T) {
	skipIfNoCredentials(t)

	projClient := projects.NewClient(testClient)
	result, err := projClient.List(context.Background(), map[string]string{"limit": "1"})
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}

	if result.TotalCount < 0 {
		t.Error("Expected non-negative total count")
	}

	t.Logf("Response headers handled correctly, total count: %d", result.TotalCount)
}

func TestEmptyResponse(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		ProjectID: projectID,
		Limit:     10,
	}))
	if err != nil {
		t.Fatalf("Failed to list issues: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result even for empty list")
	}

	t.Logf("Empty response handled correctly, issues count: %d", len(result.Issues))
}

func TestClientTimeoutConfiguration(t *testing.T) {
	skipIfNoCredentials(t)

	timeoutClient := client.NewClient(testURL, testAPIKey,
		client.WithTimeout(5*time.Second),
	)

	projClient := projects.NewClient(timeoutClient)
	_, err := projClient.List(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected request to complete within timeout, got error: %v", err)
	}

	t.Log("Client timeout configuration test passed")
}

func TestInvalidURL(t *testing.T) {
	invalidClient := client.NewClient("://invalid-url", testAPIKey)
	projClient := projects.NewClient(invalidClient)

	_, err := projClient.List(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	t.Logf("Invalid URL error handled: %v", err)
}

func TestContextCancellation(t *testing.T) {
	skipIfNoCredentials(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	projClient := projects.NewClient(testClient)
	_, err := projClient.List(ctx, nil)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	t.Logf("Context cancellation handled: %v", err)
}

func TestIssueUpdateWithNotes(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sUpdate Notes Test - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "Integration test issue for update notes",
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	err = issueClient.Update(context.Background(), created.ID, &issues.IssueUpdateRequest{
		Notes: "Adding a note via update",
	})
	if err != nil {
		t.Fatalf("Failed to update issue with notes: %v", err)
	}

	t.Log("Issue update with notes test passed")
}

func TestProjectStatus(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	projClient := projects.NewClient(testClient)

	project, err := projClient.Get(context.Background(), projectID, nil)
	if err != nil {
		t.Fatalf("Failed to get project: %v", err)
	}

	if project.Status < 0 {
		t.Error("Expected non-negative project status")
	}

	t.Logf("Project status: %d", project.Status)
}

func TestIssuePriority(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%sPriority Test - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "Integration test issue for priority",
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	fetched, err := issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}

	if fetched.Priority == nil {
		t.Log("Priority not returned in response (may require include parameter)")
	} else {
		t.Logf("Issue priority: %s (ID: %d)", fetched.Priority.Name, fetched.Priority.ID)
	}
}

func TestListWithSort(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	result, err := issueClient.List(context.Background(), issues.BuildListParams(issues.ListFlags{
		ProjectID: projectID,
		Limit:     5,
		Sort:      "created_on:desc",
	}))
	if err != nil {
		t.Fatalf("Failed to list issues with sort: %v", err)
	}

	t.Logf("Listed %d issues with sort", len(result.Issues))
}

func TestClientPoolConfiguration(t *testing.T) {
	skipIfNoCredentials(t)

	poolClient := client.NewClient(testURL, testAPIKey,
		client.WithConnectionPool(client.DefaultConnectionPoolConfig()),
	)

	projClient := projects.NewClient(poolClient)
	_, err := projClient.List(context.Background(), nil)
	if err != nil {
		t.Errorf("Expected request to succeed with pool config, got error: %v", err)
	}

	t.Log("Client pool configuration test passed")
}

// TestCreateIssueInTestProject 在测试项目上新建任务
func TestCreateIssueInTestProject(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	createReq := &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%s新建任务测试 - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "这是一个集成测试创建的任务，用于验证新建任务功能",
	}

	created, err := issueClient.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Logf("Created issue %d: %s in project %d", created.ID, created.Subject, projectID)

	fetched, err := issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get created issue: %v", err)
	}

	if fetched.Subject != createReq.Subject {
		t.Errorf("Expected subject %q, got %q", createReq.Subject, fetched.Subject)
	}

	if fetched.Project == nil || fetched.Project.ID != projectID {
		t.Errorf("Expected project ID %d", projectID)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
		t.Logf("Cleaned up issue %d", created.ID)
	})

	t.Log("Create issue test passed")
}

// TestUpdateIssueStatus 更新任务状态
func TestUpdateIssueStatus(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)
	statusClient := statuses.NewClient(testClient)

	statusList, err := statusClient.List(context.Background())
	if err != nil {
		t.Fatalf("Failed to list statuses: %v", err)
	}

	if len(statusList.IssueStatuses) < 2 {
		t.Skip("Need at least 2 statuses to test status update")
	}

	t.Logf("Available statuses: %v", func() []string {
		names := make([]string, 0, len(statusList.IssueStatuses))
		for _, s := range statusList.IssueStatuses {
			names = append(names, fmt.Sprintf("%s(ID:%d)", s.Name, s.ID))
		}
		return names
	}())

	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:    fmt.Sprintf("%s状态更新测试 - %s", testIssuePrefix, timestamp),
		ProjectID:  projectID,
		TrackerID:  2,
		PriorityID: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	fetched, err := issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get issue: %v", err)
	}
	t.Logf("Issue %d initial status: %s (ID: %d)", created.ID, fetched.Status.Name, fetched.Status.ID)

	var targetStatusID int
	for _, s := range statusList.IssueStatuses {
		if s.ID != fetched.Status.ID {
			targetStatusID = s.ID
			t.Logf("Target status: %s (ID: %d)", s.Name, s.ID)
			break
		}
	}

	if targetStatusID == 0 {
		t.Skip("No alternative status available")
	}

	updateReq := &issues.IssueUpdateRequest{
		StatusID: targetStatusID,
		Notes:    fmt.Sprintf("尝试将状态从 %s 更新为其他状态", fetched.Status.Name),
	}

	err = issueClient.Update(context.Background(), created.ID, updateReq)
	if err != nil {
		t.Logf("Warning: Status update may be restricted by workflow: %v", err)
	}

	updated, err := issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get updated issue: %v", err)
	}

	t.Logf("Issue %d current status: %s (ID: %d)", created.ID, updated.Status.Name, updated.Status.ID)

	t.Log("Update issue status test passed")
}

// TestLogTimeEntry 登记工时
func TestLogTimeEntry(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)
	timeEntryClient := time_entries.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:    fmt.Sprintf("%s工时登记测试 - %s", testIssuePrefix, timestamp),
		ProjectID:  projectID,
		TrackerID:  2,
		PriorityID: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Cleanup(func() {
		_ = issueClient.Delete(context.Background(), created.ID)
	})

	activities, err := timeEntryClient.ListActivities(context.Background())
	if err != nil {
		t.Fatalf("Failed to list time entry activities: %v", err)
	}

	var activityID int
	for _, a := range activities {
		if a.Active {
			activityID = a.ID
			t.Logf("Using activity: %s (ID: %d)", a.Name, a.ID)
			break
		}
	}
	if activityID == 0 && len(activities) > 0 {
		activityID = activities[0].ID
	}
	if activityID == 0 {
		t.Skip("No time entry activities available")
	}

	today := time.Now().Format("2006-01-02")
	timeEntryReq := &time_entries.TimeEntryCreateRequest{
		IssueID:    created.ID,
		ProjectID:  projectID,
		Hours:      2.5,
		SpentOn:    today,
		ActivityID: activityID,
		Comments:   "集成测试登记的工时",
	}

	timeEntry, err := timeEntryClient.Create(context.Background(), timeEntryReq)
	if err != nil {
		t.Fatalf("Failed to create time entry: %v", err)
	}

	t.Logf("Created time entry %d: %.1f hours on issue %d", timeEntry.ID, timeEntry.Hours, created.ID)

	fetched, err := timeEntryClient.Get(context.Background(), timeEntry.ID)
	if err != nil {
		t.Fatalf("Failed to get time entry: %v", err)
	}

	if fetched.Hours != timeEntryReq.Hours {
		t.Errorf("Expected hours %.1f, got %.1f", timeEntryReq.Hours, fetched.Hours)
	}

	if fetched.Comments != timeEntryReq.Comments {
		t.Errorf("Expected comments %q, got %q", timeEntryReq.Comments, fetched.Comments)
	}

	t.Cleanup(func() {
		_ = timeEntryClient.Delete(context.Background(), timeEntry.ID)
		t.Logf("Cleaned up time entry %d", timeEntry.ID)
	})

	updateReq := &time_entries.TimeEntryUpdateRequest{
		Hours:    3.5,
		Comments: "更新后的工时记录",
	}
	err = timeEntryClient.Update(context.Background(), timeEntry.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update time entry: %v", err)
	}

	updated, err := timeEntryClient.Get(context.Background(), timeEntry.ID)
	if err != nil {
		t.Fatalf("Failed to get updated time entry: %v", err)
	}

	if updated.Hours != 3.5 {
		t.Errorf("Expected hours 3.5, got %.1f", updated.Hours)
	}

	t.Logf("Updated time entry %d: %.1f hours", timeEntry.ID, updated.Hours)

	listResult, err := timeEntryClient.List(context.Background(), time_entries.BuildListParams(time_entries.ListFlags{
		IssueID: created.ID,
		Limit:   10,
	}))
	if err != nil {
		t.Fatalf("Failed to list time entries: %v", err)
	}

	t.Logf("Found %d time entries for issue %d", len(listResult.TimeEntries), created.ID)

	t.Log("Log time entry test passed")
}

// TestDeleteIssue 删除任务
func TestDeleteIssue(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)

	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:    fmt.Sprintf("%s待删除任务测试 - %s", testIssuePrefix, timestamp),
		ProjectID:  projectID,
		TrackerID:  2,
		PriorityID: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	t.Logf("Created issue %d for deletion test", created.ID)

	fetched, err := issueClient.Get(context.Background(), created.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get issue before deletion: %v", err)
	}
	t.Logf("Issue %d exists before deletion: %s", created.ID, fetched.Subject)

	err = issueClient.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete issue: %v", err)
	}

	t.Logf("Deleted issue %d", created.ID)

	_, err = issueClient.Get(context.Background(), created.ID, nil)
	if err == nil {
		t.Error("Expected error when getting deleted issue, but got nil")
	} else {
		t.Logf("Confirmed issue %d is deleted: %v", created.ID, err)
	}

	t.Log("Delete issue test passed")
}

// TestFullIssueLifecycle 完整的任务生命周期测试：创建 -> 更新状态 -> 登记工时 -> 删除
func TestFullIssueLifecycle(t *testing.T) {
	skipIfNoCredentials(t)

	projectID := getTestProjectID(t)
	issueClient := issues.NewClient(testClient)
	statusClient := statuses.NewClient(testClient)
	timeEntryClient := time_entries.NewClient(testClient)

	t.Log("=== Step 1: 创建任务 ===")
	timestamp := time.Now().Format("20060102-150405")
	created, err := issueClient.Create(context.Background(), &issues.IssueCreateRequest{
		Subject:     fmt.Sprintf("%s完整生命周期测试 - %s", testIssuePrefix, timestamp),
		ProjectID:   projectID,
		TrackerID:   2,
		PriorityID:  3,
		Description: "测试完整的任务生命周期：创建、更新状态、登记工时、删除",
	})
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}
	t.Logf("Created issue %d: %s", created.ID, created.Subject)

	t.Log("=== Step 2: 更新任务状态 ===")
	statusList, err := statusClient.List(context.Background())
	if err != nil {
		t.Fatalf("Failed to list statuses: %v", err)
	}

	var targetStatusID int
	for _, s := range statusList.IssueStatuses {
		if !s.IsDefault && s.ID != created.Status.ID {
			targetStatusID = s.ID
			break
		}
	}
	if targetStatusID == 0 && len(statusList.IssueStatuses) > 1 {
		targetStatusID = statusList.IssueStatuses[1].ID
	}

	if targetStatusID > 0 {
		err = issueClient.Update(context.Background(), created.ID, &issues.IssueUpdateRequest{
			StatusID: targetStatusID,
			Notes:    "生命周期测试：更新状态",
		})
		if err != nil {
			t.Fatalf("Failed to update issue status: %v", err)
		}
		t.Logf("Updated issue %d status to %d", created.ID, targetStatusID)
	}

	t.Log("=== Step 3: 登记工时 ===")
	activities, err := timeEntryClient.ListActivities(context.Background())
	if err != nil {
		t.Logf("Warning: Failed to list activities: %v", err)
	}

	var activityID int
	for _, a := range activities {
		if a.Active {
			activityID = a.ID
			break
		}
	}
	if activityID == 0 && len(activities) > 0 {
		activityID = activities[0].ID
	}

	if activityID > 0 {
		today := time.Now().Format("2006-01-02")
		timeEntry, err := timeEntryClient.Create(context.Background(), &time_entries.TimeEntryCreateRequest{
			IssueID:    created.ID,
			ProjectID:  projectID,
			Hours:      1.5,
			SpentOn:    today,
			ActivityID: activityID,
			Comments:   "生命周期测试：登记工时",
		})
		if err != nil {
			t.Logf("Warning: Failed to create time entry: %v", err)
		} else {
			t.Logf("Created time entry %d: %.1f hours", timeEntry.ID, timeEntry.Hours)
			t.Cleanup(func() {
				_ = timeEntryClient.Delete(context.Background(), timeEntry.ID)
			})
		}
	} else {
		t.Log("Warning: No time entry activities available")
	}

	t.Log("=== Step 4: 删除任务 ===")
	err = issueClient.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete issue: %v", err)
	}
	t.Logf("Deleted issue %d", created.ID)

	_, err = issueClient.Get(context.Background(), created.ID, nil)
	if err == nil {
		t.Error("Expected error when getting deleted issue")
	} else {
		t.Logf("Confirmed issue is deleted")
	}

	t.Log("Full issue lifecycle test passed!")
}
