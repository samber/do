package do

import (
	"fmt"
	"strings"
	"sync"
)

type Injector struct {
	mu       sync.RWMutex
	services map[string]any

	// It should be a graph instead of simple ordered list.
	orderedInvocation      map[string]int // map is faster than slice
	orderedInvocationIndex int
}

func New() *Injector {
	return &Injector{
		mu:       sync.RWMutex{},
		services: make(map[string]any),

		orderedInvocation:      map[string]int{},
		orderedInvocationIndex: 0,
	}
}

func (i *Injector) HealthCheck() map[string]error {
	i.mu.RLock()
	names := keys(i.services)
	i.mu.RUnlock()

	results := map[string]error{}

	for _, name := range names {
		results[name] = i.shutdownImplem(name)
	}

	return results
}

func (i *Injector) Shutdown() error {
	i.mu.RLock()
	invocations := invertMap(i.orderedInvocation)
	i.mu.RUnlock()

	for index := i.orderedInvocationIndex; index >= 0; index-- {
		name, ok := invocations[index]
		if !ok {
			continue
		}

		err := i.shutdownImplem(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Injector) healthcheckImplem(name string) error {
	i.mu.Lock()

	serviceAny, ok := i.services[name]
	if !ok {
		i.mu.Unlock()
		return fmt.Errorf("DI: could not find service `%s`", name)
	}

	i.mu.Unlock()

	service, ok := serviceAny.(healthcheckableService)
	if ok {
		err := service.healthcheck()
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Injector) shutdownImplem(name string) error {
	i.mu.Lock()

	serviceAny, ok := i.services[name]
	if !ok {
		i.mu.Unlock()
		return fmt.Errorf("DI: could not find service `%s`", name)
	}

	i.mu.Unlock()

	service, ok := serviceAny.(shutdownableService)
	if ok {
		err := service.shutdown()
		if err != nil {
			return err
		}
	}

	delete(i.services, name)
	delete(i.orderedInvocation, name)

	return nil
}

func (i *Injector) get(name string) (any, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	s, ok := i.services[name]
	return s, ok
}

func (i *Injector) set(name string, service any) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.services[name] = service
}

func (i *Injector) remove(name string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.services, name)
}

func (i *Injector) forEach(cb func(s any)) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, s := range i.services {
		cb(s)
	}
}

func (i *Injector) serviceNotFound(name string) error {
	// @TODO: use the Keys+Map functions from `golang.org/x/exp/maps` as
	// soon as it is released in stdlib.
	servicesNames := keys(i.services)
	servicesNames = mAp(servicesNames, func(name string) string {
		return fmt.Sprintf("`%s`", name)
	})

	return fmt.Errorf("DI: could not find service `%s`, available services: %s", name, strings.Join(servicesNames, ", "))
}

func (i *Injector) onServiceInvoke(name string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.orderedInvocation[name]; !ok {
		i.orderedInvocation[name] = i.orderedInvocationIndex
		i.orderedInvocationIndex++
	}
}
