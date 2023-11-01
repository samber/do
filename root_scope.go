package do

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

const DefaultRootScopeName = "[root]"

var DefaultRootScope *RootScope = New()
var noOpLogf = func(format string, args ...any) {}

// New creates a new injector.
func New() *RootScope {
	return NewWithOpts(&InjectorOpts{})
}

// NewWithOpts creates a new injector with options.
func NewWithOpts(opts *InjectorOpts) *RootScope {
	if opts.Logf == nil {
		opts.Logf = noOpLogf
	}

	root := &RootScope{
		self: newScope(DefaultRootScopeName, nil, nil),
		opts: opts,
		dag:  newDAG(),
	}
	root.self.rootScope = root

	root.opts.Logf("DI: injector created")

	return root
}

var _ Injector = (*RootScope)(nil)

// RootScope is the first level of scope tree.
type RootScope struct {
	self *Scope

	opts *InjectorOpts

	dag *DAG
}

// pass through
func (s *RootScope) ID() string                             { return s.self.ID() }
func (s *RootScope) Name() string                           { return s.self.Name() }
func (s *RootScope) Scope(name string) *Scope               { return s.self.Scope(name) }
func (s *RootScope) RootScope() *RootScope                  { return s.self.RootScope() }
func (s *RootScope) Ancestors() []*Scope                    { return []*Scope{} }
func (s *RootScope) Children() []*Scope                     { return s.self.Children() }
func (s *RootScope) ChildByID(id string) (*Scope, bool)     { return s.self.ChildByID(id) }
func (s *RootScope) ChildByName(name string) (*Scope, bool) { return s.self.ChildByName(name) }
func (s *RootScope) ListProvidedServices() []EdgeService    { return s.self.ListProvidedServices() }
func (s *RootScope) ListInvokedServices() []EdgeService     { return s.self.ListInvokedServices() }
func (s *RootScope) HealthCheck() map[string]error          { return s.self.HealthCheck() }
func (s *RootScope) HealthCheckWithContext(ctx context.Context) map[string]error {
	return s.self.HealthCheckWithContext(ctx)
}
func (s *RootScope) Shutdown() error { return s.self.Shutdown() }
func (s *RootScope) ShutdownWithContext(ctx context.Context) error {
	return s.self.ShutdownWithContext(ctx)
}
func (s *RootScope) clone(root *RootScope, parent *Scope) *Scope      { return s.self.clone(root, parent) }
func (s *RootScope) serviceExist(name string) bool                    { return s.self.serviceExist(name) }
func (s *RootScope) serviceExistRec(name string) bool                 { return s.self.serviceExistRec(name) }
func (s *RootScope) serviceGet(name string) (any, bool)               { return s.self.serviceGet(name) }
func (s *RootScope) serviceGetRec(name string) (any, *Scope, bool)    { return s.self.serviceGetRec(name) }
func (s *RootScope) serviceSet(name string, service any)              { s.self.serviceSet(name, service) }
func (s *RootScope) serviceForEach(cb func(string, *Scope, any) bool) { s.self.serviceForEach(cb) }
func (s *RootScope) serviceForEachRec(cb func(string, *Scope, any) bool) {
	s.self.serviceForEachRec(cb)
}
func (s *RootScope) serviceHealthCheck(ctx context.Context, name string) error {
	return s.self.serviceHealthCheck(ctx, name)
}
func (s *RootScope) serviceShutdown(ctx context.Context, name string) error {
	return s.self.serviceShutdown(ctx, name)
}
func (s *RootScope) onServiceInvoke(name string) { s.self.onServiceInvoke(name) }

// Clone clones injector with provided services but not with invoked instances.
func (s *RootScope) Clone() *RootScope {
	return s.CloneWithOpts(s.opts)
}

// CloneWithOpts clones injector with provided services but not with invoked instances, with options.
func (s *RootScope) CloneWithOpts(opts *InjectorOpts) *RootScope {
	clone := NewWithOpts(opts)
	clone.self = s.clone(clone, nil)

	s.opts.Logf("DI: injector cloned")

	return clone
}

// ShutdownOnSIGTERM listens for sigterm signal in order to graceful stop service.
// It will block until receiving a sigterm signal.
func (s *RootScope) ShutdownOnSIGTERM() (os.Signal, error) {
	return s.ShutdownOnSignalsWithContext(context.Background(), syscall.SIGTERM)
}

// ShutdownOnSIGTERMWithContext listens for sigterm signal in order to graceful stop service.
// It will block until receiving a sigterm signal.
func (s *RootScope) ShutdownOnSIGTERMWithContext(ctx context.Context) (os.Signal, error) {
	return s.ShutdownOnSignalsWithContext(ctx, syscall.SIGTERM)
}

// ShutdownOnSIGTERMOrInterrupt listens for sigterm or interrupt signal in order to graceful stop service.
// It will block until receiving a sigterm signal.
func (s *RootScope) ShutdownOnSIGTERMOrInterrupt() (os.Signal, error) {
	return s.ShutdownOnSignalsWithContext(context.Background(), syscall.SIGTERM, os.Interrupt)
}

// ShutdownOnSIGTERMOrInterruptWithContext listens for sigterm or interrupt signal in order to graceful stop service.
// It will block until receiving a sigterm signal.
func (s *RootScope) ShutdownOnSIGTERMOrInterruptWithContext(ctx context.Context) (os.Signal, error) {
	return s.ShutdownOnSignalsWithContext(ctx, syscall.SIGTERM, os.Interrupt)
}

// ShutdownOnSignals listens for signals defined in signals parameter in order to graceful stop service.
// It will block until receiving any of these signal.
// If no signal is provided in signals parameter, syscall.SIGTERM will be added as default signal.
func (s *RootScope) ShutdownOnSignals(signals ...os.Signal) (os.Signal, error) {
	return s.ShutdownOnSignalsWithContext(context.Background(), signals...)
}

// ShutdownOnSignalsWithContext listens for signals defined in signals parameter in order to graceful stop service.
// It will block until receiving any of these signal.
// If no signal is provided in signals parameter, syscall.SIGTERM will be added as default signal.
func (s *RootScope) ShutdownOnSignalsWithContext(ctx context.Context, signals ...os.Signal) (os.Signal, error) {
	// Make sure there is at least syscall.SIGTERM as a signal
	if len(signals) < 1 {
		signals = append(signals, syscall.SIGTERM)
	}

	ch := make(chan os.Signal, 5)
	signal.Notify(ch, signals...)

	sig := <-ch
	signal.Stop(ch)
	close(ch)

	return sig, s.ShutdownWithContext(ctx)
}
