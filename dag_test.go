package do

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewServiceDescription checks the creation of a new ServiceDescription.
func TestNewServiceDescription(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	expected := ServiceDescription{"foo", "bar", "baz"}
	actual := newServiceDescription("foo", "bar", "baz")

	is.Equal(expected, actual)
}

// TestNewDAG checks the initialization of a new DAG.
func TestNewDAG(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	dag := newDAG()
	expectedDependencies := map[ServiceDescription]map[ServiceDescription]struct{}{}
	expectedDependents := map[ServiceDescription]map[ServiceDescription]struct{}{}

	is.Equal(expectedDependencies, dag.dependencies)
	is.Equal(expectedDependents, dag.dependents)
}

// TestDAG_addDependency checks the addition of dependencies to the DAG.
func TestDAG_addDependency(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	edge1 := newServiceDescription("scope1", "scope1", "service1")
	edge2 := newServiceDescription("scope2", "scope2", "service2")
	edge3 := newServiceDescription("scope3", "scope3", "service3")

	dag := newDAG()

	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")

	expectedDependencies := map[ServiceDescription]map[ServiceDescription]struct{}{edge1: {edge2: {}}}
	expectedDependents := map[ServiceDescription]map[ServiceDescription]struct{}{edge2: {edge1: {}}}

	is.Equal(expectedDependencies, dag.dependencies)
	is.Equal(expectedDependents, dag.dependents)

	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")

	expectedDependencies = map[ServiceDescription]map[ServiceDescription]struct{}{edge1: {edge2: {}}, edge3: {edge2: {}}}
	expectedDependents = map[ServiceDescription]map[ServiceDescription]struct{}{edge2: {edge1: {}, edge3: {}}}

	is.Equal(expectedDependencies, dag.dependencies)
	is.Equal(expectedDependents, dag.dependents)
}

// TestDAG_removeService checks the removal of dependencies to the DAG.
func TestDAG_removeService(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	edge1 := newServiceDescription("scope1", "scope1", "service1")
	// edge2 := newServiceDescription("scope2", "scope2", "service2")
	edge3 := newServiceDescription("scope3", "scope3", "service3")

	dag := newDAG()

	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")
	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")

	dag.removeService("scope2", "scope2", "service2")

	expectedDependencies := map[ServiceDescription]map[ServiceDescription]struct{}{edge1: {}, edge3: {}}
	expectedDependents := map[ServiceDescription]map[ServiceDescription]struct{}{}

	is.Equal(expectedDependencies, dag.dependencies)
	is.Equal(expectedDependents, dag.dependents)
}

// TestDAG_explainService checks the explanation of dependencies for a service in the DAG.
func TestDAG_explainService(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	edge1 := newServiceDescription("scope1", "scope1", "service1")
	edge2 := newServiceDescription("scope2", "scope2", "service2")
	edge3 := newServiceDescription("scope3", "scope3", "service3")

	dag := newDAG()
	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")
	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")

	// edge1
	a, b := dag.explainService("scope1", "scope1", "service1")
	is.ElementsMatch([]ServiceDescription{edge2}, a)
	is.ElementsMatch([]ServiceDescription{}, b)

	// edge2
	a, b = dag.explainService("scope2", "scope2", "service2")
	is.ElementsMatch([]ServiceDescription{}, a)
	is.ElementsMatch([]ServiceDescription{edge1, edge3}, b)

	// edge3
	a, b = dag.explainService("scope3", "scope3", "service3")
	is.ElementsMatch([]ServiceDescription{edge2}, a)
	is.ElementsMatch([]ServiceDescription{}, b)

	// not found
	a, b = dag.explainService("scopeX", "scopeX", "serviceX")
	is.ElementsMatch([]ServiceDescription{}, a)
	is.ElementsMatch([]ServiceDescription{}, b)
}
