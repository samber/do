package do

import (
	"time"
)

const DefaultStructTagKey = "do"

type InjectorOpts struct {
	HookBeforeRegistration []func(scope *Scope, serviceName string)
	HookAfterRegistration  []func(scope *Scope, serviceName string)
	HookBeforeInvocation   []func(scope *Scope, serviceName string)
	HookAfterInvocation    []func(scope *Scope, serviceName string, err error)
	HookBeforeShutdown     []func(scope *Scope, serviceName string)
	HookAfterShutdown      []func(scope *Scope, serviceName string, err error)

	Logf func(format string, args ...any)

	HealthCheckParallelism   uint          // default: all jobs are executed in parallel
	HealthCheckGlobalTimeout time.Duration // default: no timeout
	HealthCheckTimeout       time.Duration // default: no timeout

	StructTagKey string
}

func (o *InjectorOpts) AddBeforeRegistrationHook(hook func(*Scope, string)) {
	o.HookBeforeRegistration = append(o.HookBeforeRegistration, hook)
}

func (o *InjectorOpts) AddAfterRegistrationHook(hook func(*Scope, string)) {
	o.HookAfterRegistration = append(o.HookAfterRegistration, hook)
}

func (o *InjectorOpts) AddBeforeInvocationHook(hook func(*Scope, string)) {
	o.HookBeforeInvocation = append(o.HookBeforeInvocation, hook)
}

func (o *InjectorOpts) AddAfterInvocationHook(hook func(*Scope, string, error)) {
	o.HookAfterInvocation = append(o.HookAfterInvocation, hook)
}

func (o *InjectorOpts) AddBeforeShutdownHook(hook func(*Scope, string)) {
	o.HookBeforeShutdown = append(o.HookBeforeShutdown, hook)
}

func (o *InjectorOpts) AddAfterShutdownHook(hook func(*Scope, string, error)) {
	o.HookAfterShutdown = append(o.HookAfterShutdown, hook)
}

func (o *InjectorOpts) onBeforeRegistration(scope *Scope, serviceName string) {
	for _, fn := range o.HookBeforeRegistration {
		fn(scope, serviceName)
	}
}

func (o *InjectorOpts) onAfterRegistration(scope *Scope, serviceName string) {
	for _, fn := range o.HookAfterRegistration {
		fn(scope, serviceName)
	}
}

func (o *InjectorOpts) onBeforeInvocation(scope *Scope, serviceName string) {
	for _, fn := range o.HookBeforeInvocation {
		fn(scope, serviceName)
	}
}

func (o *InjectorOpts) onAfterInvocation(scope *Scope, serviceName string, err error) {
	for _, fn := range o.HookAfterInvocation {
		fn(scope, serviceName, err)
	}
}

func (o *InjectorOpts) onBeforeShutdown(scope *Scope, serviceName string) {
	for _, fn := range o.HookBeforeShutdown {
		fn(scope, serviceName)
	}
}

func (o *InjectorOpts) onAfterShutdown(scope *Scope, serviceName string, err error) {
	for _, fn := range o.HookAfterShutdown {
		fn(scope, serviceName, err)
	}
}
