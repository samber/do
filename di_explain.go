package do

// ExplainService returns a list of dependencies and dependents of a service.
func ExplainService[T any](scope Injector) (dependencies []EdgeService, dependents []EdgeService, ok bool) {
	name := inferServiceName[T]()

	return ExplainNamedService(scope, name)
}

// ExplainService returns a list of dependencies and dependents of a named service.
func ExplainNamedService(scope Injector, name string) (dependencies []EdgeService, dependents []EdgeService, ok bool) {
	_i := getInjectorOrDefault(scope)

	_, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return nil, nil, false
	}

	dependencies, dependents = _i.RootScope().dag.explainService(serviceScope.ID(), serviceScope.Name(), name)
	return dependencies, dependents, true
}
