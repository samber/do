package do

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type transiantTest struct {
	foobar string
}

// nolint:unused
type transiantTestHeathcheckerOK struct {
	foobar string
}

// nolint:unused
func (t *transiantTestHeathcheckerOK) HealthCheck() error {
	return nil
}

// nolint:unused
type transiantTestHeathcheckerKO struct {
	foobar string
}

// nolint:unused
func (t *transiantTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

// nolint:unused
type transiantTestShutdownerOK struct {
	foobar string
}

// nolint:unused
func (t *transiantTestShutdownerOK) Shutdown() error {
	return nil
}

// nolint:unused
type transiantTestShutdownerKO struct {
	foobar string
}

// nolint:unused
func (t *transiantTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

func TestNewServiceTransiant(t *testing.T) {
	// @TODO
}

func TestServiceTransiant_getName(t *testing.T) {
	is := assert.New(t)

	test := transiantTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transiantTest, error) {
		return test, nil
	}

	service1 := newServiceTransiant("foobar1", provider1)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceTransiant("foobar2", provider2)
	is.Equal("foobar2", service2.getName())
}

func TestServiceTransiant_getType(t *testing.T) {
	is := assert.New(t)

	test := transiantTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transiantTest, error) {
		return test, nil
	}

	service1 := newServiceTransiant("foobar1", provider1)
	is.Equal(ServiceTypeTransiant, service1.getType())

	service2 := newServiceTransiant("foobar2", provider2)
	is.Equal(ServiceTypeTransiant, service2.getType())
}

func TestServiceTransiant_getInstance(t *testing.T) {
	is := assert.New(t)

	test := transiantTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transiantTest, error) {
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
	service1 := newServiceTransiant("foobar", provider1).(*ServiceTransiant[int])
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceTransiant("hello", provider2).(*ServiceTransiant[transiantTest])
	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service3 := newServiceTransiant("baz", provider3)
		_, _ = service3.getInstance(i)
	})

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service4 := newServiceTransiant("plop", provider4)
		instance4, err4 := service4.getInstance(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceTransiant("plop", provider5)
		instance5, err5 := service5.getInstance(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

func TestServiceTransiant_isHealthchecker(t *testing.T) {
	// @TODO
}

func TestServiceTransiant_healthcheck(t *testing.T) {
	// @TODO
}

func TestServiceTransiant_isShutdowner(t *testing.T) {
	// @TODO
}

func TestServiceTransiant_shutdown(t *testing.T) {
	// @TODO
}

func TestServiceTransiant_clone(t *testing.T) {
	// @TODO
	is := assert.New(t)

	// initial
	service1 := newServiceTransiant("foobar", func(i Injector) (transiantTest, error) {
		return transiantTest{foobar: "foobar"}, nil
	}).(*ServiceTransiant[transiantTest])
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone().(*ServiceTransiant[transiantTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceTransiant_locate(t *testing.T) {
	// @TODO
}
