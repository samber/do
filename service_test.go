package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferServiceName(t *testing.T) {
	t.Parallel()
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

func TestServiceIsAssignable(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	svc1 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.True(serviceIsAssignable[*lazyTestHeathcheckerOK](svc1))
	is.True(serviceIsAssignable[Healthchecker](svc1))
	is.False(serviceIsAssignable[Shutdowner](svc1))
	is.False(serviceIsAssignable[*lazyTestHeathcheckerKO](svc1))
}
