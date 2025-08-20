package do

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInferServiceName(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	t.Parallel()
	is := assert.New(t)

	// more tests in the package
	is.Equal("int", inferServiceName[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", inferServiceName[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", inferServiceName[*eagerTest]())
	is.Equal("github.com/samber/do/v2.Healthchecker", inferServiceName[Healthchecker]())
}

func TestInferServiceProviderStacktrace(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	// @TODO
}

func TestInferServiceInfo(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	// @TODO
}

func TestDoServiceCanCastToGeneric(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	t.Parallel()
	is := assert.New(t)

	svc1 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.True(serviceCanCastToGeneric[*lazyTestHeathcheckerOK](svc1))
	is.True(serviceCanCastToGeneric[Healthchecker](svc1))
	is.False(serviceCanCastToGeneric[Shutdowner](svc1))
	is.False(serviceCanCastToGeneric[*lazyTestHeathcheckerKO](svc1))

	svc2 := newServiceLazy("foobar", func(i Injector) (iTestHeathchecker, error) {
		return &lazyTestHeathcheckerOK{}, nil
	})
	is.True(serviceCanCastToGeneric[Healthchecker](svc2))
}

func TestDoServiceCanCastToType(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	t.Parallel()
	is := assert.New(t)

	svc1 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.True(serviceCanCastToType(svc1, reflect.TypeOf((**lazyTestHeathcheckerOK)(nil)).Elem()))
	is.True(serviceCanCastToType(svc1, reflect.TypeOf((*Healthchecker)(nil)).Elem()))
	is.False(serviceCanCastToType(svc1, reflect.TypeOf((*Shutdowner)(nil)).Elem()))
	is.False(serviceCanCastToType(svc1, reflect.TypeOf((**lazyTestHeathcheckerKO)(nil)).Elem()))

	svc2 := newServiceLazy("foobar", func(i Injector) (iTestHeathchecker, error) {
		return &lazyTestHeathcheckerOK{}, nil
	})
	is.True(serviceCanCastToType(svc2, reflect.TypeOf((*Healthchecker)(nil)).Elem()))
}
