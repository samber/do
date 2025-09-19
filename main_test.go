package do

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type ctxTestkey string

const (
	ctxTestKey ctxTestkey = "test-key"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// https://github.com/stretchr/testify/issues/1101
func testWithTimeout(t *testing.T, timeout time.Duration) {
	t.Helper()

	testFinished := make(chan struct{})
	t.Cleanup(func() { close(testFinished) })

	go func() {
		select {
		case <-testFinished:
		case <-time.After(timeout):
			t.Errorf("test timed out after %s", timeout)
			os.Exit(1)
		}
	}()
}

// pkgName holds the current package name (eg. "do"). It won't break
// if the package name is changed, in case of a fork or anything.
var pkgName = func() string {
	t := reflect.TypeOf(__pkg_probe_type{})
	s := t.String() // eg. "do.__pkg_probe_type"
	if i := strings.IndexByte(s, '.'); i > 0 {
		return s[:i]
	}
	return "do"
}()

type __pkg_probe_type struct{}
