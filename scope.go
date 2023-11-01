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

	// It should be a graph instead of simple ordered list.
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
func (s *Scope) Scope(name string) *Scope {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.childScopes[name]; ok {
		panic(fmt.Errorf("DI: scope `%s` has already been declared", name))
	}

	child := newScope(name, s.rootScope, s)
	s.childScopes[name] = child

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
	results := map[string]error{}

	s.mu.RLock()
	for _, name := range keys(s.services) {
		results[name] = s.serviceHealthCheck(ctx, name)
	}
	s.mu.RUnlock()

	s.logf("requested healthcheck")

	// @TODO: we should not check status of services that are not inherited (overriden in a child tree)
	for _, ancestor := range s.Ancestors() {
		heath := ancestor.HealthCheck()
		for name, err := range heath {
			if _, ok := results[name]; !ok {
				results[name] = err
			}
		}
	}

	s.logf("got healthcheck results: %v", results)

	return results
}

// Shutdown shutdowns the scope and all its children.
func (s *Scope) Shutdown() error {
	return s.ShutdownWithContext(context.Background())
}

// ShutdownWithContext shutdowns the scope and all its children.
func (s *Scope) ShutdownWithContext(ctx context.Context) error {
	s.mu.RLock()
	children := s.childScopes
	invocations := invertMap(s.orderedInvocation)
	s.mu.RUnlock()

	s.logf("requested shutdown")

	// first shutdown children
	for k, child := range children {
		err := child.Shutdown()
		if err != nil {
			return err
		}

		s.mu.Lock()
		delete(s.childScopes, k)
		s.mu.Unlock()
	}

	// then shutdown scope services
	for index := s.orderedInvocationIndex; index >= 0; index-- {
		name, ok := invocations[index]
		if !ok {
			continue
		}

		err := s.serviceShutdown(ctx, name)
		if err != nil {
			return err
		}
	}

	s.logf("shutdowned services")

	return nil
}

func (s *Scope) clone(root *RootScope, parent *Scope) *Scope {
	clone := newScope(s.name, root, parent)

	s.mu.Lock()
	defer s.mu.Unlock()

	for name, serviceAny := range s.services {
		if service, ok := serviceAny.(serviceClone); ok {
			clone.services[name] = service.clone()
		} else {
			clone.services[name] = service
		}
		defer clone.onServiceRegistration(name)
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

func (s *Scope) serviceSet(name string, service any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.services[name] = service

	// defering hook call will unlock mutex
	defer s.onServiceRegistration(name)
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
		return serviceNotFound(s, name)
	}

	s.mu.RUnlock()

	service, ok := serviceAny.(serviceHealthcheck)
	if ok {
		s.logf("requested healthcheck for service %s", name)

		err := service.healthcheck(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Scope) serviceShutdown(ctx context.Context, name string) error {
	s.mu.RLock()

	serviceAny, ok := s.services[name]
	if !ok {
		s.mu.RUnlock()
		return serviceNotFound(s, name)
	}

	s.mu.RUnlock()

	service, ok := serviceAny.(serviceShutdown)
	if ok {
		s.logf("requested shutdown for service %s", name)

		err := service.shutdown(ctx)
		if err != nil {
			return err
		}
	} else {
		panic(fmt.Errorf("DI: service `%s` is not shutdowner", name))
	}

	s.mu.Lock()
	delete(s.services, name)
	delete(s.orderedInvocation, name)
	s.mu.Unlock()

	s.onServiceShutdown(name)

	return nil
}

/**********************************
 *             Hooks              *
 **********************************/

func (s *Scope) onServiceRegistration(name string) {
	root := s.RootScope()
	if root == nil {
		return
	}

	if root.opts.HookAfterRegistration != nil {
		root.opts.HookAfterRegistration(s, name)
	}
}

func (s *Scope) onServiceInvoke(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orderedInvocation[name]; !ok {
		s.orderedInvocation[name] = s.orderedInvocationIndex
		s.orderedInvocationIndex++
	}
}

func (s *Scope) onServiceShutdown(name string) {
	root := s.RootScope()
	if root == nil {
		return
	}

	if root.opts.HookAfterShutdown != nil {
		root.opts.HookAfterShutdown(s, name)
	}
}

func (s *Scope) logf(format string, args ...any) {
	root := s.RootScope()
	if root == nil {
		return
	}

	format = fmt.Sprintf("DI <scope=%s>: %s", s.name, format)
	args = append([]any{s.name}, args...)

	root.opts.Logf(format, args...)
}
