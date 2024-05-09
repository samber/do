package do

import "time"

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
