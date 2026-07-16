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
	// Go may collapse generic type arguments to "[...]" in runtime function names.
	// Regardless of the exact formatting, shortFuncName should only strip the
	// pointer-receiver prefix, not "*" characters inside type arguments.
	is.Equal("ServiceLazy[...].getInstance", shortFuncName("github.com/samber/do/v2.(*ServiceLazy[...]).getInstance"))
	is.Equal("ServiceLazy[*int].getInstance", shortFuncName("github.com/samber/do/v2.(*ServiceLazy[*int]).getInstance"))
}
