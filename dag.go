package do

import (
	"sync"
)

// newEdgeService creates a new EdgeService with the provided scope ID, scope name, and service name.
func newEdgeService(scopeID string, scopeName string, serviceName string) EdgeService {
	return EdgeService{
		ScopeID:   scopeID,
		ScopeName: scopeName,
		Service:   serviceName,
	}
}

// EdgeService represents a service in the DAG, identified by scope ID, scope name, and service name.
type EdgeService struct {
	ScopeID   string
	ScopeName string
	Service   string
}

// newDAG creates a new DAG (Directed Acyclic Graph) with initialized dependencies and dependents maps.
func newDAG() *DAG {
	return &DAG{
		mu:           sync.RWMutex{},
		dependencies: map[EdgeService]map[EdgeService]struct{}{},
		dependents:   map[EdgeService]map[EdgeService]struct{}{},
	}
}

// DAG represents a Directed Acyclic Graph of services, tracking dependencies and dependents.
type DAG struct {
	mu           sync.RWMutex
	dependencies map[EdgeService]map[EdgeService]struct{}
	dependents   map[EdgeService]map[EdgeService]struct{}
}

// addDependency adds a dependency relationship from one service to another in the DAG.
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
func (d *DAG) removeService(scopeID, scopeName, serviceName string) {
	edge := newEdgeService(scopeID, scopeName, serviceName)

	d.mu.Lock()
	defer d.mu.Unlock()

	dependencies, dependents := d.explainServiceImplem(edge)

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
func (d *DAG) explainService(scopeID, scopeName, serviceName string) (dependencies, dependents []EdgeService) {
	edge := newEdgeService(scopeID, scopeName, serviceName)

	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.explainServiceImplem(edge)
}

func (d *DAG) explainServiceImplem(edge EdgeService) (dependencies, dependents []EdgeService) {
	dependencies, dependents = []EdgeService{}, []EdgeService{}

	if kv, ok := d.dependencies[edge]; ok {
		dependencies = keys(kv)
	}

	if kv, ok := d.dependents[edge]; ok {
		dependents = keys(kv)
	}

	return dependencies, dependents
}
