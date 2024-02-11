package do

import (
	"context"
	"fmt"
)

var _ Injector = (*virtualScope)(nil)

// virtualScope is a simple wrapper to Injector (Scope or RootScope or virtualScope) that
// contains the invoker name.
// It is used to track the dependency graph.
type virtualScope struct {
	self         Injector
	invokerChain []string
}

// pass through
func (s *virtualScope) ID() string                             { return s.self.ID() }
func (s *virtualScope) Name() string                           { return s.self.Name() }
func (s *virtualScope) Scope(name string) *Scope               { return s.self.Scope(name) }
func (s *virtualScope) RootScope() *RootScope                  { return s.self.RootScope() }
func (s *virtualScope) Ancestors() []*Scope                    { return s.self.Ancestors() }
func (s *virtualScope) Children() []*Scope                     { return s.self.Children() }
func (s *virtualScope) ChildByID(id string) (*Scope, bool)     { return s.self.ChildByID(id) }
func (s *virtualScope) ChildByName(name string) (*Scope, bool) { return s.self.ChildByName(name) }
func (s *virtualScope) ListProvidedServices() []EdgeService    { return s.self.ListProvidedServices() }
func (s *virtualScope) ListInvokedServices() []EdgeService     { return s.self.ListInvokedServices() }
func (s *virtualScope) HealthCheck() map[string]error          { return s.self.HealthCheck() }
func (s *virtualScope) HealthCheckWithContext(ctx context.Context) map[string]error {
	return s.self.HealthCheckWithContext(ctx)
}
func (s *virtualScope) Shutdown() *ShutdownErrors { return s.self.Shutdown() }
func (s *virtualScope) ShutdownWithContext(ctx context.Context) *ShutdownErrors {
	return s.self.ShutdownWithContext(ctx)
}
func (s *virtualScope) clone(r *RootScope, p *Scope) *Scope              { return s.self.clone(r, p) }
func (s *virtualScope) serviceExist(name string) bool                    { return s.self.serviceExist(name) }
func (s *virtualScope) serviceExistRec(name string) bool                 { return s.self.serviceExistRec(name) }
func (s *virtualScope) serviceGet(name string) (any, bool)               { return s.self.serviceGet(name) }
func (s *virtualScope) serviceGetRec(n string) (any, *Scope, bool)       { return s.self.serviceGetRec(n) }
func (s *virtualScope) serviceSet(name string, service any)              { s.self.serviceSet(name, service) }
func (s *virtualScope) serviceForEach(cb func(string, *Scope, any) bool) { s.self.serviceForEach(cb) }
func (s *virtualScope) serviceForEachRec(cb func(string, *Scope, any) bool) {
	s.self.serviceForEachRec(cb)
}
func (s *virtualScope) serviceHealthCheck(ctx context.Context, n string) error {
	return s.self.serviceHealthCheck(ctx, n)
}
func (s *virtualScope) serviceShutdown(ctx context.Context, name string) error {
	return s.self.serviceShutdown(ctx, name)
}
func (s *virtualScope) onServiceInvoke(name string) { s.self.onServiceInvoke(name) }

// detectCircularDependency checks for circular dependencies in the virtualScope.
// It returns ErrCircularDependency if the provided service name creates a circular dependency in the invoker chain.
func (s *virtualScope) detectCircularDependency(name string) error {
	if contains(s.invokerChain, name) {
		return fmt.Errorf("%w: %s", ErrCircularDependency, humanReadableInvokerChain(append(s.invokerChain, name)))
	}
	return nil
}

// addDependency adds a dependency to the DAG in the virtualScope.
func (s *virtualScope) addDependency(injector Injector, name string, serviceScope *Scope) {
	injector.RootScope().dag.addDependency(injector.ID(), injector.Name(), s.getLastInvokerName(), serviceScope.ID(), serviceScope.Name(), name)
}

// getLastInvokerName retrieves the last invoker name from the invoker chain in the virtualScope.
func (s *virtualScope) getLastInvokerName() string {
	if len(s.invokerChain) > 0 {
		return s.invokerChain[len(s.invokerChain)-1]
	}
	return ""
}
