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
	is.Equal("github.com/samber/do/v2.Healthchecker", inferServiceName[Healthchecker]())
}

func TestInferServiceProviderStacktrace(t *testing.T) {
	// @TODO
}

func TestInferServiceInfo(t *testing.T) {
	// @TODO
}

func TestServiceCanCastTo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	svc1 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.True(serviceCanCastTo[*lazyTestHeathcheckerOK](svc1))
	is.True(serviceCanCastTo[Healthchecker](svc1))
	is.False(serviceCanCastTo[Shutdowner](svc1))
	is.False(serviceCanCastTo[*lazyTestHeathcheckerKO](svc1))

	svc2 := newServiceLazy("foobar", func(i Injector) (iTestHeathchecker, error) {
		return &lazyTestHeathcheckerOK{}, nil
	})
	is.True(serviceCanCastTo[Healthchecker](svc2))
}
