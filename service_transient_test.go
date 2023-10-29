package do

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type transientTest struct {
	foobar string
}

// nolint:unused
type transientTestHeathcheckerOK struct {
	foobar string
}

// nolint:unused
func (t *transientTestHeathcheckerOK) HealthCheck() error {
	return nil
}

// nolint:unused
type transientTestHeathcheckerKO struct {
	foobar string
}

// nolint:unused
func (t *transientTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

// nolint:unused
type transientTestShutdownerOK struct {
	foobar string
}

// nolint:unused
func (t *transientTestShutdownerOK) Shutdown() error {
	return nil
}

// nolint:unused
type transientTestShutdownerKO struct {
	foobar string
}

// nolint:unused
func (t *transientTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

func TestNewServiceTransient(t *testing.T) {
	// @TODO
}

func TestServiceTransient_getName(t *testing.T) {
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal("foobar2", service2.getName())
}

func TestServiceTransient_getType(t *testing.T) {
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal(ServiceTypeTransient, service1.getType())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal(ServiceTypeTransient, service2.getType())
}

func TestServiceTransient_getInstance(t *testing.T) {
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}
	provider3 := func(i Injector) (int, error) {
		panic("error")
	}
	provider4 := func(i Injector) (int, error) {
		panic(fmt.Errorf("error"))
	}
	provider5 := func(i Injector) (int, error) {
		return 42, fmt.Errorf("error")
	}

	i := New()

	// basic type
	service1 := newServiceTransient("foobar", provider1).(*ServiceTransient[int])
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceTransient("hello", provider2).(*ServiceTransient[transientTest])
	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service3 := newServiceTransient("baz", provider3)
		_, _ = service3.getInstance(i)
	})

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service4 := newServiceTransient("plop", provider4)
		instance4, err4 := service4.getInstance(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceTransient("plop", provider5)
		instance5, err5 := service5.getInstance(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

func TestServiceTransient_isHealthchecker(t *testing.T) {
	// @TODO
}

func TestServiceTransient_healthcheck(t *testing.T) {
	// @TODO
}

func TestServiceTransient_isShutdowner(t *testing.T) {
	// @TODO
}

func TestServiceTransient_shutdown(t *testing.T) {
	// @TODO
}

func TestServiceTransient_clone(t *testing.T) {
	// @TODO
	is := assert.New(t)

	// initial
	service1 := newServiceTransient("foobar", func(i Injector) (transientTest, error) {
		return transientTest{foobar: "foobar"}, nil
	}).(*ServiceTransient[transientTest])
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone().(*ServiceTransient[transientTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceTransient_source(t *testing.T) {
	// @TODO
}
