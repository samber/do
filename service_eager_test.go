package do

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type eagerTest struct {
	foobar string
}

var _ Healthchecker = (*eagerTestHeathcheckerOK)(nil)

type eagerTestHeathcheckerOK struct {
	foobar string
}

func (t *eagerTestHeathcheckerOK) HealthCheck() error {
	return nil
}

var _ Healthchecker = (*eagerTestHeathcheckerKO)(nil)

type eagerTestHeathcheckerKO struct {
	foobar string
}

func (t *eagerTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

var _ ShutdownerWithContextAndError = (*eagerTestShutdownerOK)(nil)

type eagerTestShutdownerOK struct {
	foobar string
}

func (t *eagerTestShutdownerOK) Shutdown(ctx context.Context) error {
	return nil
}

var _ ShutdownerWithError = (*eagerTestShutdownerKO)(nil)

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
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceEager("foobar2", test)
	is.Equal("foobar2", service2.getName())
}

func TestServiceEager_getTypeName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("int", service1.getTypeName())

	service2 := newServiceEager("foobar2", test)
	is.Equal("github.com/samber/do/v2.eagerTest", service2.getTypeName())
}

func TestServiceEager_getServiceType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal(ServiceTypeEager, service1.getServiceType())

	service2 := newServiceEager("foobar2", test)
	is.Equal(ServiceTypeEager, service2.getServiceType())
}

func TestServiceEager_getReflectType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("int", service1.getReflectType().String())

	service2 := newServiceEager("foobar2", test)
	is.Equal("do.eagerTest", service2.getReflectType().String())

	service3 := newServiceEager("foobar3", (Healthchecker)(nil))
	is.Equal("do.Healthchecker", service3.getReflectType().String())

	service4 := newServiceEager[Healthchecker]("foobar4", nil)
	is.Equal("do.Healthchecker", service4.getReflectType().String())

	service5 := newServiceEager("foobar1", &test)
	is.Equal("*do.eagerTest", service5.getReflectType().String())
}

func TestServiceEager_getInstanceAny(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar", test)
	instance1, err1 := service1.getInstanceAny(nil)
	is.Nil(err1)
	is.Equal(test, instance1)

	service2 := newServiceEager("foobar", 42)
	instance2, err2 := service2.getInstanceAny(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}

func TestServiceEager_getInstance(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar", test)
	instance1, err1 := service1.getInstance(nil)
	is.Nil(err1)
	is.Equal(test, instance1)

	service2 := newServiceEager("foobar", 42)
	instance2, err2 := service2.getInstance(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}

func TestServiceEager_isHealthchecker(t *testing.T) {
	t.Parallel()
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

// @TODO: missing tests for context
func TestServiceEager_healthcheck(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.healthcheck(ctx)
	is.Nil(err1)

	// healthcheck ok
	service2 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err2 := service2.healthcheck(ctx)
	is.Nil(err2)

	// healthcheck ko
	service3 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	err3 := service3.healthcheck(ctx)
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)
}

func TestServiceEager_isShutdowner(t *testing.T) {
	t.Parallel()
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

// @TODO: missing tests for context
func TestServiceEager_shutdown(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.shutdown(ctx)
	is.Nil(err1)

	// shutdown ok
	service2 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err2 := service2.shutdown(ctx)
	is.Nil(err2)

	// shutdown ko
	service3 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	err3 := service3.shutdown(ctx)
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)
}

func TestServiceEager_clone(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	// initial
	service1 := newServiceEager("foobar", test)
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone().(*serviceEager[eagerTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceEager_source(t *testing.T) {
	// @TODO
}
