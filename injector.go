package do

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var DefaultInjector = New()

func getInjectorOrDefault(i *Injector) *Injector {
	if i != nil {
		return i
	}

	return DefaultInjector
}

func New() *Injector {
	return NewWithOpts(&InjectorOpts{})
}

type InjectorOpts struct {
	HookAfterRegistration func(injector *Injector, serviceName string)
	HookAfterShutdown     func(injector *Injector, serviceName string)

	Logf func(format string, args ...any)
}

func NewWithOpts(opts *InjectorOpts) *Injector {
	logf := opts.Logf
	if logf == nil {
		logf = func(format string, args ...any) {}
	}

	logf("injector created")

	return &Injector{
		mu:       sync.RWMutex{},
		services: make(map[string]any),

		orderedInvocation:      map[string]int{},
		orderedInvocationIndex: 0,

		hookAfterRegistration: opts.HookAfterRegistration,
		hookAfterShutdown:     opts.HookAfterShutdown,

		logf: logf,
	}
}

type Injector struct {
	mu       sync.RWMutex
	services map[string]any

	// It should be a graph instead of simple ordered list.
	orderedInvocation      map[string]int // map is faster than slice
	orderedInvocationIndex int

	hookAfterRegistration func(injector *Injector, serviceName string)
	hookAfterShutdown     func(injector *Injector, serviceName string)

	logf func(format string, args ...any)
}

func (i *Injector) ListProvidedServices() []string {
	names := []string{}

	i.mu.RLock()
	defer i.mu.RUnlock()

	for name := range i.services {
		names = append(names, name)
	}

	i.logf("exported list of services: %v", names)

	return names
}

func (i *Injector) ListInvokedServices() []string {
	names := []string{}

	i.mu.RLock()
	defer i.mu.RUnlock()

	for name := range i.orderedInvocation {
		names = append(names, name)
	}

	i.logf("exported list of invoked services: %v", names)

	return names
}

func (i *Injector) HealthCheck() map[string]error {
	i.mu.RLock()
	names := keys(i.services)
	i.mu.RUnlock()

	i.logf("requested healthcheck")

	results := map[string]error{}

	for _, name := range names {
		results[name] = i.healthcheckImplem(name)
	}

	i.logf("got healthcheck results: %v", results)

	return results
}

func (i *Injector) Shutdown() error {
	i.mu.RLock()
	invocations := invertMap(i.orderedInvocation)
	i.mu.RUnlock()

	i.logf("requested shutdown")

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

	i.logf("shutdowned services")

	return nil
}

// ShutdownOnSIGTERM listens for sigterm signal in order to graceful stop service.
// It will block until receiving a sigterm signal.
func (i *Injector) ShutdownOnSIGTERM() error {
	return i.ShutdownOnSignals(syscall.SIGTERM)
}

// ShutdownOnSignals listens for signals defined in signals parameter in order to graceful stop service.
// It will block until receiving any of these signal.
// If no signal is provided in signals parameter, syscall.SIGTERM will be added as default signal.
func (i *Injector) ShutdownOnSignals(signals ...os.Signal) error {
	// Make sure there is at least syscall.SIGTERM as a signal
	if len(signals) < 1 {
		signals = append(signals, syscall.SIGTERM)
	}
	ch := make(chan os.Signal, 1)

	signal.Notify(ch, signals...)

	<-ch
	signal.Stop(ch)
	close(ch)

	return i.Shutdown()
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
		i.logf("requested healthcheck for service %s", name)

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
		i.logf("requested shutdown for service %s", name)

		err := service.shutdown()
		if err != nil {
			return err
		}
	}

	delete(i.services, name)
	delete(i.orderedInvocation, name)

	i.onServiceShutdown(name)

	return nil
}

func (i *Injector) exists(name string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	_, ok := i.services[name]
	return ok
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

	// defering hook call will unlock mutex
	defer i.onServiceRegistration(name)
}

func (i *Injector) remove(name string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.services, name)
}

func (i *Injector) forEach(cb func(name string, service any)) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for name, service := range i.services {
		cb(name, service)
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

func (i *Injector) onServiceRegistration(name string) {
	if i.hookAfterRegistration != nil {
		i.hookAfterRegistration(i, name)
	}
}

func (i *Injector) onServiceShutdown(name string) {
	if i.hookAfterShutdown != nil {
		i.hookAfterShutdown(i, name)
	}
}

// Clone clones injector with provided services but not with invoked instances.
func (i *Injector) Clone() *Injector {
	return i.CloneWithOpts(&InjectorOpts{})
}

// CloneWithOpts clones injector with provided services but not with invoked instances, with options.
func (i *Injector) CloneWithOpts(opts *InjectorOpts) *Injector {
	clone := NewWithOpts(opts)

	i.mu.RLock()
	defer i.mu.RUnlock()

	for name, serviceAny := range i.services {
		if service, ok := serviceAny.(cloneableService); ok {
			clone.services[name] = service.clone()
		} else {
			clone.services[name] = service
		}
		defer clone.onServiceRegistration(name)
	}

	i.logf("injector cloned")

	return clone
}
