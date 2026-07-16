package stacktrace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortFuncName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.Equal("FuncName", shortFuncName("github.com/palantir/shield/package.FuncName"))
	is.Equal("Receiver.MethodName", shortFuncName("github.com/palantir/shield/package.Receiver.MethodName"))
	is.Equal("PtrReceiver.MethodName", shortFuncName("github.com/palantir/shield/package.(*PtrReceiver).MethodName"))
	// Go collapses generic type arguments to a literal "[...]" in runtime
	// function names, so a pointer-receiver method on a generic type never
	// contains more than one '(', '*', or ')' to strip.
	is.Equal("ServiceLazy[...].getInstance", shortFuncName("github.com/samber/do/v2.(*ServiceLazy[...]).getInstance"))
}
