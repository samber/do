package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServiceName(t *testing.T) {
	is := assert.New(t)

	type testStruct struct{} //nolint:unused

	type testInterface interface{} //nolint:unused

	name := generateServiceName[testStruct]()
	is.Equal("github.com/samber/do.testStruct", name)
	name = generateServiceName[[]testStruct]()
	is.Equal("github.com/samber/do.[]testStruct", name)

	name = generateServiceName[*testStruct]()
	is.Equal("github.com/samber/do.*testStruct", name)
	name = generateServiceName[*[]testStruct]()
	is.Equal("github.com/samber/do.*[]testStruct", name)
	name = generateServiceName[[]*testStruct]()
	is.Equal("github.com/samber/do.[]*testStruct", name)
	name = generateServiceName[*[]*testStruct]()
	is.Equal("github.com/samber/do.*[]*testStruct", name)
	name = generateServiceName[*[]*[]**testStruct]()
	is.Equal("github.com/samber/do.*[]*[]**testStruct", name)

	name = generateServiceName[***testStruct]()
	is.Equal("github.com/samber/do.***testStruct", name)
	name = generateServiceName[testInterface]()
	is.Equal("github.com/samber/do.*testInterface", name)
	name = generateServiceName[*testInterface]()
	is.Equal("github.com/samber/do.*testInterface", name)
	name = generateServiceName[***testInterface]()
	is.Equal("github.com/samber/do.***testInterface", name)

	name = generateServiceName[int]()
	is.Equal("int", name)
	name = generateServiceName[[]int]()
	is.Equal("[]int", name)

	name = generateServiceName[*int]()
	is.Equal("*int", name)
	name = generateServiceName[*[]int]()
	is.Equal("*[]int", name)
	name = generateServiceName[[]*int]()
	is.Equal("[]*int", name)
	name = generateServiceName[*[]*int]()
	is.Equal("*[]*int", name)
	name = generateServiceName[*[]*[]**int]()
	is.Equal("*[]*[]**int", name)

	// name = generateServiceName[any]()
	// is.Equal("any", name)

	// name = generateServiceName[*any]()
	// is.Equal("any", name)
}
