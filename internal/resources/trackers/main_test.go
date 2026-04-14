// internal/resources/trackers/main_test.go
package trackers

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.LeakTestMain(m)
}
