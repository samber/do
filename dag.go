package do

import (
	"sync"
)

// newServiceDescription creates a new ServiceDescription with the provided scope ID, scope name, and service name.
// This function is used internally to create consistent ServiceDescription instances.
//
// Parameters:
//   - scopeID: The unique identifier of the scope
//   - scopeName: The human-readable name of the scope
//   - serviceName: The name of the service
//
// Returns a new ServiceDescription instance.
func newServiceDescription(scopeID string, scopeName string, serviceName string) ServiceDescription {
	return ServiceDescription{
		ScopeID:   scopeID,
		ScopeName: scopeName,
		Service:   serviceName,
	}
}

// ServiceDescription represents a service in the dependency graph (DAG), identified by scope ID, scope name, and service name.
// This type is used to uniquely identify services across the entire scope hierarchy for dependency tracking.
//
// Fields:
//   - ScopeID: The unique identifier of the scope containing the service
//   - ScopeName: The human-readable name of the scope containing the service
//   - Service: The name of the service within the scope
type ServiceDescription struct {
	ScopeID   string
	ScopeName string
	Service   string
}

// newDAG creates a new DAG (Directed Acyclic Graph) with initialized dependencies and dependents maps.
// This function initializes a new dependency graph for tracking service relationships.
//
// Returns a new DAG instance ready for dependency tracking.
func newDAG() *DAG {
	return &DAG{
		mu:           sync.RWMutex{},
		dependencies: map[ServiceDescription]map[ServiceDescription]struct{}{},
		dependents:   map[ServiceDescription]map[ServiceDescription]struct{}{},
	}
}

// DAG represents a Directed Acyclic Graph of services, tracking dependencies and dependents.
// This type manages the relationships between services to ensure proper initialization order
// and detect circular dependencies.
//
// The DAG maintains two maps:
//   - dependencies: Maps each service to the services it depends on
//   - dependents: Maps each service to the services that depend on it
//
// Fields:
//   - mu: Read-write mutex for thread-safe access to the graph
//   - dependencies: Map of services to their dependencies
//   - dependents: Map of services to their dependents
type DAG struct {
	mu           sync.RWMutex
	dependencies map[ServiceDescription]map[ServiceDescription]struct{}
	dependents   map[ServiceDescription]map[ServiceDescription]struct{}
}

// addDependency adds a dependency relationship from one service to another in the DAG.
// This function establishes that the 'from' service depends on the 'to' service,
// which affects the order of service initialization and shutdown.
//
// Parameters:
//   - fromScopeID: The scope ID of the dependent service
//   - fromScopeName: The scope name of the dependent service
//   - fromServiceName: The name of the dependent service
//   - toScopeID: The scope ID of the dependency service
//   - toScopeName: The scope name of the dependency service
//   - toServiceName: The name of the dependency service
//
// This function is thread-safe and updates both the dependencies and dependents maps.
func (d *DAG) addDependency(fromScopeID, fromScopeName, fromServiceName, toScopeID, toScopeName, toServiceName string) {
	from := newServiceDescription(fromScopeID, fromScopeName, fromServiceName)
	to := newServiceDescription(toScopeID, toScopeName, toServiceName)

	d.mu.Lock()
	defer d.mu.Unlock()

	// from -> to
	if _, ok := d.dependencies[from]; !ok {
		d.dependencies[from] = map[ServiceDescription]struct{}{}
	}
	d.dependencies[from][to] = struct{}{}

	// from <- to
	if _, ok := d.dependents[to]; !ok {
		d.dependents[to] = map[ServiceDescription]struct{}{}
	}
	d.dependents[to][from] = struct{}{}
}

// removeService removes a dependency relationship between services in the DAG.
// This function is called when a service is being removed from the container,
// and it cleans up all dependency relationships involving that service.
//
// Parameters:
//   - scopeID: The scope ID of the service to remove
//   - scopeName: The scope name of the service to remove
//   - serviceName: The name of the service to remove
//
// This function removes the service from both dependencies and dependents maps,
// ensuring the graph remains consistent.
func (d *DAG) removeService(scopeID, scopeName, serviceName string) {
	desc := newServiceDescription(scopeID, scopeName, serviceName)

	d.mu.Lock()
	defer d.mu.Unlock()

	dependencies, dependents := d.explainServiceUnsafe(desc)

	for _, dependency := range dependencies {
		delete(d.dependents[dependency], desc)
	}

	// should be empty, because we remove dependencies in the inverse invocation order
	for _, dependent := range dependents {
		delete(d.dependencies[dependent], desc)
	}

	delete(d.dependencies, desc)
	delete(d.dependents, desc)
}

// explainService provides information about a service's dependencies and dependents in the DAG.
// This function returns the list of services that the specified service depends on,
// as well as the list of services that depend on the specified service.
//
// Parameters:
//   - scopeID: The scope ID of the service to explain
//   - scopeName: The scope name of the service to explain
//   - serviceName: The name of the service to explain
//
// Returns two slices:
//   - dependencies: Services that the specified service depends on
//   - dependents: Services that depend on the specified service
//
// This function is thread-safe and provides read-only access to the dependency graph.
func (d *DAG) explainService(scopeID, scopeName, serviceName string) (dependencies, dependents []ServiceDescription) {
	desc := newServiceDescription(scopeID, scopeName, serviceName)

	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.explainServiceUnsafe(desc)
}

// explainServiceUnsafe is the internal implementation of explainService.
// This function performs the actual work of retrieving dependency information
// without acquiring locks (assumes the caller has already acquired appropriate locks).
//
// Parameters:
//   - desc: The ServiceDescription to explain
//
// Returns two slices:
//   - dependencies: Services that the specified service depends on
//   - dependents: Services that depend on the specified service
func (d *DAG) explainServiceUnsafe(desc ServiceDescription) (dependencies, dependents []ServiceDescription) {
	dependencies, dependents = []ServiceDescription{}, []ServiceDescription{}

	if kv, ok := d.dependencies[desc]; ok {
		dependencies = keys(kv)
	}

	if kv, ok := d.dependents[desc]; ok {
		dependents = keys(kv)
	}

	return dependencies, dependents
}
