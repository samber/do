package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

func TestNew(t *testing.T) {
	is := assert.New(t)

	i := New()
	is.NotNil(i)

	is.NotNil(i.opts.Logf)
	is.Nil(i.opts.HookAfterRegistration)
	is.Nil(i.opts.HookAfterShutdown)
}

func TestNewWithOpts(t *testing.T) {
	is := assert.New(t)

	i := NewWithOpts(&InjectorOpts{
		HookAfterRegistration: func(scope *Scope, serviceName string) {},
		HookAfterShutdown:     func(scope *Scope, serviceName string) {},
		Logf:                  func(format string, args ...any) {},
	})
	is.NotNil(i)

	is.NotNil(i.opts.HookAfterRegistration)
	is.NotNil(i.opts.HookAfterShutdown)
	is.NotNil(i.opts.Logf)

	is.NotNil(i.self)
	is.Equal("[root]", i.self.name)
	is.Equal(i.self.rootScope, i)
	is.Nil(i.self.parentScope)
}

func TestRootScope_RootScope(t *testing.T) {
	is := assert.New(t)

	i := New()
	is.Equal(i, i.RootScope())
}

func TestRootScope_Ancestors(t *testing.T) {
	is := assert.New(t)

	i := New()
	is.Len(i.Ancestors(), 0)
}

func TestRootScope_Clone(t *testing.T) {
	is := assert.New(t)

	opts := &InjectorOpts{
		HookAfterRegistration: func(scope *Scope, serviceName string) {},
		HookAfterShutdown:     func(scope *Scope, serviceName string) {},
		Logf:                  func(format string, args ...any) {},
	}

	i := NewWithOpts(opts)
	clone := i.Clone()

	is.Equal(i.opts, clone.opts)

	is.NotNil(i.opts.HookAfterRegistration)
	is.NotNil(i.opts.HookAfterShutdown)
	is.NotNil(i.opts.Logf)
	is.NotNil(clone.opts.HookAfterRegistration)
	is.NotNil(clone.opts.HookAfterShutdown)
	is.NotNil(clone.opts.Logf)
}

func TestRootScope_CloneWithOpts(t *testing.T) {
	is := assert.New(t)

	i := New()
	clone := i.CloneWithOpts(&InjectorOpts{
		HookAfterRegistration: func(scope *Scope, serviceName string) {},
		HookAfterShutdown:     func(scope *Scope, serviceName string) {},
		Logf:                  func(format string, args ...any) {},
	})

	is.Nil(i.opts.HookAfterRegistration)
	is.Nil(i.opts.HookAfterShutdown)
	is.NotNil(i.opts.Logf)
	is.NotNil(clone.opts.HookAfterRegistration)
	is.NotNil(clone.opts.HookAfterShutdown)
	is.NotNil(clone.opts.Logf)

	// scope must be added only to initial scope
	i.Scope("foobar")
	is.Len(i.Children(), 1)
	is.Len(clone.Children(), 0)
}

func TestRootScope_ShutdownOnSIGTERM(t *testing.T) {
	// @TODO
}

func TestRootScope_ShutdownOnSignals(t *testing.T) {
	// @TODO
}
