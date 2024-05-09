package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectorOpts_addHook(t *testing.T) {
	is := assert.New(t)

	hookBeforeRegistration := func(scope *Scope, serviceName string) {}
	hookAfterRegistration := func(scope *Scope, serviceName string) {}
	hookBeforeInvocation := func(scope *Scope, serviceName string) {}
	hookAfterInvocation := func(scope *Scope, serviceName string, err error) {}
	hookBeforeShutdown := func(scope *Scope, serviceName string) {}
	hookAfterShutdown := func(scope *Scope, serviceName string, err error) {}

	opts := &InjectorOpts{}

	//
	is.Len(opts.HookBeforeRegistration, 0)
	is.Len(opts.HookAfterRegistration, 0)
	is.Len(opts.HookBeforeInvocation, 0)
	is.Len(opts.HookAfterInvocation, 0)
	is.Len(opts.HookBeforeShutdown, 0)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddBeforeRegistrationHook(hookBeforeRegistration)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 0)
	is.Len(opts.HookBeforeInvocation, 0)
	is.Len(opts.HookAfterInvocation, 0)
	is.Len(opts.HookBeforeShutdown, 0)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddAfterRegistrationHook(hookAfterRegistration)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 1)
	is.Len(opts.HookBeforeInvocation, 0)
	is.Len(opts.HookAfterInvocation, 0)
	is.Len(opts.HookBeforeShutdown, 0)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddBeforeInvocationHook(hookBeforeInvocation)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 1)
	is.Len(opts.HookBeforeInvocation, 1)
	is.Len(opts.HookAfterInvocation, 0)
	is.Len(opts.HookBeforeShutdown, 0)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddAfterInvocationHook(hookAfterInvocation)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 1)
	is.Len(opts.HookBeforeInvocation, 1)
	is.Len(opts.HookAfterInvocation, 1)
	is.Len(opts.HookBeforeShutdown, 0)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddBeforeShutdownHook(hookBeforeShutdown)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 1)
	is.Len(opts.HookBeforeInvocation, 1)
	is.Len(opts.HookAfterInvocation, 1)
	is.Len(opts.HookBeforeShutdown, 1)
	is.Len(opts.HookAfterShutdown, 0)

	//
	opts.AddAfterShutdownHook(hookAfterShutdown)
	is.Len(opts.HookBeforeRegistration, 1)
	is.Len(opts.HookAfterRegistration, 1)
	is.Len(opts.HookBeforeInvocation, 1)
	is.Len(opts.HookAfterInvocation, 1)
	is.Len(opts.HookBeforeShutdown, 1)
	is.Len(opts.HookAfterShutdown, 1)
}

func TestInjectorOpts_onEvent(t *testing.T) {
	is := assert.New(t)

	result := ""

	hookBeforeRegistration := func(scope *Scope, serviceName string) { result += "a" }
	hookAfterRegistration := func(scope *Scope, serviceName string) { result += "b" }
	hookBeforeInvocation := func(scope *Scope, serviceName string) { result += "c" }
	hookAfterInvocation := func(scope *Scope, serviceName string, err error) { result += "d" }
	hookBeforeShutdown := func(scope *Scope, serviceName string) { result += "e" }
	hookAfterShutdown := func(scope *Scope, serviceName string, err error) { result += "f" }

	opts := &InjectorOpts{
		HookBeforeRegistration: []func(scope *Scope, serviceName string){hookBeforeRegistration},
		HookAfterRegistration:  []func(scope *Scope, serviceName string){hookAfterRegistration},
		HookBeforeInvocation:   []func(scope *Scope, serviceName string){hookBeforeInvocation},
		HookAfterInvocation:    []func(scope *Scope, serviceName string, err error){hookAfterInvocation},
		HookBeforeShutdown:     []func(scope *Scope, serviceName string){hookBeforeShutdown},
		HookAfterShutdown:      []func(scope *Scope, serviceName string, err error){hookAfterShutdown},
	}

	opts.onBeforeRegistration(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterRegistration(&Scope{id: "id", name: "name"}, "name")
	opts.onBeforeInvocation(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterInvocation(&Scope{id: "id", name: "name"}, "name", nil)
	opts.onBeforeShutdown(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterShutdown(&Scope{id: "id", name: "name"}, "name", nil)

	is.Equal("abcdef", result)

	opts.AddBeforeRegistrationHook(func(scope *Scope, serviceName string) { result += "1" })
	opts.AddAfterRegistrationHook(func(scope *Scope, serviceName string) { result += "2" })
	opts.AddBeforeInvocationHook(func(scope *Scope, serviceName string) { result += "3" })
	opts.AddAfterInvocationHook(func(scope *Scope, serviceName string, err error) { result += "4" })
	opts.AddBeforeShutdownHook(func(scope *Scope, serviceName string) { result += "5" })
	opts.AddAfterShutdownHook(func(scope *Scope, serviceName string, err error) { result += "6" })

	result = ""

	opts.onBeforeRegistration(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterRegistration(&Scope{id: "id", name: "name"}, "name")
	opts.onBeforeInvocation(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterInvocation(&Scope{id: "id", name: "name"}, "name", nil)
	opts.onBeforeShutdown(&Scope{id: "id", name: "name"}, "name")
	opts.onAfterShutdown(&Scope{id: "id", name: "name"}, "name", nil)

	is.Equal("a1b2c3d4e5f6", result)
}
