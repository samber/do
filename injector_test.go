package do

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerNew(t *testing.T) {
	is := assert.New(t)

	i := New()
	is.NotNil(i)
	is.Empty(i.services)
}

func TestContainerGet(t *testing.T) {
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

func TestContainerSet(t *testing.T) {
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

func TestContainerRemove(t *testing.T) {
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

func TestContainerServiceNotFound(t *testing.T) {
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

	expected := fmt.Errorf("DI: could not find service `hello`, available services: `foo`, `bar`")
	is.Equal(expected, i.serviceNotFound("hello"))
}

func TestContainerOnServiceInvoke(t *testing.T) {
	is := assert.New(t)

	i := New()

	i.onServiceInvoke("foo")
	i.onServiceInvoke("bar")

	is.Equal(0, i.orderedInvocation["foo"])
	is.Equal(1, i.orderedInvocation["bar"])
	is.Equal(2, i.orderedInvocationIndex)
}
