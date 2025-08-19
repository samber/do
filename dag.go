package do

import (
	"sync"
)

// newEdgeService creates a new EdgeService with the provided scope ID, scope name, and service name.
// This function is used internally to create consistent EdgeService instances.
//
// Parameters:
//   - scopeID: The unique identifier of the scope
//   - scopeName: The human-readable name of the scope
//   - serviceName: The name of the service
//
// Returns a new EdgeService instance.
func newEdgeService(scopeID string, scopeName string, serviceName string) EdgeService {
	return EdgeService{
		ScopeID:   scopeID,
		ScopeName: scopeName,
		Service:   serviceName,
	}
}

// EdgeService represents a service in the dependency graph (DAG), identified by scope ID, scope name, and service name.
// This type is used to uniquely identify services across the entire scope hierarchy for dependency tracking.
//
// Fields:
//   - ScopeID: The unique identifier of the scope containing the service
//   - ScopeName: The human-readable name of the scope containing the service
//   - Service: The name of the service within the scope
type EdgeService struct {
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
		dependencies: map[EdgeService]map[EdgeService]struct{}{},
		dependents:   map[EdgeService]map[EdgeService]struct{}{},
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
	dependencies map[EdgeService]map[EdgeService]struct{}
	dependents   map[EdgeService]map[EdgeService]struct{}
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
	from := newEdgeService(fromScopeID, fromScopeName, fromServiceName)
	to := newEdgeService(toScopeID, toScopeName, toServiceName)

	d.mu.Lock()
	defer d.mu.Unlock()

	// from -> to
	if _, ok := d.dependencies[from]; !ok {
		d.dependencies[from] = map[EdgeService]struct{}{}
	}
	d.dependencies[from][to] = struct{}{}

	// from <- to
	if _, ok := d.dependents[to]; !ok {
		d.dependents[to] = map[EdgeService]struct{}{}
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
	edge := newEdgeService(scopeID, scopeName, serviceName)

	d.mu.Lock()
	defer d.mu.Unlock()

	dependencies, dependents := d.explainServiceUnsafe(edge)

	for _, dependency := range dependencies {
		delete(d.dependents[dependency], edge)
	}

	// should be empty, because we remove dependencies in the inverse invocation order
	for _, dependent := range dependents {
		delete(d.dependencies[dependent], edge)
	}

	delete(d.dependencies, edge)
	delete(d.dependents, edge)
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
func (d *DAG) explainService(scopeID, scopeName, serviceName string) (dependencies, dependents []EdgeService) {
	edge := newEdgeService(scopeID, scopeName, serviceName)

	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.explainServiceUnsafe(edge)
}

// explainServiceUnsafe is the internal implementation of explainService.
// This function performs the actual work of retrieving dependency information
// without acquiring locks (assumes the caller has already acquired appropriate locks).
//
// Parameters:
//   - edge: The EdgeService to explain
//
// Returns two slices:
//   - dependencies: Services that the specified service depends on
//   - dependents: Services that depend on the specified service
func (d *DAG) explainServiceUnsafe(edge EdgeService) (dependencies, dependents []EdgeService) {
	dependencies, dependents = []EdgeService{}, []EdgeService{}

	if kv, ok := d.dependencies[edge]; ok {
		dependencies = keys(kv)
	}

	if kv, ok := d.dependents[edge]; ok {
		dependents = keys(kv)
	}

	return dependencies, dependents
}
