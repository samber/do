package do

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewEdgeService checks the creation of a new EdgeService.
func TestNewEdgeService(t *testing.T) {
	is := assert.New(t)

	expected := EdgeService{"foo", "bar", "baz"}
	actual := newEdgeService("foo", "bar", "baz")

	is.Equal(expected, actual)
}

// TestNewDAG checks the initialization of a new DAG.
func TestNewDAG(t *testing.T) {
	is := assert.New(t)

	dag := newDAG()
	expectedDependencies := unSyncMap(new(sync.Map))
	expectedDependents := unSyncMap(new(sync.Map))

	is.Equal(expectedDependencies, unSyncMap(dag.dependencies))
	is.Equal(expectedDependents, unSyncMap(dag.dependents))
}

// TestDAG_addDependency checks the addition of dependencies to the DAG.
func TestDAG_addDependency(t *testing.T) {
	is := assert.New(t)

	edge1 := newEdgeService("scope1", "scope1", "service1")
	edge2 := newEdgeService("scope2", "scope2", "service2")
	edge3 := newEdgeService("scope3", "scope3", "service3")

	dag := newDAG()

	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")

	expectedDependencies := map[interface{}]interface{}{edge1: map[interface{}]interface{}{edge2: struct{}{}}}
	expectedDependents := map[interface{}]interface{}{edge2: map[interface{}]interface{}{edge1: struct{}{}}}

	is.Equal(expectedDependencies, unSyncMap(dag.dependencies))
	is.Equal(expectedDependents, unSyncMap(dag.dependents))

	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")

	expectedDependencies[edge3] = map[interface{}]interface{}{edge2: struct{}{}}
	expectedDependents[edge2] = map[interface{}]interface{}{edge1: struct{}{}, edge3: struct{}{}}

	is.Equal(expectedDependencies, unSyncMap(dag.dependencies))
	is.Equal(expectedDependents, unSyncMap(dag.dependents))
}

// TestDAG_explainService checks the explanation of dependencies for a service in the DAG.
func TestDAG_explainService(t *testing.T) {
	is := assert.New(t)

	edge1 := newEdgeService("scope1", "scope1", "service1")
	edge2 := newEdgeService("scope2", "scope2", "service2")
	edge3 := newEdgeService("scope3", "scope3", "service3")

	dag := newDAG()
	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")
	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")

	// edge1
	a, b := dag.explainService("scope1", "scope1", "service1")
	is.ElementsMatch([]EdgeService{edge2}, a)
	is.ElementsMatch([]EdgeService{}, b)

	// edge2
	a, b = dag.explainService("scope2", "scope2", "service2")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{edge1, edge3}, b)

	// edge3
	a, b = dag.explainService("scope3", "scope3", "service3")
	is.ElementsMatch([]EdgeService{edge2}, a)
	is.ElementsMatch([]EdgeService{}, b)

	// not found
	a, b = dag.explainService("scopeX", "scopeX", "serviceX")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
}

func unSyncMap(syncMap *sync.Map) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})

	syncMap.Range(func(key, value interface{}) bool {
		if vSyncMap, ok := value.(*sync.Map); ok {
			result[key] = unSyncMap(vSyncMap)
		} else {
			result[key] = value
		}

		return true
	})

	return result
}
