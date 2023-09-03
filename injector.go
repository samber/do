package do

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
	Shutdown() error
	clone(*RootScope, *Scope) *Scope

	// service lifecycle
	serviceExist(string) bool
	serviceGet(string) (any, bool)
	serviceGetRec(string) (any, *Scope, bool)
	serviceSet(string, any)
	serviceForEach(func(string, any))
	serviceHealthCheck(string) error
	serviceShutdown(string) error
	onServiceInvoke(string)
}

func getInjectorOrDefault(i Injector) Injector {
	if i != nil {
		return i
	}

	return DefaultRootScope
}
