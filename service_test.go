package do

import (
	"reflect"
	"testing"
	"time"

	"github.com/samber/do/v2/stacktrace"
	"github.com/stretchr/testify/assert"
)

func TestInferServiceName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// more tests in the package
	is.Equal("int", inferServiceName[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", inferServiceName[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", inferServiceName[*eagerTest]())
	is.Equal("github.com/samber/do/v2.Healthchecker", inferServiceName[Healthchecker]())
}

func TestInferServiceProviderStacktrace(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with lazy service
	lazyService := newServiceLazy("lazy-service", func(i Injector) (int, error) {
		return 42, nil
	})
	frame, ok := inferServiceProviderStacktrace(lazyService)
	is.True(ok)
	is.NotNil(frame)
	is.Contains(frame.File, "service_test.go")
	is.Contains(frame.Function, "TestInferServiceProviderStacktrace")

	// Test with eager service
	eagerService := newServiceEager("eager-service", func(i Injector) (int, error) {
		return 42, nil
	})
	frame2, ok2 := inferServiceProviderStacktrace(eagerService)
	is.True(ok2)
	is.NotNil(frame2)
	is.Contains(frame2.File, "service_eager.go")
	is.Contains(frame2.Function, "newServiceEager")

	// Test with transient service (should return false)
	transientService := newServiceTransient("transient-service", func(i Injector) (int, error) {
		return 42, nil
	})
	frame3, ok3 := inferServiceProviderStacktrace(transientService)
	is.False(ok3)
	is.Equal(stacktrace.Frame{}, frame3)

	// Test with alias service
	i := New()
	Provide(i, func(i Injector) (int, error) {
		return 42, nil
	})
	aliasService := newServiceAlias[int, int]("alias-service", i, "int")
	frame4, ok4 := inferServiceProviderStacktrace(aliasService)
	is.True(ok4)
	is.NotNil(frame4)
	is.Contains(frame4.File, "service_alias.go") // Provider frame points to the service_alias.go file
	is.Contains(frame4.Function, "newServiceAlias")
}

func TestInferServiceInfo(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with lazy service
	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})

	info, ok := inferServiceInfo(i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK")
	is.True(ok)
	is.Equal("*github.com/samber/do/v2.lazyTestHeathcheckerOK", info.name)
	is.Equal(ServiceTypeLazy, info.serviceType)
	is.Equal(time.Duration(0), info.serviceBuildTime) // Not built yet
	is.False(info.healthchecker)                      // Not built yet
	is.False(info.shutdowner)                         // Not built yet

	// Build the service
	_, _ = Invoke[*lazyTestHeathcheckerOK](i)

	info2, ok2 := inferServiceInfo(i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK")
	is.True(ok2)
	is.Equal("*github.com/samber/do/v2.lazyTestHeathcheckerOK", info2.name)
	is.Equal(ServiceTypeLazy, info2.serviceType)
	is.Positive(info2.serviceBuildTime) // Should have build time now
	is.True(info2.healthchecker)        // Should be healthchecker now
	is.False(info2.shutdowner)          // Not a shutdowner

	// Test with eager service
	i2 := New()
	ProvideValue(i2, &eagerTest{foobar: "foobar"})

	info3, ok3 := inferServiceInfo(i2, "*github.com/samber/do/v2.eagerTest")
	is.True(ok3)
	is.Equal("*github.com/samber/do/v2.eagerTest", info3.name)
	is.Equal(ServiceTypeEager, info3.serviceType)
	is.Equal(time.Duration(0), info3.serviceBuildTime) // Eager services don't have build time
	is.False(info3.healthchecker)
	is.False(info3.shutdowner)

	// Test with transient service
	i3 := New()
	ProvideTransient(i3, func(i Injector) (int, error) {
		return 42, nil
	})

	info4, ok4 := inferServiceInfo(i3, "int")
	is.True(ok4)
	is.Equal("int", info4.name)
	is.Equal(ServiceTypeTransient, info4.serviceType)
	is.Equal(time.Duration(0), info4.serviceBuildTime) // Transient services don't have build time
	is.False(info4.healthchecker)
	is.False(info4.shutdowner)

	// Test with service that doesn't exist
	info5, ok5 := inferServiceInfo(i, "non-existent")
	is.False(ok5)
	is.Equal(serviceInfo{}, info5)
}

func TestDoServiceCanCastToGeneric(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
