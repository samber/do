package do

import (
	"time"
)

// DefaultStructTagKey is the default tag key used for struct field injection.
// When using struct injection, fields can be tagged with `do:""` or `do:"service-name"`
// to specify which service should be injected.
const DefaultStructTagKey = "do"

// InjectorOpts contains all configuration options for the dependency injection container.
// These options control logging, hooks, health checks, and other behavioral aspects
// of the DI container.
type InjectorOpts struct {
	// HookBeforeRegistration is called before a service is registered in a scope.
	// This hook can be used for validation, logging, or other pre-registration tasks.
	HookBeforeRegistration []func(scope *Scope, serviceName string)

	// HookAfterRegistration is called after a service is successfully registered in a scope.
	// This hook can be used for logging, metrics collection, or other post-registration tasks.
	HookAfterRegistration []func(scope *Scope, serviceName string)

	// HookBeforeInvocation is called before a service is invoked (instantiated).
	// This hook can be used for logging, metrics, or other pre-invocation tasks.
	HookBeforeInvocation []func(scope *Scope, serviceName string)

	// HookAfterInvocation is called after a service is invoked, with the result error.
	// This hook can be used for logging, metrics collection, or error handling.
	HookAfterInvocation []func(scope *Scope, serviceName string, err error)

	// HookBeforeShutdown is called before a service is shut down.
	// This hook can be used for cleanup preparation or logging.
	HookBeforeShutdown []func(scope *Scope, serviceName string)

	// HookAfterShutdown is called after a service is shut down, with the result error.
	// This hook can be used for logging, metrics collection, or error handling.
	HookAfterShutdown []func(scope *Scope, serviceName string, err error)

	// Logf is the logging function used by the DI container for internal logging.
	// If not provided, no logging will occur. This function should handle the format
	// string and arguments similar to fmt.Printf.
	Logf func(format string, args ...any)

	// HealthCheckParallelism controls the number of concurrent health checks that can run simultaneously.
	// Default: all health checks run in parallel (unlimited).
	// Setting this to a positive number limits the concurrency for better resource management.
	HealthCheckParallelism uint

	// HealthCheckGlobalTimeout sets a global timeout for all health check operations.
	// This timeout applies to the entire health check process across all services.
	// Default: no timeout (health checks can run indefinitely).
	HealthCheckGlobalTimeout time.Duration

	// HealthCheckTimeout sets a timeout for individual service health checks.
	// This timeout applies to each service's health check method.
	// Default: no timeout (individual health checks can run indefinitely).
	HealthCheckTimeout time.Duration

	// StructTagKey specifies the tag key used for struct field injection.
	// Default: "do" (see DefaultStructTagKey constant).
	// This allows customization of the struct tag format for injection.
	StructTagKey string
}

func (o *InjectorOpts) copy() *InjectorOpts {
	return &InjectorOpts{
		HookBeforeRegistration:   o.HookBeforeRegistration,
		HookAfterRegistration:    o.HookAfterRegistration,
		HookBeforeInvocation:     o.HookBeforeInvocation,
		HookAfterInvocation:      o.HookAfterInvocation,
		HookBeforeShutdown:       o.HookBeforeShutdown,
		HookAfterShutdown:        o.HookAfterShutdown,
		Logf:                     o.Logf,
		HealthCheckParallelism:   o.HealthCheckParallelism,
		HealthCheckGlobalTimeout: o.HealthCheckGlobalTimeout,
		HealthCheckTimeout:       o.HealthCheckTimeout,
		StructTagKey:             o.StructTagKey,
	}
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
