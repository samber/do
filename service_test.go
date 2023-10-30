package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferServiceName(t *testing.T) {
	is := assert.New(t)

	// more tests in the package
	is.Equal("int", inferServiceName[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", inferServiceName[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", inferServiceName[*eagerTest]())
	is.Equal("*github.com/samber/do/v2.Healthchecker", inferServiceName[Healthchecker]())
}

func TestInferServiceProviderStacktrace(t *testing.T) {
	// @TODO
}

func TestInferServiceInfo(t *testing.T) {
	// @TODO
}
