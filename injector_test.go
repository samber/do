package do

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInjectorNew(t *testing.T) {
	is := assert.New(t)

	i := New()
	is.NotNil(i)
	is.Empty(i.services)
}

func TestInjectorNewWithOpts(t *testing.T) {
	is := assert.New(t)

	count := 0

	i := NewWithOpts(&InjectorOpts{
		HookAfterRegistration: func(injector *Injector, serviceName string) {
			is.Equal("foobar", serviceName)
			count++
		},
		HookAfterShutdown: func(injector *Injector, serviceName string) {
			is.Equal("foobar", serviceName)
			count++
		},
	})

	ProvideNamedValue(i, "foobar", 42)

	is.NotPanics(func() {
		MustInvokeNamed[int](i, "foobar")
	})

	i.Shutdown()

	is.Equal(2, count)
}

func TestInjectorListProvidedServices(t *testing.T) {
	is := assert.New(t)

	i := New()

	is.NotPanics(func() {
		ProvideValue[int](i, 42)
		ProvideValue[float64](i, 21)
	})

	is.NotPanics(func() {
		services := i.ListProvidedServices()
		is.ElementsMatch([]string{"int", "float64"}, services)
	})
}

func TestInjectorListInvokedServices(t *testing.T) {
	is := assert.New(t)

	i := New()

	is.NotPanics(func() {
		ProvideValue[int](i, 42)
		ProvideValue[float64](i, 21)
		MustInvoke[int](i)
	})

	is.NotPanics(func() {
		services := i.ListInvokedServices()
		is.Equal([]string{"int"}, services)
	})
}

type testHealthCheck struct {
	foobar string
}

func (t *testHealthCheck) HealthCheck() error {
	return fmt.Errorf("broken")
}

func TestInjectorHealthCheck(t *testing.T) {
	is := assert.New(t)

	i := New()

	is.NotPanics(func() {
		ProvideValue[int](i, 42)
		ProvideNamed(i, "testHealthCheck", func(i *Injector) (*testHealthCheck, error) {
			return &testHealthCheck{}, nil
		})
	})

	// before invocation
	is.NotPanics(func() {
		got := i.HealthCheck()
		expected := map[string]error{
			"int":             nil,
			"testHealthCheck": nil,
		}

		is.Equal(expected, got)
	})

	is.NotPanics(func() {
		MustInvokeNamed[int](i, "int")
		MustInvokeNamed[*testHealthCheck](i, "testHealthCheck")
	})

	// after invocation
	is.NotPanics(func() {
		got := i.HealthCheck()
		expected := map[string]error{
			"int":             nil,
			"testHealthCheck": fmt.Errorf("broken"),
		}

		is.Equal(expected, got)
	})
}

func TestInjectorGet(t *testing.T) {
	is := assert.New(t)

	i := New()

	service := &ServiceEager[int]{
		name:     "foobar",
		instance: 42,
	}
	i.services["foobar"] = service

	// existing service
	{
		s1, ok1 := i.get("foobar")
		is.True(ok1)
		is.NotEmpty(s1)
		if ok1 {
			s, ok := s1.(Service[int])
			is.True(ok)

			v, err := s.getInstance(i)
			is.Nil(err)
			is.Equal(42, v)
		}
	}

	// not existing service
	{
		s2, ok2 := i.get("foobaz")
		is.False(ok2)
		is.Empty(s2)
	}
}

func TestInjectorSet(t *testing.T) {
	is := assert.New(t)

	i := New()

	service1 := &ServiceEager[int]{
		name:     "foobar",
		instance: 42,
	}

	service2 := &ServiceEager[int]{
		name:     "foobar",
		instance: 21,
	}

	i.set("foobar", service1)
	is.Len(i.services, 1)

	s1, ok1 := i.services["foobar"]
	is.True(ok1)
	is.True(reflect.DeepEqual(service1, s1))

	// erase
	i.set("foobar", service2)
	is.Len(i.services, 1)

	s2, ok2 := i.services["foobar"]
	is.True(ok2)
	is.True(reflect.DeepEqual(service2, s2))
}

func TestInjectorRemove(t *testing.T) {
	is := assert.New(t)

	i := New()

	service := &ServiceEager[int]{
		name:     "foobar",
		instance: 42,
	}

	i.set("foobar", service)
	is.Len(i.services, 1)
	i.remove("foobar")
	is.Len(i.services, 0)
}

func TestInjectorServiceNotFound(t *testing.T) {
	is := assert.New(t)

	i := New()

	service1 := &ServiceEager[int]{
		name:     "foo",
		instance: 42,
	}

	service2 := &ServiceEager[int]{
		name:     "bar",
		instance: 21,
	}

	i.set("foo", service1)
	i.set("bar", service2)
	is.Len(i.services, 2)

	err := i.serviceNotFound("hello")
	is.ErrorContains(err, "DI: could not find service `hello`, available services:")
	is.ErrorContains(err, "`foo`")
	is.ErrorContains(err, "`bar`")
}

func TestInjectorOnServiceInvoke(t *testing.T) {
	is := assert.New(t)

	i := New()

	i.onServiceInvoke("foo")
	i.onServiceInvoke("bar")

	is.Equal(0, i.orderedInvocation["foo"])
	is.Equal(1, i.orderedInvocation["bar"])
	is.Equal(2, i.orderedInvocationIndex)
}
