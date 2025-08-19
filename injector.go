package do

import "context"

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
	ListProvidedServices() []EdgeService

	// ListInvokedServices returns only the services that have been actually invoked in the current scope and its ancestors.
	ListInvokedServices() []EdgeService

	// HealthCheck performs health checks on all services in the current scope and its ancestors.
	HealthCheck() map[string]error

	// HealthCheckWithContext performs health checks with context support for cancellation and timeouts.
	HealthCheckWithContext(context.Context) map[string]error

	// Shutdown gracefully shuts down the injector and all its descendant scopes.
	Shutdown() *ShutdownErrors

	// ShutdownWithContext gracefully shuts down the injector and all its descendant scopes with context support.
	ShutdownWithContext(context.Context) *ShutdownErrors

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
