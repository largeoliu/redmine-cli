// internal/testutil/leak.go
package testutil

import (
	"fmt"
	"os"
	"testing"

	"go.uber.org/goleak"
)

var (
	goleakFind = goleak.Find
	osExitFunc = os.Exit
)

// LeakTestMain provides TestMain function with goroutine leak detection
// Usage:
//
//	func TestMain(m *testing.M) {
//	    testutil.LeakTestMain(m)
//	}
//
// For custom options, use LeakTestMainWithOptions
func LeakTestMain(m *testing.M) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("runtime.gopark"),
	}
	LeakTestMainWithOptions(m, opts...)
}

// LeakTestMainWithOptions provides TestMain function with custom options
func LeakTestMainWithOptions(m *testing.M, opts ...goleak.Option) {
	defer func() {
		if err := goleakFind(opts...); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "goroutine leak detected: %v\n", err)
			osExitFunc(1)
		}
	}()

	osExitFunc(m.Run())
}

// VerifyNone detects current goroutine leaks
// Can be used in tests, for example:
//
//	func TestSomething(t *testing.T) {
//	    defer testutil.VerifyNone(t)
//	    // test code
//	}
func VerifyNone(t *testing.T, opts ...goleak.Option) {
	t.Helper()
	if err := goleakFind(opts...); err != nil {
		t.Errorf("goroutine leak detected: %v", err)
	}
}

// VerifyNoneWithDelay runs VerifyNone in t.Cleanup. The second parameter is for backward compatibility.
func VerifyNoneWithDelay(t *testing.T, _ int, opts ...goleak.Option) {
	t.Helper()
	t.Cleanup(func() {
		VerifyNone(t, opts...)
	})
}

// DefaultLeakOptions returns default goleak options
func DefaultLeakOptions() []goleak.Option {
	return []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("runtime.gopark"),
		goleak.IgnoreTopFunction("internal/poll.runtime_pollWait"),
	}
}

// IgnoreCurrentGoroutines returns an option to ignore current goroutines
func IgnoreCurrentGoroutines() goleak.Option {
	return goleak.IgnoreCurrent()
}
