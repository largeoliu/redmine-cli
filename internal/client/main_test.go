// internal/client/main_test.go
package client

import (
	"testing"

	"github.com/largeoliu/redmine-cli/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.LeakTestMain(m)
}
