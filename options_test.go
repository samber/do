package do

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInjectorOpts_addHook(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	hookBeforeRegistration := func(scope *Scope, serviceName string) {}
	hookAfterRegistration := func(scope *Scope, serviceName string) {}
	hookBeforeInvocation := func(scope *Scope, serviceName string) {}
	hookAfterInvocation := func(scope *Scope, serviceName string, err error) {}
	hookBeforeShutdown := func(scope *Scope, serviceName string) {}
	hookAfterShutdown := func(scope *Scope, serviceName string, err error) {}

	i := New()

	//
	is.Len(i.opts.HookBeforeRegistration, 0)
	is.Len(i.opts.HookAfterRegistration, 0)
	is.Len(i.opts.HookBeforeInvocation, 0)
	is.Len(i.opts.HookAfterInvocation, 0)
	is.Len(i.opts.HookBeforeShutdown, 0)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddBeforeRegistrationHook(hookBeforeRegistration)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 0)
	is.Len(i.opts.HookBeforeInvocation, 0)
	is.Len(i.opts.HookAfterInvocation, 0)
	is.Len(i.opts.HookBeforeShutdown, 0)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddAfterRegistrationHook(hookAfterRegistration)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 1)
	is.Len(i.opts.HookBeforeInvocation, 0)
	is.Len(i.opts.HookAfterInvocation, 0)
	is.Len(i.opts.HookBeforeShutdown, 0)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddBeforeInvocationHook(hookBeforeInvocation)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 1)
	is.Len(i.opts.HookBeforeInvocation, 1)
	is.Len(i.opts.HookAfterInvocation, 0)
	is.Len(i.opts.HookBeforeShutdown, 0)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddAfterInvocationHook(hookAfterInvocation)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 1)
	is.Len(i.opts.HookBeforeInvocation, 1)
	is.Len(i.opts.HookAfterInvocation, 1)
	is.Len(i.opts.HookBeforeShutdown, 0)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddBeforeShutdownHook(hookBeforeShutdown)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 1)
	is.Len(i.opts.HookBeforeInvocation, 1)
	is.Len(i.opts.HookAfterInvocation, 1)
	is.Len(i.opts.HookBeforeShutdown, 1)
	is.Len(i.opts.HookAfterShutdown, 0)

	//
	i.AddAfterShutdownHook(hookAfterShutdown)
	is.Len(i.opts.HookBeforeRegistration, 1)
	is.Len(i.opts.HookAfterRegistration, 1)
	is.Len(i.opts.HookBeforeInvocation, 1)
	is.Len(i.opts.HookAfterInvocation, 1)
	is.Len(i.opts.HookBeforeShutdown, 1)
	is.Len(i.opts.HookAfterShutdown, 1)
}

func TestInjectorOpts_onEvent(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	result := ""

	hookBeforeRegistration := func(scope *Scope, serviceName string) { result += "a" }
	hookAfterRegistration := func(scope *Scope, serviceName string) { result += "b" }
	hookBeforeInvocation := func(scope *Scope, serviceName string) { result += "c" }
	hookAfterInvocation := func(scope *Scope, serviceName string, err error) { result += "d" }
	hookBeforeShutdown := func(scope *Scope, serviceName string) { result += "e" }
	hookAfterShutdown := func(scope *Scope, serviceName string, err error) { result += "f" }

	i := NewWithOpts(&InjectorOpts{
		HookBeforeRegistration: []func(scope *Scope, serviceName string){hookBeforeRegistration},
		HookAfterRegistration:  []func(scope *Scope, serviceName string){hookAfterRegistration},
		HookBeforeInvocation:   []func(scope *Scope, serviceName string){hookBeforeInvocation},
		HookAfterInvocation:    []func(scope *Scope, serviceName string, err error){hookAfterInvocation},
		HookBeforeShutdown:     []func(scope *Scope, serviceName string){hookBeforeShutdown},
		HookAfterShutdown:      []func(scope *Scope, serviceName string, err error){hookAfterShutdown},
	})

	i.opts.onBeforeRegistration(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterRegistration(&Scope{id: "id", name: "name"}, "name")
	i.opts.onBeforeInvocation(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterInvocation(&Scope{id: "id", name: "name"}, "name", nil)
	i.opts.onBeforeShutdown(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterShutdown(&Scope{id: "id", name: "name"}, "name", nil)

	is.Equal("abcdef", result)

	i.AddBeforeRegistrationHook(func(scope *Scope, serviceName string) { result += "1" })
	i.AddAfterRegistrationHook(func(scope *Scope, serviceName string) { result += "2" })
	i.AddBeforeInvocationHook(func(scope *Scope, serviceName string) { result += "3" })
	i.AddAfterInvocationHook(func(scope *Scope, serviceName string, err error) { result += "4" })
	i.AddBeforeShutdownHook(func(scope *Scope, serviceName string) { result += "5" })
	i.AddAfterShutdownHook(func(scope *Scope, serviceName string, err error) { result += "6" })

	result = ""

	i.opts.onBeforeRegistration(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterRegistration(&Scope{id: "id", name: "name"}, "name")
	i.opts.onBeforeInvocation(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterInvocation(&Scope{id: "id", name: "name"}, "name", nil)
	i.opts.onBeforeShutdown(&Scope{id: "id", name: "name"}, "name")
	i.opts.onAfterShutdown(&Scope{id: "id", name: "name"}, "name", nil)

	is.Equal("a1b2c3d4e5f6", result)
}
