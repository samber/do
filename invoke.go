package do

import (
	"fmt"
	"strings"
)

func invoke[T any](i Injector, name string) (T, error) {
	_i := getInjectorOrDefault(i)

	invokerName := ""
	invokerChain := []string{}
	if vs, ok := _i.(*virtualScope); ok {
		if len(vs.invokerChain) > 0 {
			invokerName = vs.invokerChain[len(vs.invokerChain)-1]
		}

		invokerChain = vs.invokerChain

		if contains(invokerChain, name) {
			return empty[T](), fmt.Errorf("DI: circular dependency detected: %s -> %s", strings.Join(invokerChain, " -> "), name)
		}
	}

	serviceAny, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return empty[T](), serviceNotFound(_i, name)
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return empty[T](), serviceNotFound(_i, name)
	}

	instance, err := service.getInstance(&virtualScope{invokerChain: append(invokerChain, name), self: serviceScope})
	if err != nil {
		return empty[T](), err
	}

	if _, ok := _i.(*virtualScope); ok {
		_i.RootScope().dag.addDependency(_i.ID(), _i.Name(), invokerName, serviceScope.ID(), serviceScope.Name(), name)
	}

	serviceScope.onServiceInvoke(name)
	_i.RootScope().opts.Logf("DI: service %s invoked", name)

	return instance, nil
}

func serviceNotFound(injector Injector, name string) error {
	// @TODO: use the Keys+Map functions from `golang.org/x/exp/maps` as
	// soon as it is released in stdlib.
	services := injector.ListProvidedServices()
	servicesNames := mAp(services, func(edge EdgeService, _ int) string {
		return fmt.Sprintf("`%s`", edge.Service)
	})

	return fmt.Errorf("DI: could not find service `%s`, available services: %s", name, strings.Join(servicesNames, ", "))
}
