// internal/testutil/fixtures.go
package testutil

import (
	"time"

	"github.com/largeoliu/redmine-cli/internal/resources/issues"
	"github.com/largeoliu/redmine-cli/internal/resources/projects"
	"github.com/largeoliu/redmine-cli/internal/resources/users"
)

// SampleIssue 返回一个示例 Issue 用于测试
func SampleIssue() issues.Issue {
	createdOn := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedOn := time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC)

	return issues.Issue{
		ID:          1,
		Subject:     "Sample Issue Title",
		Description: "This is a sample issue description for testing purposes.",
		Project: &issues.Reference{
			ID:   1,
			Name: "Sample Project",
		},
		Tracker: &issues.Reference{
			ID:   1,
			Name: "Bug",
		},
		Status: &issues.Reference{
			ID:   1,
			Name: "New",
		},
		Priority: &issues.Reference{
			ID:   2,
			Name: "Normal",
		},
		Author: &issues.Reference{
			ID:   1,
			Name: "John Doe",
		},
		AssignedTo: &issues.Reference{
			ID:   2,
			Name: "Jane Smith",
		},
		Category: &issues.Reference{
			ID:   1,
			Name: "Development",
		},
		StartDate:    "2024-01-15",
		DueDate:      "2024-01-31",
		DoneRatio:    50,
		CreatedOn:    &createdOn,
		UpdatedOn:    &updatedOn,
		PrivateNotes: false,
	}
}

// SampleIssueList 返回一个示例 Issue 列表用于测试
func SampleIssueList() issues.IssueList {
	issue1 := SampleIssue()
	issue2 := SampleIssue()
	issue2.ID = 2
	issue2.Subject = "Second Sample Issue"
	issue2.Description = "Another sample issue for testing list operations."
	issue2.DoneRatio = 0

	return issues.IssueList{
		Issues:     []issues.Issue{issue1, issue2},
		TotalCount: 2,
		Limit:      25,
		Offset:     0,
	}
}

// SampleProject 返回一个示例 Project 用于测试
func SampleProject() projects.Project {
	createdOn := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	updatedOn := time.Date(2024, 1, 10, 16, 30, 0, 0, time.UTC)

	return projects.Project{
		ID:          1,
		Name:        "Sample Project",
		Identifier:  "sample-project",
		Description: "This is a sample project description for testing purposes.",
		Homepage:    "https://example.com/sample-project",
		Status:      1,
		CreatedOn:   &createdOn,
		UpdatedOn:   &updatedOn,
		Trackers: []projects.Reference{
			{ID: 1, Name: "Bug"},
			{ID: 2, Name: "Feature"},
			{ID: 3, Name: "Support"},
		},
		IssueCategories: []projects.Reference{
			{ID: 1, Name: "Development"},
			{ID: 2, Name: "Testing"},
		},
		EnabledModules: []projects.Reference{
			{ID: 1, Name: "issue_tracking"},
			{ID: 2, Name: "time_tracking"},
		},
	}
}

// SampleUser 返回一个示例 User 用于测试
func SampleUser() users.User {
	createdOn := time.Date(2023, 6, 15, 8, 0, 0, 0, time.UTC)
	lastLoginOn := time.Date(2024, 1, 20, 12, 30, 0, 0, time.UTC)

	return users.User{
		ID:                 1,
		Login:              "johndoe",
		Firstname:          "John",
		Lastname:           "Doe",
		Mail:               "john.doe@example.com",
		CreatedOn:          &createdOn,
		LastLoginOn:        &lastLoginOn,
		Admin:              false,
		Status:             1,
		MustChangePassword: false,
		AvatarURL:          "https://example.com/avatar/johndoe.png",
	}
}
