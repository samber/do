package do

import (
	"context"
	"time"
)

// Injector is the main interface for dependency injection containers.
// It provides methods for service registration, resolution, lifecycle management,
// and scope hierarchy management.
//
// The Injector interface is implemented by both RootScope and Scope types,
// allowing for a consistent API across different levels of the scope hierarchy.
type Injector interface {
	// Public API methods for scope and service management

	// ID returns the unique identifier of the injector.
	ID() string

	// Name returns the human-readable name of the injector.
	Name() string

	// Scope creates a new child scope with the given name.
	// Optional package functions can be provided to execute during scope creation.
	Scope(string, ...func(Injector)) *Scope

	// RootScope returns the root scope of the injector hierarchy.
	RootScope() *RootScope

	// Ancestors returns the list of all parent scopes in order from immediate parent to root.
	Ancestors() []*Scope

	// Children returns the list of immediate child scopes.
	Children() []*Scope

	// ChildByID searches for a child scope by its unique ID across the entire scope hierarchy.
	ChildByID(string) (*Scope, bool)

	// ChildByName searches for a child scope by its name across the entire scope hierarchy.
	ChildByName(string) (*Scope, bool)

	// ListProvidedServices returns all services available in the current scope and all its ancestor scopes.
	ListProvidedServices() []ServiceDescription

	// ListInvokedServices returns only the services that have been actually invoked in the current scope and its ancestors.
	ListInvokedServices() []ServiceDescription

	// HealthCheck performs health checks on all services in the current scope and its ancestors.
	HealthCheck() map[string]error

	// HealthCheckWithContext performs health checks with context support for cancellation and timeouts.
	HealthCheckWithContext(context.Context) map[string]error

	// Shutdown gracefully shuts down the injector and all its descendant scopes.
	Shutdown() *ShutdownReport

	// ShutdownWithContext gracefully shuts down the injector and all its descendant scopes with context support.
	ShutdownWithContext(context.Context) *ShutdownReport

	// clone creates a deep copy of the injector with all its services and child scopes.
	clone(*RootScope, *Scope) *Scope

	// Internal service lifecycle methods

	// serviceExist checks if a service with the given name exists in the current scope.
	serviceExist(string) bool
	// serviceExistRec checks if a service with the given name exists in the current scope or any of its ancestor scopes.
	serviceExistRec(string) bool
	// serviceGet retrieves a service from the current scope by name.
	serviceGet(string) (any, bool)
	// serviceGetRec retrieves a service by name from the current scope or any of its ancestor scopes.
	serviceGetRec(string) (any, *Scope, bool)
	// serviceSet registers a service in the current scope with the given name.
	// Note: This method is not protected against double registration.
	serviceSet(string, any)
	// serviceForEach iterates over all services in the current scope.
	serviceForEach(func(string, *Scope, any) bool)
	// serviceForEachRec iterates over all services in the current scope and all ancestor scopes.
	serviceForEachRec(func(string, *Scope, any) bool)
	// serviceHealthCheck performs a health check on a specific service in the current scope.
	serviceHealthCheck(context.Context, string) error
	// serviceShutdown gracefully shuts down a specific service in the current scope.
	serviceShutdown(context.Context, string) error
	// onServiceInvoke is called whenever a service is invoked in this scope.
	onServiceInvoke(string)
}

// getInjectorOrDefault returns the provided injector if it's not nil, otherwise returns the default root scope.
// This function ensures that DI operations always have a valid injector to work with.
//
// Parameters:
//   - i: The injector to check, can be nil
//
// Returns the provided injector if not nil, or DefaultRootScope if nil.
func getInjectorOrDefault(i Injector) Injector {
	if i != nil {
		return i
	}

	return DefaultRootScope
}

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

	// HookBeforeHealthCheck is called before a service health check.
	// This hook can be used for cleanup preparation or logging.
	HookBeforeHealthCheck []func(scope *Scope, serviceName string)

	// HookAfterHealthCheck is called after a service is health checked.
	// This hook can be used for logging, metrics collection, or error handling.
	HookAfterHealthCheck []func(scope *Scope, serviceName string)

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
	StructTagKey          string
}

func (o *InjectorOpts) copy() *InjectorOpts {
	return &InjectorOpts{
		HookBeforeRegistration:   append([]func(*Scope, string){}, o.HookBeforeRegistration...),
		HookAfterRegistration:    append([]func(*Scope, string){}, o.HookAfterRegistration...),
		HookBeforeInvocation:     append([]func(*Scope, string){}, o.HookBeforeInvocation...),
		HookAfterInvocation:      append([]func(*Scope, string, error){}, o.HookAfterInvocation...),
		HookBeforeShutdown:       append([]func(*Scope, string){}, o.HookBeforeShutdown...),
		HookAfterShutdown:        append([]func(*Scope, string, error){}, o.HookAfterShutdown...),
		HookBeforeHealthCheck:    append([]func(*Scope, string){}, o.HookBeforeHealthCheck...),
		HookAfterHealthCheck:     append([]func(*Scope, string){}, o.HookAfterHealthCheck...),
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

func (o *InjectorOpts) onBeforeHealthCheck(scope *Scope, serviceName string) {
	for _, fn := range o.HookBeforeHealthCheck {
		fn(scope, serviceName)
	}
}

func (o *InjectorOpts) onAfterHealthCheck(scope *Scope, serviceName string) {
	for _, fn := range o.HookAfterHealthCheck {
		fn(scope, serviceName)
	}
}
