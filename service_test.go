package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferServiceType(t *testing.T) {
	is := assert.New(t)

	svc1 := newServiceLazy[int]("foobar1", func(i Injector) (int, error) { return 42, nil })
	svc2 := newServiceEager[int]("foobar2", 42)
	svc3 := newServiceTransient[int]("foobar3", func(i Injector) (int, error) { return 42, nil })
	svc4 := newServiceAlias[int, int]("foobar4", New(), "foobar5")

	is.Equal(ServiceTypeLazy, inferServiceType[int](svc1))
	is.Equal(ServiceTypeEager, inferServiceType[int](svc2))
	is.Equal(ServiceTypeTransient, inferServiceType[int](svc3))
	is.Panics(func() {
		is.Equal(ServiceTypeTransient, inferServiceType[string](any(svc3).(Service[string])))
	})
	is.Equal(ServiceTypeAlias, inferServiceType[int](svc4))
}

func TestInferServiceName(t *testing.T) {
	is := assert.New(t)

	// more tests in the package
	is.Equal("int", inferServiceName[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", inferServiceName[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", inferServiceName[*eagerTest]())
	is.Equal("*github.com/samber/do/v2.Healthchecker", inferServiceName[Healthchecker]())
}

func TestInferServiceStacktrace(t *testing.T) {
	// @TODO
}

func TestInferServiceInfo(t *testing.T) {
	// @TODO
}
