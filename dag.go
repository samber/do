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
		dependencies: new(sync.Map),
		dependents:   new(sync.Map),
	}
}

// DAG represents a Directed Acyclic Graph of services, tracking dependencies and dependents.
type DAG struct {
	dependencies *sync.Map
	dependents   *sync.Map
}

// addDependency adds a dependency relationship from one service to another in the DAG.
func (d *DAG) addDependency(fromScopeID, fromScopeName, fromServiceName, toScopeID, toScopeName, toServiceName string) {
	from := newEdgeService(fromScopeID, fromScopeName, fromServiceName)
	to := newEdgeService(toScopeID, toScopeName, toServiceName)

	d.addToMap(d.dependencies, from, to)
	d.addToMap(d.dependents, to, from)
}

// addToMap is a helper function to add a key-value pair to a sync.Map, creating a new sync.Map for the value if necessary.
func (d *DAG) addToMap(dependencyMap *sync.Map, key, value interface{}) {
	valueMap := new(sync.Map)
	valueMap.Store(value, struct{}{})

	if actual, loaded := dependencyMap.LoadOrStore(key, valueMap); loaded {
		actual.(*sync.Map).Store(value, struct{}{})
	}
}

// explainService provides information about a service's dependencies and dependents in the DAG.
func (d *DAG) explainService(scopeID, scopeName, serviceName string) (dependencies, dependents []EdgeService) {
	edge := newEdgeService(scopeID, scopeName, serviceName)

	dependencies = d.getServicesFromMap(d.dependencies, edge)
	dependents = d.getServicesFromMap(d.dependents, edge)

	return dependencies, dependents
}

// getServicesFromMap is a helper function to retrieve services related to a specific key from a sync.Map.
func (d *DAG) getServicesFromMap(serviceMap *sync.Map, edge EdgeService) []EdgeService {
	var services []EdgeService

	if kv, ok := serviceMap.Load(edge); ok {
		kv.(*sync.Map).Range(func(key, value interface{}) bool {
			edgeService, ok := key.(EdgeService)
			if ok {
				services = append(services, edgeService)
			}
			return ok
		})
	}

	return services
}
