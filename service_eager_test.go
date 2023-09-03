package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type eagerTest struct {
	foobar string
}

type eagerTestHeathcheckerOK struct {
	foobar string
}

func (t *eagerTestHeathcheckerOK) HealthCheck() error {
	return nil
}

type eagerTestHeathcheckerKO struct {
	foobar string
}

func (t *eagerTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

type eagerTestShutdownerOK struct {
	foobar string
}

func (t *eagerTestShutdownerOK) Shutdown() error {
	return nil
}

type eagerTestShutdownerKO struct {
	foobar string
}

func (t *eagerTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

func TestNewServiceEager(t *testing.T) {
	// @TODO
}

func TestServiceEager_getName(t *testing.T) {
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceEager("foobar2", test)
	is.Equal("foobar2", service2.getName())
}

func TestServiceEager_getType(t *testing.T) {
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal(ServiceTypeEager, service1.getType())

	service2 := newServiceEager("foobar2", test)
	is.Equal(ServiceTypeEager, service2.getType())
}

func TestServiceEager_getInstance(t *testing.T) {
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar", test).(*ServiceEager[eagerTest])
	instance1, err1 := service1.getInstance(nil)
	is.Nil(err1)
	is.Equal(test, instance1)

	service2 := newServiceEager("foobar", 42).(*ServiceEager[int])
	instance2, err2 := service2.getInstance(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}

func TestServiceEager_isHealthchecker(t *testing.T) {
	is := assert.New(t)

	// no healthcheck
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	is.False(service1.isHealthchecker())

	// healthcheck ok
	service2 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	is.True(service2.isHealthchecker())

	// healthcheck ko
	service3 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	is.True(service3.isHealthchecker())
}

func TestServiceEager_healthcheck(t *testing.T) {
	is := assert.New(t)

	// no healthcheck
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.healthcheck()
	is.Nil(err1)

	// healthcheck ok
	service2 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err2 := service2.healthcheck()
	is.Nil(err2)

	// healthcheck ko
	service3 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	err3 := service3.healthcheck()
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)
}

func TestServiceEager_isShutdowner(t *testing.T) {
	is := assert.New(t)

	// no shutdown
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	is.False(service1.isShutdowner())

	// shutdown ok
	service2 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	is.True(service2.isShutdowner())

	// shutdown ko
	service3 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	is.True(service3.isShutdowner())
}

func TestServiceEager_shutdown(t *testing.T) {
	is := assert.New(t)

	// no shutdown
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.shutdown()
	is.Nil(err1)

	// shutdown ok
	service2 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err2 := service2.shutdown()
	is.Nil(err2)

	// shutdown ko
	service3 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	err3 := service3.shutdown()
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)
}

func TestServiceEager_clone(t *testing.T) {
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	// initial
	service1 := newServiceEager("foobar", test).(*ServiceEager[eagerTest])
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone().(*ServiceEager[eagerTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceEager_locate(t *testing.T) {
	// @TODO
}
