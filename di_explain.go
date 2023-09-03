package do

func ExplainService[T any](scope Injector) (dependencies []EdgeService, dependents []EdgeService, ok bool) {
	name := inferServiceName[T]()

	return ExplainNamedService(scope, name)
}

func ExplainNamedService(scope Injector, name string) (dependencies []EdgeService, dependents []EdgeService, ok bool) {
	_i := getInjectorOrDefault(scope)

	_, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return nil, nil, false
	}

	dependencies, dependents = _i.RootScope().dag.explainService(serviceScope.ID(), serviceScope.Name(), name)
	return dependencies, dependents, true
}
