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
	return NewFrameFromPtr(reflect.ValueOf(provider).Pointer())
}

func TestStacktrace(t *testing.T) {
	is := assert.New(t)

	frame, ok := example1()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.True(strings.HasSuffix(frame.File, "github.com/samber/do/stacktrace/stacktrace_test.go"))
	is.Equal("example1", frame.Function)
	is.Equal(12, frame.Line)
}

func TestNewFrameFromPtr(t *testing.T) {
	is := assert.New(t)

	frame, ok := example2()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.True(strings.HasSuffix(frame.File, "github.com/samber/do/stacktrace/stacktrace_test.go"))
	is.Equal("provider", frame.Function)
	is.Equal(16, frame.Line)
}

func TestFrame_String(t *testing.T) {
	is := assert.New(t)

	frame, ok := example1()
	is.True(ok)
	is.NotNil(frame)
	is.NotEmpty(frame)
	is.True(strings.Contains(frame.String(), "github.com/samber/do/stacktrace/stacktrace_test.go:example1:12"))
}
