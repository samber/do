package stacktrace

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func example1() (Frame, bool) {
	return NewFrameFromCaller()
}

func provider(i any) (int, error) {
	return 42, nil
}

func example2() (Frame, bool) {
	return NewFrameFromPC(reflect.ValueOf(provider).Pointer())
}

func TestStacktrace(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	frame, ok := example1()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.True(strings.HasSuffix(frame.File, "do/stacktrace/stacktrace_test.go"))
	is.Equal("example1", frame.Function)
	is.Equal(12, frame.Line)
}

func TestNewFrameFromPtr(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	frame, ok := example2()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.True(strings.HasSuffix(frame.File, "do/stacktrace/stacktrace_test.go"))
	is.Equal("provider", frame.Function)
	is.Equal(16, frame.Line)
}

func TestFrame_String(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	frame, ok := example1()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.Contains(frame.String(), "do/stacktrace/stacktrace_test.go:example1:12")
}

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
