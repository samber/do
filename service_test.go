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

	is.Equal(ServiceTypeLazy, inferServiceType(svc1))
	is.Equal(ServiceTypeEager, inferServiceType(svc2))
	is.Equal(ServiceTypeTransient, inferServiceType(svc3))
	is.Panics(func() {
		is.Equal(ServiceTypeTransient, inferServiceType[string](any(svc3).(Service[string])))
	})
}

func TestInferServiceStacktrace(t *testing.T) {
	// @TODO
}

func TestInferServiceInfo(t *testing.T) {
	// @TODO
}
