// internal/resources/projects/main_test.go
package projects

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.LeakTestMain(m)
}
