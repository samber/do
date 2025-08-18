package do

import (
	"context"
	"fmt"
	"sync"
)

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

var _ Injector = (*Scope)(nil)

// Scope is a DI container. It must be created with injector.Scope("name") method.
type Scope struct {
	id          string            // immutable
	name        string            // immutable
	rootScope   *RootScope        // immutable
	parentScope *Scope            // immutable
	childScopes map[string]*Scope // append only

	mu       sync.RWMutex
	services map[string]any

	// Storing the invocation order is not needed anymore, but we keep it
	// for improved observability in unit tests.
	orderedInvocation      map[string]int // map is faster than slice
	orderedInvocationIndex int
}

// ID returns the unique identifier of the scope.
func (s *Scope) ID() string {
	return s.id
}

// Name returns the name of the scope.
func (s *Scope) Name() string {
	return s.name
}

// Scope creates a new child scope.
func (s *Scope) Scope(name string, packages ...func(Injector)) *Scope {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.childScopes[name]; ok {
		panic(fmt.Errorf("DI: scope `%s` has already been declared", name))
	}

	child := newScope(name, s.rootScope, s)
	s.childScopes[name] = child

	for _, pkg := range packages {
		pkg(child)
	}

	return child
}

// RootScope returns the root scope.
func (s *Scope) RootScope() *RootScope {
	return s.rootScope
}

// Ancestors returns the list of parent scopes recursively.
func (s *Scope) Ancestors() []*Scope {
	if s.parentScope == nil {
		return []*Scope{}
	}

	return append([]*Scope{s.parentScope}, s.parentScope.Ancestors()...)
}

// Children returns the list of immediate child scopes.
func (s *Scope) Children() []*Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scopes := []*Scope{}
	for _, scope := range s.childScopes {
		scopes = append(scopes, scope)
	}

	return scopes
}

// ChildByID returns the child scope recursively by its ID.
func (s *Scope) ChildByID(id string) (*Scope, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, scope := range s.childScopes {
		if scope.id == id {
			return scope, true
		}

		if child, ok := scope.ChildByID(id); ok {
			return child, true
		}
	}

	return nil, false
}

// ChildByName returns the child scope recursively by its name.
func (s *Scope) ChildByName(name string) (*Scope, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, scope := range s.childScopes {
		if scope.name == name {
			return scope, true
		}

		if child, ok := scope.ChildByName(name); ok {
			return child, true
		}
	}

	return nil, false
}

// ListProvidedServices returns the list of services provided by the scope.
func (s *Scope) ListProvidedServices() []EdgeService {
	s.mu.RLock()
	edges := mAp(keys(s.services), func(name string, _ int) EdgeService {
		return newEdgeService(s.id, s.name, name)
	})
	s.mu.RUnlock()

	for _, ancestor := range s.Ancestors() {
		edges = append(edges, ancestor.ListProvidedServices()...)
	}

	s.logf("exported list of services: %v", edges)

	return orderedUniq(edges)
}

// ListInvokedServices returns the list of services invoked by the scope.
func (s *Scope) ListInvokedServices() []EdgeService {
	s.mu.RLock()
	edges := mAp(keys(s.orderedInvocation), func(name string, _ int) EdgeService {
		return newEdgeService(s.id, s.name, name)
	})
	s.mu.RUnlock()

	for _, ancestor := range s.Ancestors() {
		edges = append(edges, ancestor.ListInvokedServices()...)
	}

	s.logf("exported list of invoked services: %v", edges)

	return orderedUniq(edges)
}

// HealthCheck returns the healthcheck results of the scope, in a map of service name to error.
func (s *Scope) HealthCheck() map[string]error {
	return s.HealthCheckWithContext(context.Background())
}

// HealthCheckWithContext returns the healthcheck results of the scope, in a map of service name to error.
func (s *Scope) HealthCheckWithContext(ctx context.Context) map[string]error {
	s.logf("requested healthcheck")

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

	s.logf("got healthcheck results: %v", results)

	return results
}

func (s *Scope) asyncHealthCheckWithContext(ctx context.Context) map[string]<-chan error {
	asyncResults := map[string]<-chan error{}

	s.mu.RLock()
	for _, name := range keys(s.services) {
		asyncResults[name] = s.rootScope.queueServiceHealthcheck(ctx, s, name)
	}
	s.mu.RUnlock()

	// @TODO: we should not check status of services that are not inherited (overriden in a child tree)
	for _, ancestor := range s.Ancestors() {
		heath := ancestor.asyncHealthCheckWithContext(ctx)
		for name, err := range heath {
			if _, ok := asyncResults[name]; !ok {
				asyncResults[name] = err
			}
		}
	}

	return asyncResults
}

// Shutdown shutdowns the scope and all its children.
func (s *Scope) Shutdown() *ShutdownErrors {
	return s.ShutdownWithContext(context.Background())
}

// ShutdownWithContext shutdowns the scope and all its children.
func (s *Scope) ShutdownWithContext(ctx context.Context) *ShutdownErrors {
	s.logf("requested shutdown")
	err1 := s.shutdownChildrenInParallel(ctx)
	err2 := s.shutdownServicesInParallel(ctx)
	s.logf("shutdowned services")

	err := mergeShutdownErrors(err1, err2)
	if err.Len() > 0 {
		return err
	}

	return nil
}

// shutdownChildrenInParallel runs a parallel shutdown of children scopes.
func (s *Scope) shutdownChildrenInParallel(ctx context.Context) *ShutdownErrors {
	s.mu.RLock()
	children := s.childScopes
	s.mu.RUnlock()

	errors := make([]*ShutdownErrors, len(children))

	var wg sync.WaitGroup
	for index, scope := range values(children) {
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
//
// We look for services having no dependents. Then we shutdown them.
// And repeat, until every scope services have been shutdown.
func (s *Scope) shutdownServicesInParallel(ctx context.Context) *ShutdownErrors {
	err := newShutdownErrors()

	listServices := func() []string {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return keys(s.services)
	}

	// var wg sync.WaitGroup

	for len(listServices()) > 0 {
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
			e := s.shutdownServicesWithoutDependenciesInParallel(ctx, services)
			err = mergeShutdownErrors(err, e)
		}
	}

	return err
}

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

func (s *Scope) clone(root *RootScope, parent *Scope) *Scope {
	clone := newScope(s.name, root, parent)

	s.mu.Lock()
	defer s.mu.Unlock()

	for name, serviceAny := range s.services {
		s.rootScope.opts.onBeforeRegistration(clone, name)

		if service, ok := serviceAny.(serviceClone); ok {
			clone.services[name] = service.clone()
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

func (s *Scope) serviceExist(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.services[name]
	return ok
}

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

func (s *Scope) serviceGet(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	svc, ok := s.services[name]
	return svc, ok
}

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

// serviceSet is not protected against double registration.
// Above layers should check if the service is already registered.
// It permits service override.
func (s *Scope) serviceSet(name string, service any) {
	s.RootScope().opts.onBeforeRegistration(s, name)

	s.mu.Lock()
	s.services[name] = service
	s.mu.Unlock()

	s.RootScope().opts.onAfterRegistration(s, name)
}

func (s *Scope) serviceForEach(cb func(name string, scope *Scope, service any) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, service := range s.services {
		keepGoing := cb(name, s, service)
		if !keepGoing {
			break
		}
	}
}

func (s *Scope) serviceForEachRec(cb func(name string, scope *Scope, service any) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, service := range s.services {
		keepGoing := cb(name, s, service)
		if !keepGoing {
			return
		}
	}

	if s.parentScope != nil {
		s.parentScope.serviceForEachRec(cb)
	}
}

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
		s.logf("requested healthcheck for service %s", name)

		// Timeout error is not triggered when the service is not an healthchecker.
		// If healthchecker does not support context.Timeout, error will be triggered raceWithTimeout().
		return raceWithTimeout(
			ctx,
			service.healthcheck,
		)
	}

	return nil
}

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

func (s *Scope) onServiceInvoke(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orderedInvocation[name]; !ok {
		s.orderedInvocation[name] = s.orderedInvocationIndex
		s.orderedInvocationIndex++
	}
}

func (s *Scope) logf(format string, args ...any) {
	format = fmt.Sprintf("DI <scope=%s>: %s", s.name, format)
	args = append([]any{s.name}, args...)
	s.RootScope().opts.Logf(format, args...)
}
