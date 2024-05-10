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
func New(packages ...func(Injector)) *RootScope {
	return NewWithOpts(&InjectorOpts{}, packages...)
}

// NewWithOpts creates a new injector with options.
func NewWithOpts(opts *InjectorOpts, packages ...func(Injector)) *RootScope {
	if opts.Logf == nil {
		opts.Logf = noOpLogf
	}

	if opts.HookBeforeRegistration == nil {
		opts.HookBeforeRegistration = []func(*Scope, string){}
	}
	if opts.HookAfterRegistration == nil {
		opts.HookAfterRegistration = []func(*Scope, string){}
	}
	if opts.HookBeforeInvocation == nil {
		opts.HookBeforeInvocation = []func(*Scope, string){}
	}
	if opts.HookAfterInvocation == nil {
		opts.HookAfterInvocation = []func(*Scope, string, error){}
	}
	if opts.HookBeforeShutdown == nil {
		opts.HookBeforeShutdown = []func(*Scope, string){}
	}
	if opts.HookAfterShutdown == nil {
		opts.HookAfterShutdown = []func(*Scope, string, error){}
	}

	root := &RootScope{
		self:            newScope(DefaultRootScopeName, nil, nil),
		opts:            opts,
		dag:             newDAG(),
		healthCheckPool: nil,
	}
	root.self.rootScope = root

	if opts.HealthCheckParallelism > 0 {
		root.healthCheckPool = newJobPool[error](opts.HealthCheckParallelism)
		root.healthCheckPool.start()
	}

	root.opts.Logf("DI: injector created")

	for _, pkg := range packages {
		pkg(root)
	}

	return root
}

var _ Injector = (*RootScope)(nil)

// RootScope is the first level of scope tree.
type RootScope struct {
	self            *Scope
	opts            *InjectorOpts
	dag             *DAG
	healthCheckPool *jobPool[error]
}

// pass through
func (s *RootScope) ID() string                                    { return s.self.ID() }
func (s *RootScope) Name() string                                  { return s.self.Name() }
func (s *RootScope) Scope(name string, p ...func(Injector)) *Scope { return s.self.Scope(name, p...) }
func (s *RootScope) RootScope() *RootScope                         { return s.self.RootScope() }
func (s *RootScope) Ancestors() []*Scope                           { return []*Scope{} }
func (s *RootScope) Children() []*Scope                            { return s.self.Children() }
func (s *RootScope) ChildByID(id string) (*Scope, bool)            { return s.self.ChildByID(id) }
func (s *RootScope) ChildByName(name string) (*Scope, bool)        { return s.self.ChildByName(name) }
func (s *RootScope) ListProvidedServices() []EdgeService           { return s.self.ListProvidedServices() }
func (s *RootScope) ListInvokedServices() []EdgeService            { return s.self.ListInvokedServices() }
func (s *RootScope) HealthCheck() map[string]error                 { return s.self.HealthCheck() }
func (s *RootScope) HealthCheckWithContext(ctx context.Context) map[string]error {
	return s.self.HealthCheckWithContext(ctx)
}
func (s *RootScope) Shutdown() *ShutdownErrors { return s.ShutdownWithContext(context.Background()) }
func (s *RootScope) ShutdownWithContext(ctx context.Context) *ShutdownErrors {
	defer func() {
		if s.healthCheckPool != nil {
			s.healthCheckPool.stop()
			s.healthCheckPool = nil
		}
	}()

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

func (s *RootScope) queueServiceHealthcheck(ctx context.Context, scope *Scope, serviceName string) <-chan error {
	cancel := func() {}
	if s.opts.HealthCheckTimeout > 0 {
		// `ctx` might already contain a timeout, but we add another one
		ctx, cancel = context.WithTimeout(ctx, s.opts.HealthCheckTimeout)
	}

	// when no pooling policy has been defined
	if s.opts.HealthCheckParallelism == 0 || s.healthCheckPool == nil {
		err := make(chan error, 1) // a single message will be sent (nil or error)

		go func() {
			defer cancel()

			ctx.Done()
			err <- scope.serviceHealthCheck(ctx, serviceName)
			close(err)
		}()

		return err
	}

	// delegate execution to the healthcheck pool
	return s.healthCheckPool.rpc(func() error {
		defer cancel()
		return scope.serviceHealthCheck(ctx, serviceName)
	})
}

/**
 * RootScope stuff
 */

// AddBeforeRegistrationHook adds a hook that will be called before a service is registered.
func (s *RootScope) AddBeforeRegistrationHook(hook func(*Scope, string)) {
	s.opts.HookBeforeRegistration = append(s.opts.HookBeforeRegistration, hook)
}

// AddAfterRegistrationHook adds a hook that will be called after a service is registered.
func (s *RootScope) AddAfterRegistrationHook(hook func(*Scope, string)) {
	s.opts.HookAfterRegistration = append(s.opts.HookAfterRegistration, hook)
}

// AddBeforeInvocationHook adds a hook that will be called before a service is invoked.
func (s *RootScope) AddBeforeInvocationHook(hook func(*Scope, string)) {
	s.opts.HookBeforeInvocation = append(s.opts.HookBeforeInvocation, hook)
}

// AddAfterInvocationHook adds a hook that will be called after a service is invoked.
func (s *RootScope) AddAfterInvocationHook(hook func(*Scope, string, error)) {
	s.opts.HookAfterInvocation = append(s.opts.HookAfterInvocation, hook)
}

// AddBeforeShutdownHook adds a hook that will be called before a service is shutdown.
func (s *RootScope) AddBeforeShutdownHook(hook func(*Scope, string)) {
	s.opts.HookBeforeShutdown = append(s.opts.HookBeforeShutdown, hook)
}

// AddAfterShutdownHook adds a hook that will be called after a service is shutdown.
func (s *RootScope) AddAfterShutdownHook(hook func(*Scope, string, error)) {
	s.opts.HookAfterShutdown = append(s.opts.HookAfterShutdown, hook)
}

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

// ShutdownOnSignals listens for signals defined in signals parameter in order to graceful stop service.
// It will block until receiving any of these signal.
// If no signal is provided in signals parameter, syscall.SIGTERM and os.Interrupt will be added as default signal.
func (s *RootScope) ShutdownOnSignals(signals ...os.Signal) (os.Signal, *ShutdownErrors) {
	return s.ShutdownOnSignalsWithContext(context.Background(), signals...)
}

// ShutdownOnSignalsWithContext listens for signals defined in signals parameter in order to graceful stop service.
// It will block until receiving any of these signal.
// If no signal is provided in signals parameter, syscall.SIGTERM and os.Interrupt will be added as default signal.
func (s *RootScope) ShutdownOnSignalsWithContext(ctx context.Context, signals ...os.Signal) (os.Signal, *ShutdownErrors) {
	// Make sure there is at least syscall.SIGTERM and os.Interrupt as a signal
	if len(signals) < 1 {
		signals = append(signals, syscall.SIGTERM, os.Interrupt)
	}

	ch := make(chan os.Signal, 5)
	signal.Notify(ch, signals...)

	sig := <-ch
	signal.Stop(ch)
	close(ch)

	return sig, s.ShutdownWithContext(ctx)
}
