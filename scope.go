package do

import (
	"context"
	"fmt"
	"sync"
)

// newScope creates a new Scope instance with the provided parameters.
// This is an internal function used by the scope creation methods.
//
// Parameters:
//   - name: The human-readable name of the scope
//   - root: The root scope that this scope belongs to
//   - parent: The immediate parent scope, or nil for the root scope
//
// Returns a fully initialized Scope with all internal fields set to their default values.
func newScope(name string, root *RootScope, parent *Scope) *Scope {
	return &Scope{
		id:          must1(newUUID()),
		name:        name,
		rootScope:   root,
		parentScope: parent,
		childScopes: map[string]*Scope{},

		mu:       sync.RWMutex{},
		services: make(map[string]any),

		orderedInvocation:      map[string]int{},
		orderedInvocationIndex: 0,
	}
}

// Ensure Scope implements the Injector interface at compile time
var _ Injector = (*Scope)(nil)

// Scope represents a dependency injection container that can contain services
// and child scopes. Scopes form a hierarchical tree structure where child scopes
// can access services from their parent scopes, but not vice versa.
//
// Key features:
//   - Hierarchical service resolution (child scopes can access parent services)
//   - Isolated service registration (services in child scopes don't affect parents)
//   - Thread-safe operations
//   - Service lifecycle management (health checks, shutdown)
//   - Observability and debugging support
type Scope struct {
	id          string            // Unique identifier for the scope (immutable)
	name        string            // Human-readable name for the scope (immutable)
	rootScope   *RootScope        // Reference to the root scope (immutable)
	parentScope *Scope            // Reference to the immediate parent scope (immutable)
	childScopes map[string]*Scope // Map of child scopes (append only)

	mu       sync.RWMutex   // Mutex for thread-safe operations
	services map[string]any // Map of registered services

	// Storing the invocation order is not needed anymore, but we keep it
	// for improved observability in unit tests.
	orderedInvocation      map[string]int // Map tracking service invocation order (faster than slice)
	orderedInvocationIndex int            // Counter for tracking invocation order
}

// ID returns the unique identifier of the scope.
// This ID is generated using UUID and is immutable throughout the scope's lifetime.
func (s *Scope) ID() string {
	return s.id
}

// Name returns the human-readable name of the scope.
// This name is provided when creating the scope and is immutable.
func (s *Scope) Name() string {
	return s.name
}

// Scope creates a new child scope with the given name.
// Child scopes inherit access to services from their parent scopes,
// but services registered in child scopes are not accessible to parents.
//
// Parameters:
//   - name: The name for the new child scope (must be unique within the parent)
//   - packages: Optional package functions to execute in the new scope
//
// Returns the newly created child scope.
//
// Panics if a scope with the same name already exists in the parent.
func (s *Scope) Scope(name string, packages ...func(Injector)) *Scope {
	s.mu.Lock()

	if _, ok := s.childScopes[name]; ok {
		s.mu.Unlock()
		panic(fmt.Errorf("DI: scope `%s` has already been declared", name))
	}

	child := newScope(name, s.rootScope, s)
	s.childScopes[name] = child

	s.mu.Unlock()

	// Execute any package functions in the new scope
	for _, pkg := range packages {
		pkg(child)
	}

	return child
}

// RootScope returns the root scope of the scope hierarchy.
// All scopes in a hierarchy share the same root scope, regardless of their depth.
func (s *Scope) RootScope() *RootScope {
	return s.rootScope
}

// Ancestors returns the list of all parent scopes in order from immediate parent to root.
// This is useful for understanding the scope hierarchy and for operations that need
// to traverse up the scope tree.
//
// Returns an empty slice for the root scope, and a slice of parent scopes
// for child scopes, ordered from immediate parent to root.
func (s *Scope) Ancestors() []*Scope {
	if s.parentScope == nil {
		return []*Scope{}
	}

	return append([]*Scope{s.parentScope}, s.parentScope.Ancestors()...)
}

// Children returns the list of immediate child scopes.
// This method only returns direct children, not grandchildren or deeper descendants.
//
// Returns a slice of child scopes. The order is not guaranteed to be stable.
func (s *Scope) Children() []*Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scopes := []*Scope{}
	for _, scope := range s.childScopes {
		scopes = append(scopes, scope)
	}

	return scopes
}

// ChildByID searches for a child scope by its unique ID across the entire scope hierarchy.
// This method performs a recursive search through all descendant scopes.
//
// Parameters:
//   - id: The unique ID of the scope to find
//
// Returns the found scope and true if found, or nil and false if not found.
func (s *Scope) ChildByID(id string) (*Scope, bool) {
	s.mu.RLock()
	children := make([]*Scope, 0, len(s.childScopes))
	for _, c := range s.childScopes {
		children = append(children, c)
	}
	s.mu.RUnlock()

	for _, scope := range children {
		if scope.id == id {
			return scope, true
		}

		if child, ok := scope.ChildByID(id); ok {
			return child, true
		}
	}

	return nil, false
}

// ChildByName searches for a child scope by its name across the entire scope hierarchy.
// This method performs a recursive search through all descendant scopes.
//
// Parameters:
//   - name: The name of the scope to find
//
// Returns the found scope and true if found, or nil and false if not found.
func (s *Scope) ChildByName(name string) (*Scope, bool) {
	s.mu.RLock()
	children := make([]*Scope, 0, len(s.childScopes))
	for _, c := range s.childScopes {
		children = append(children, c)
	}
	s.mu.RUnlock()

	for _, scope := range children {
		if scope.name == name {
			return scope, true
		}

		if child, ok := scope.ChildByName(name); ok {
			return child, true
		}
	}

	return nil, false
}

// ListProvidedServices returns all services available in the current scope and all its ancestor scopes.
// This provides a complete view of the service hierarchy, showing all services
// that can be accessed from the current scope.
//
// Returns a slice of EdgeService objects representing all available services,
// including those inherited from parent scopes.
func (s *Scope) ListProvidedServices() []EdgeService {
	services := []EdgeService{}

	// Add services from current scope
	s.mu.RLock()
	for name := range s.services {
		services = append(services, newEdgeService(s.id, s.name, name))
	}
	s.mu.RUnlock()

	// Add services from ancestor scopes
	for _, ancestor := range s.Ancestors() {
		services = append(services, ancestor.ListProvidedServices()...)
	}

	services = orderedUniq(services)

	s.logf("exported list of services: %v", services)

	return services
}

// ListInvokedServices returns only the services that have been actually invoked
// (instantiated) in the current scope and its ancestors.
// This is useful for understanding which services are actually being used
// and for debugging dependency issues.
//
// Returns a slice of EdgeService objects representing only the invoked services.
func (s *Scope) ListInvokedServices() []EdgeService {
	services := []EdgeService{}

	// Add invoked services from current scope
	s.mu.RLock()
	for name := range s.orderedInvocation {
		services = append(services, newEdgeService(s.id, s.name, name))
	}
	s.mu.RUnlock()

	// Add invoked services from ancestor scopes
	for _, ancestor := range s.Ancestors() {
		services = append(services, ancestor.ListInvokedServices()...)
	}

	services = orderedUniq(services)

	s.logf("exported list of invoked services: %v", services)

	return services
}

// HealthCheck performs health checks on all services in the current scope and its ancestors
// that implement the Healthchecker interface.
//
// Returns a map of service names to error values. A nil error indicates a successful health check.
func (s *Scope) HealthCheck() map[string]error {
	return s.HealthCheckWithContext(context.Background())
}

// HealthCheckWithContext performs health checks on all services in the current scope and its ancestors
// that implement the Healthchecker interface, with context support for cancellation and timeouts.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a map of service names to error values. A nil error indicates a successful health check.
func (s *Scope) HealthCheckWithContext(ctx context.Context) map[string]error {
	s.logf("requested health check")

	if s.rootScope.opts.HealthCheckGlobalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.rootScope.opts.HealthCheckGlobalTimeout)
		defer cancel()
	}

	results := map[string]error{}

	asyncResults := s.asyncHealthCheckWithContext(ctx)
	for name, err := range asyncResults {
		results[name] = <-err
	}

	s.logf("got health check results: %v", results)

	return results
}

func (s *Scope) asyncHealthCheckWithContext(ctx context.Context) map[string]<-chan error {
	asyncResults := map[string]<-chan error{}

	s.mu.RLock()
	for name := range s.services {
		asyncResults[name] = s.rootScope.queueServiceHealthcheck(ctx, s, name)
	}
	s.mu.RUnlock()

	// @TODO: We should not check the status of services that are not inherited (overridden in a child tree)
	for _, ancestor := range s.Ancestors() {
		health := ancestor.asyncHealthCheckWithContext(ctx)
		for name, err := range health {
			if _, ok := asyncResults[name]; !ok {
				asyncResults[name] = err
			}
		}
	}

	return asyncResults
}

// Shutdown gracefully shuts down the scope and all its children.
// This method calls ShutdownWithContext with a background context.
//
// Returns a ShutdownErrors object containing any errors that occurred during shutdown.
func (s *Scope) Shutdown() *ShutdownErrors {
	return s.ShutdownWithContext(context.Background())
}

// ShutdownWithContext gracefully shuts down the scope and all its children with context support.
// This method performs shutdown operations in parallel for better performance.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a ShutdownErrors object containing any errors that occurred during shutdown.
func (s *Scope) ShutdownWithContext(ctx context.Context) *ShutdownErrors {
	s.logf("requested shutdown")
	err1 := s.shutdownChildrenInParallel(ctx)
	err2 := s.shutdownServicesInParallel(ctx)
	s.logf("shut down services")

	err := mergeShutdownErrors(err1, err2)
	if err.Len() > 0 {
		return err
	}

	return nil
}

// shutdownChildrenInParallel runs a parallel shutdown of children scopes.
// This method shuts down all child scopes concurrently and then removes them
// from the scope hierarchy.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a ShutdownErrors object containing any errors from child scope shutdowns.
func (s *Scope) shutdownChildrenInParallel(ctx context.Context) *ShutdownErrors {
	// Snapshot children under lock
	s.mu.RLock()
	children := make([]*Scope, 0, len(s.childScopes))
	for _, c := range s.childScopes {
		children = append(children, c)
	}
	s.mu.RUnlock()

	errors := make([]*ShutdownErrors, len(children))

	var wg sync.WaitGroup
	for index, scope := range children {
		wg.Add(1)

		go func(s *Scope, i int) {
			errors[i] = s.ShutdownWithContext(ctx)
			wg.Done()
		}(scope, index)
	}
	wg.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.childScopes = make(map[string]*Scope) // scopes are removed from DI container
	return mergeShutdownErrors(errors...)
}

// shutdownServicesInParallel runs a parallel shutdown of scope services.
// This method implements a dependency-aware shutdown algorithm that shuts down
// services in the correct order to avoid dependency issues.
//
// We look for services having no dependents. Then we shutdown them.
// And repeat, until every scope services have been shutdown.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a ShutdownErrors object containing any errors from service shutdowns.
func (s *Scope) shutdownServicesInParallel(ctx context.Context) *ShutdownErrors {
	err := newShutdownErrors()

	listServices := func() []string {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return keys(s.services)
	}

	for len(listServices()) > 0 {
		if ctx.Err() != nil {
			break
		}
		services := listServices()
		servicesToShutdown := []string{}

		// loop over the service that have not been shutdown already
		for _, name := range services {
			// Check the service has no dependents (dependencies allowed here).
			// Services having dependents must be shutdown first.
			// The next iteration will shutdown current service.
			_, dependents := s.rootScope.dag.explainService(s.id, s.name, name)
			if len(dependents) == 0 {
				servicesToShutdown = append(servicesToShutdown, name)
			}
		}

		if len(servicesToShutdown) > 0 {
			e := s.shutdownServicesWithoutDependenciesInParallel(ctx, servicesToShutdown)
			err = mergeShutdownErrors(err, e)
		} else {
			// In this branch, we expect that there is a circular dependency. We shutdown all services, without taking care of order.
			// This is a fallback mechanism to ensure all services are eventually shut down.
			e := s.shutdownServicesWithoutDependenciesInParallel(ctx, services)
			err = mergeShutdownErrors(err, e)
		}
	}

	return err
}

// shutdownServicesWithoutDependenciesInParallel shuts down multiple services concurrently
// without considering dependency order. This method is used when services have no dependents
// or when handling circular dependencies.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - serviceNames: List of service names to shut down
//
// Returns a ShutdownErrors object containing any errors from the shutdown operations.
func (s *Scope) shutdownServicesWithoutDependenciesInParallel(ctx context.Context, serviceNames []string) *ShutdownErrors {
	if len(serviceNames) == 0 {
		return nil
	}

	err := newShutdownErrors()
	mu := sync.Mutex{}

	var wg sync.WaitGroup
	wg.Add(len(serviceNames))

	for _, name := range serviceNames {
		go func(n string) {
			e := s.serviceShutdown(ctx, n)

			mu.Lock()
			err.Add(s.id, s.name, n, e)
			mu.Unlock()

			wg.Done()
		}(name)
	}

	wg.Wait()

	return err
}

// clone creates a deep copy of the scope with all its services and child scopes.
// This method is used for scope isolation and testing scenarios where you need
// a complete copy of the scope hierarchy.
//
// Parameters:
//   - root: The root scope for the cloned scope hierarchy
//   - parent: The parent scope for the cloned scope
//
// Returns a new Scope instance that is a deep copy of the original.
func (s *Scope) clone(root *RootScope, parent *Scope) *Scope {
	clone := newScope(s.name, root, parent)

	s.mu.Lock()
	defer s.mu.Unlock()

	for name, serviceAny := range s.services {
		s.rootScope.opts.onBeforeRegistration(clone, name)

		if service, ok := serviceAny.(serviceClone); ok {
			clone.services[name] = service.clone(clone)
		} else {
			clone.services[name] = service
		}

		s.rootScope.opts.onAfterRegistration(clone, name)
	}

	for name, index := range s.childScopes {
		clone.childScopes[name] = index.clone(root, clone)
	}

	s.logf("injector cloned")

	return clone
}

/**********************************
 *        Service lifecycle       *
 **********************************/

// serviceExist checks if a service with the given name exists in the current scope.
// This method only checks the current scope, not parent scopes.
//
// Parameters:
//   - name: The name of the service to check
//
// Returns true if the service exists in the current scope, false otherwise.
func (s *Scope) serviceExist(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.services[name]
	return ok
}

// serviceExistRec checks if a service with the given name exists in the current scope
// or any of its ancestor scopes. This method performs a recursive search up the scope hierarchy.
//
// Parameters:
//   - name: The name of the service to check
//
// Returns true if the service exists in the current scope or any ancestor scope, false otherwise.
func (s *Scope) serviceExistRec(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.services[name]
	if ok {
		return ok
	}

	if s.parentScope == nil {
		return false
	}

	return s.parentScope.serviceExistRec(name)
}

// serviceGet retrieves a service from the current scope by name.
// This method only searches in the current scope, not parent scopes.
//
// Parameters:
//   - name: The name of the service to retrieve
//
// Returns the service instance and true if found, or nil and false if not found.
func (s *Scope) serviceGet(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, ok := s.services[name]
	return svc, ok
}

// serviceGetRec retrieves a service by name from the current scope or any of its ancestor scopes.
// This method performs a recursive search up the scope hierarchy and returns the scope
// where the service was found.
//
// Parameters:
//   - name: The name of the service to retrieve
//
// Returns the service instance, the scope where it was found, and true if found,
// or nil, nil, and false if not found.
func (s *Scope) serviceGetRec(name string) (any, *Scope, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, ok := s.services[name]
	if ok {
		return svc, s, ok
	}

	if s.parentScope == nil {
		return nil, nil, false
	}

	return s.parentScope.serviceGetRec(name)
}

// serviceSet registers a service in the current scope with the given name.
// This method is not protected against double registration - the calling layer
// should check if the service is already registered. It permits service override.
//
// Parameters:
//   - name: The name to register the service under
//   - service: The service instance to register
func (s *Scope) serviceSet(name string, service any) {
	s.RootScope().opts.onBeforeRegistration(s, name)

	s.mu.Lock()
	s.services[name] = service
	s.mu.Unlock()

	s.RootScope().opts.onAfterRegistration(s, name)
}

// serviceForEach iterates over all services in the current scope and calls the provided callback
// for each service. The iteration stops if the callback returns false.
//
// Parameters:
//   - cb: Callback function that receives the service name, scope, and service instance.
//     Return true to continue iteration, false to stop.
func (s *Scope) serviceForEach(cb func(name string, scope *Scope, service any) bool) {
	// Take a snapshot under read lock to avoid iterating a map while it may be mutated
	s.mu.RLock()
	snapshot := make([]struct {
		name string
		svc  any
	}, 0, len(s.services))
	for name, service := range s.services {
		snapshot = append(snapshot, struct {
			name string
			svc  any
		}{name: name, svc: service})
	}
	s.mu.RUnlock()

	for _, item := range snapshot {
		keepGoing := cb(item.name, s, item.svc)
		if !keepGoing {
			break
		}
	}
}

// serviceForEachRec iterates over all services in the current scope and all ancestor scopes,
// calling the provided callback for each service. The iteration stops if the callback returns false.
//
// Parameters:
//   - cb: Callback function that receives the service name, scope, and service instance.
//     Return true to continue iteration, false to stop.
func (s *Scope) serviceForEachRec(cb func(name string, scope *Scope, service any) bool) {
	// Snapshot current services and parent under read lock to avoid deadlocks and map races
	s.mu.RLock()
	snapshot := make([]struct {
		name string
		svc  any
	}, 0, len(s.services))
	for name, service := range s.services {
		snapshot = append(snapshot, struct {
			name string
			svc  any
		}{name: name, svc: service})
	}
	parent := s.parentScope // immutable, but we copy its reference for safety
	s.mu.RUnlock()

	for _, item := range snapshot {
		keepGoing := cb(item.name, s, item.svc)
		if !keepGoing {
			return
		}
	}

	if parent != nil {
		parent.serviceForEachRec(cb)
	}
}

// serviceHealthCheck performs a health check on a specific service in the current scope.
// This method checks if the service implements the Healthchecker interface and calls
// its health check method if available.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - name: The name of the service to health check
//
// Returns an error if the health check fails, or nil if the service is healthy
// or doesn't implement the Healthchecker interface.
func (s *Scope) serviceHealthCheck(ctx context.Context, name string) error {
	s.mu.RLock()

	serviceAny, ok := s.services[name]
	if !ok {
		s.mu.RUnlock()
		return serviceNotFound(s, ErrServiceNotFound, []string{name})
	}

	s.mu.RUnlock()

	service, ok := serviceAny.(serviceHealthcheck)
	if ok {
		s.logf("requested health check for service %s", name)

		// A timeout error is not triggered when the service is not a healthchecker.
		// If the healthchecker does not support context.Timeout, the error will be triggered by raceWithTimeout().
		return raceWithTimeout(
			ctx,
			service.healthcheck,
		)
	}

	return nil
}

// serviceShutdown gracefully shuts down a specific service in the current scope.
// This method checks if the service implements the Shutdowner interface and calls
// its shutdown method. The service is then removed from the scope.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - name: The name of the service to shut down
//
// Returns an error if the shutdown fails, or nil if the shutdown was successful.
// Panics if the service doesn't implement the Shutdowner interface.
func (s *Scope) serviceShutdown(ctx context.Context, name string) error {
	s.mu.RLock()
	serviceAny, ok := s.services[name]
	s.mu.RUnlock()

	if !ok {
		return serviceNotFound(s, ErrServiceNotFound, []string{name})
	}

	var err error

	service, ok := serviceAny.(serviceShutdown)
	if ok {
		s.logf("requested shutdown for service %s", name)

		s.RootScope().opts.onBeforeShutdown(s, name)
		err = service.shutdown(ctx)
		s.RootScope().opts.onAfterShutdown(s, name, err)
	} else {
		panic(fmt.Errorf("DI: service `%s` is not shutdowner", name))
	}

	s.mu.Lock()
	delete(s.services, name) // service is removed from DI container
	delete(s.orderedInvocation, name)
	s.RootScope().dag.removeService(s.id, s.name, name)
	s.mu.Unlock()

	return err
}

/**********************************
 *             Hooks              *
 **********************************/

// onServiceInvoke is called whenever a service is invoked in this scope.
// This method tracks the order of service invocations for observability and debugging purposes.
//
// Parameters:
//   - name: The name of the service that was invoked
func (s *Scope) onServiceInvoke(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orderedInvocation[name]; !ok {
		s.orderedInvocation[name] = s.orderedInvocationIndex
		s.orderedInvocationIndex++
	}
}

// logf logs a formatted message with the scope name prefix.
// This method is used internally for consistent logging across the scope.
//
// Parameters:
//   - format: The format string for the log message
//   - args: Arguments to format the message with
func (s *Scope) logf(format string, args ...any) {
	format = fmt.Sprintf("DI <scope=%s>: %s", s.name, format)
	s.RootScope().opts.Logf(format, args...)
}
