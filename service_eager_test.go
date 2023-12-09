package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceEagerName(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	service1 := newServiceEager("foobar", 42)
	is.Equal("foobar", service1.getName())

	service2 := newServiceEager("foobar", _test)
	is.Equal("foobar", service2.getName())
}

func TestServiceEagerInstance(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	service1 := newServiceEager("foobar", _test)
	is.Equal(&ServiceEager[test]{name: "foobar", instance: _test}, service1)

	instance1, err1 := service1.getInstance(nil)
	is.Nil(err1)
	is.Equal(_test, instance1)

	service2 := newServiceEager("foobar", 42)
	is.Equal(&ServiceEager[int]{name: "foobar", instance: 42}, service2)

	instance2, err2 := service2.getInstance(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}
