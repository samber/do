package do

import "context"

// Injector is a DI container.
type Injector interface {
	// api
	ID() string
	Name() string
	Scope(string, ...func(Injector)) *Scope
	RootScope() *RootScope
	Ancestors() []*Scope
	Children() []*Scope
	ChildByID(string) (*Scope, bool)
	ChildByName(string) (*Scope, bool)
	ListProvidedServices() []EdgeService
	ListInvokedServices() []EdgeService
	HealthCheck() map[string]error
	HealthCheckWithContext(context.Context) map[string]error
	Shutdown() *ShutdownErrors
	ShutdownWithContext(context.Context) *ShutdownErrors
	clone(*RootScope, *Scope) *Scope

	// service lifecycle
	serviceExist(string) bool
	serviceExistRec(string) bool
	serviceGet(string) (any, bool)
	serviceGetRec(string) (any, *Scope, bool)
	serviceSet(string, any) // serviceSet is not protected against double registration
	serviceForEach(func(string, *Scope, any) bool)
	serviceForEachRec(func(string, *Scope, any) bool)
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
