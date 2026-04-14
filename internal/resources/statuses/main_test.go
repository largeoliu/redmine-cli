// internal/resources/statuses/main_test.go
package statuses

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.LeakTestMain(m)
}
