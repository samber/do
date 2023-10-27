package do

import "context"

type Injector interface {
	// api
	ID() string
	Name() string
	Scope(string) *Scope
	RootScope() *RootScope
	Ancestors() []*Scope
	Children() []*Scope
	ChildByID(string) (*Scope, bool)
	ChildByName(string) (*Scope, bool)
	ListProvidedServices() []EdgeService
	ListInvokedServices() []EdgeService
	HealthCheck() map[string]error
	HealthCheckWithContext(context.Context) map[string]error
	Shutdown() error
	ShutdownWithContext(context.Context) error
	clone(*RootScope, *Scope) *Scope

	// service lifecycle
	serviceExist(string) bool
	serviceGet(string) (any, bool)
	serviceGetRec(string) (any, *Scope, bool)
	serviceSet(string, any)
	serviceForEach(func(string, any))
	serviceHealthCheck(context.Context, string) error
	serviceShutdown(context.Context, string) error
	onServiceInvoke(string)
}

func getInjectorOrDefault(i Injector) Injector {
	if i != nil {
		return i
	}

	return DefaultRootScope
}
