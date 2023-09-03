package do

import (
	"sync"
)

func newEdgeService(scopeID string, scopeName string, serviceName string) EdgeService {
	return EdgeService{
		ScopeID:   scopeID,
		ScopeName: scopeName,
		Service:   serviceName,
	}
}

type EdgeService struct {
	ScopeID   string
	ScopeName string
	Service   string
}

func newDAG() *DAG {
	return &DAG{
		mu:           sync.RWMutex{},
		dependencies: map[EdgeService]map[EdgeService]struct{}{},
		dependents:   map[EdgeService]map[EdgeService]struct{}{},
	}
}

type DAG struct {
	mu           sync.RWMutex
	dependencies map[EdgeService]map[EdgeService]struct{}
	dependents   map[EdgeService]map[EdgeService]struct{}
}

func (d *DAG) addDependency(fromScopeID string, fromScopeName string, fromServiceName string, toScopeID string, toScopeName string, toServiceName string) {
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

func (d *DAG) explainService(scopeID string, scopeName string, serviceName string) (dependencies []EdgeService, dependents []EdgeService) {
	edge := newEdgeService(scopeID, scopeName, serviceName)

	dependencies = []EdgeService{}
	dependents = []EdgeService{}

	d.mu.RLock()
	defer d.mu.RUnlock()

	if kv, ok := d.dependencies[edge]; ok {
		dependencies = keys(kv)
	}

	if kv, ok := d.dependents[edge]; ok {
		dependents = keys(kv)
	}

	return dependencies, dependents
}
