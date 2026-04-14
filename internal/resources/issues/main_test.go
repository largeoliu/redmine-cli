// internal/resources/issues/main_test.go
package issues

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.LeakTestMain(m)
}
