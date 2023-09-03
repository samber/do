package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEdgeService(t *testing.T) {
	is := assert.New(t)

	is.Equal(EdgeService{"foo", "bar", "baz"}, newEdgeService("foo", "bar", "baz"))
}

func TestNewDAG(t *testing.T) {
	is := assert.New(t)

	dag := newDAG()
	is.Equal(map[EdgeService]map[EdgeService]struct{}{}, dag.dependencies)
	is.Equal(map[EdgeService]map[EdgeService]struct{}{}, dag.dependents)
}

func TestDAG_addDependency(t *testing.T) {
	is := assert.New(t)

	dag := newDAG()
	is.Equal(map[EdgeService]map[EdgeService]struct{}{}, dag.dependencies)
	is.Equal(map[EdgeService]map[EdgeService]struct{}{}, dag.dependents)

	edge1 := newEdgeService("scope1", "scope1", "service1")
	edge2 := newEdgeService("scope2", "scope2", "service2")
	edge3 := newEdgeService("scope3", "scope3", "service3")

	dag.addDependency("scope1", "scope1", "service1", "scope2", "scope2", "service2")
	is.Equal(map[EdgeService]map[EdgeService]struct{}{edge1: {edge2: {}}}, dag.dependencies)
	is.Equal(map[EdgeService]map[EdgeService]struct{}{edge2: {edge1: {}}}, dag.dependents)

	dag.addDependency("scope3", "scope3", "service3", "scope2", "scope2", "service2")
	is.Equal(map[EdgeService]map[EdgeService]struct{}{edge1: {edge2: {}}, edge3: {edge2: {}}}, dag.dependencies)
	is.Equal(map[EdgeService]map[EdgeService]struct{}{edge2: {edge1: {}, edge3: {}}}, dag.dependents)
}

func TestDAG_explainService(t *testing.T) {
	is := assert.New(t)

	edge1 := newEdgeService("scope1", "scope1", "service1")
	edge2 := newEdgeService("scope2", "scope2", "service2")
	edge3 := newEdgeService("scope3", "scope3", "service3")

	dag := newDAG()
	dag.dependencies = map[EdgeService]map[EdgeService]struct{}{edge1: {edge2: {}}, edge3: {edge2: {}}}
	dag.dependents = map[EdgeService]map[EdgeService]struct{}{edge2: {edge1: {}, edge3: {}}}

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
