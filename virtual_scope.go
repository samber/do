package do

var _ Injector = (*virtualScope)(nil)

// virtualScope is a simple wrapper to Injector (Scope or RootScope or virtualScope) that
// contains the invoker name.
// It is used to track the dependency graph.
type virtualScope struct {
	self         Injector
	invokerChain []string
}

// pass through
func (s *virtualScope) ID() string                                 { return s.self.ID() }
func (s *virtualScope) Name() string                               { return s.self.Name() }
func (s *virtualScope) Scope(name string) *Scope                   { return s.self.Scope(name) }
func (s *virtualScope) RootScope() *RootScope                      { return s.self.RootScope() }
func (s *virtualScope) Ancestors() []*Scope                        { return s.self.Ancestors() }
func (s *virtualScope) Children() []*Scope                         { return s.self.Children() }
func (s *virtualScope) ChildByID(id string) (*Scope, bool)         { return s.self.ChildByID(id) }
func (s *virtualScope) ChildByName(name string) (*Scope, bool)     { return s.self.ChildByName(name) }
func (s *virtualScope) ListProvidedServices() []EdgeService        { return s.self.ListProvidedServices() }
func (s *virtualScope) ListInvokedServices() []EdgeService         { return s.self.ListInvokedServices() }
func (s *virtualScope) HealthCheck() map[string]error              { return s.self.HealthCheck() }
func (s *virtualScope) Shutdown() error                            { return s.self.Shutdown() }
func (s *virtualScope) clone(r *RootScope, p *Scope) *Scope        { return s.self.clone(r, p) }
func (s *virtualScope) serviceExist(name string) bool              { return s.self.serviceExist(name) }
func (s *virtualScope) serviceGet(name string) (any, bool)         { return s.self.serviceGet(name) }
func (s *virtualScope) serviceGetRec(n string) (any, *Scope, bool) { return s.self.serviceGetRec(n) }
func (s *virtualScope) serviceSet(name string, service any)        { s.self.serviceSet(name, service) }
func (s *virtualScope) serviceForEach(cb func(string, any))        { s.self.serviceForEach(cb) }
func (s *virtualScope) serviceHealthCheck(n string) error          { return s.self.serviceHealthCheck(n) }
func (s *virtualScope) serviceShutdown(name string) error          { return s.self.serviceShutdown(name) }
func (s *virtualScope) onServiceInvoke(name string)                { s.self.onServiceInvoke(name) }
