package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////////////
// 							Explicit aliases
/////////////////////////////////////////////////////////////////////////////

func TestAs(t *testing.T) {
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i))
	is.EqualError(As[*lazyTestShutdownerOK, Healthchecker](i), "DI: `*github.com/samber/do/v2.lazyTestShutdownerOK` is not `github.com/samber/do/v2.Healthchecker`")
	is.EqualError(As[*lazyTestHeathcheckerKO, Healthchecker](i), "DI: service `*github.com/samber/do/v2.lazyTestHeathcheckerKO` has not been declared")
	is.EqualError(As[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i), "DI: service `*github.com/samber/do/v2.lazyTestShutdownerOK` has not been declared")
}

func TestMustAs(t *testing.T) {
	// @TODO
}

func TestAsNamed(t *testing.T) {
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.Nil(AsNamed[*lazyTestHeathcheckerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK", "github.com/samber/do/v2.Healthchecker"))
	is.EqualError(AsNamed[*lazyTestShutdownerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "github.com/samber/do/v2.Healthchecker"), "DI: `*github.com/samber/do/v2.lazyTestShutdownerOK` is not `github.com/samber/do/v2.Healthchecker`")
	is.EqualError(AsNamed[*lazyTestHeathcheckerKO, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerKO", "github.com/samber/do/v2.Healthchecker"), "DI: service `*github.com/samber/do/v2.lazyTestHeathcheckerKO` has not been declared")
	is.EqualError(AsNamed[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "*github.com/samber/do/v2.lazyTestShutdownerOK"), "DI: service `*github.com/samber/do/v2.lazyTestShutdownerOK` has not been declared")
}

func TestMustAsNamed(t *testing.T) {
	// @TODO
}

/////////////////////////////////////////////////////////////////////////////
// 							Implicit aliases
/////////////////////////////////////////////////////////////////////////////

func TestInvokeAs(t *testing.T) {
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "hello world"}, nil
	})

	// found
	svc0, err := InvokeAs[*lazyTestHeathcheckerOK](i)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc0)
	is.Nil(err)

	// found via interface
	svc1, err := InvokeAs[Healthchecker](i)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc1)
	is.Nil(err)

	// not found
	svc2, err := InvokeAs[Shutdowner](i)
	is.Empty(svc2)
	is.EqualError(err, "DI: could not find service satisfying interface `github.com/samber/do/v2.Shutdowner`, available services: `*github.com/samber/do/v2.lazyTestHeathcheckerOK`")
}

type otherHealthChecker struct {
	foobar string
}

func (o *otherHealthChecker) HealthCheck() error {
	return nil
}

func TestInvokeAs_picksLatest(t *testing.T) {
	is := assert.New(t)

	lazy := "hello world"
	other := "hello from other"

	matches := 0
	failure := 0
	for i := 0; i < 100; i++ {
		injector := New()
		Provide(injector, func(i Injector) (*lazyTestHeathcheckerOK, error) {
			return &lazyTestHeathcheckerOK{foobar: lazy}, nil
		})
		Provide(injector, func(i Injector) (*otherHealthChecker, error) {
			return &otherHealthChecker{foobar: other}, nil
		})

		// should find via interface
		svc, err := InvokeAs[Healthchecker](injector)
		if is.EqualValues(&otherHealthChecker{foobar: other}, svc, "iteration %d should have returned the other health checker", i) {
			matches++
		} else {
			failure++
		}
		is.Nil(err)
	}

	is.Equal(100, matches, "all iterations should have returned the other health checker")
	is.Equal(0, failure, "no iterations should have failed")
}

func TestInvokeAs_picksLatestWithOverrides(t *testing.T) {
	is := assert.New(t)

	lazy := "hello world"
	other := "hello from other"
	third := "hello from third, still other"

	matches := 0
	failure := 0
	for i := 0; i < 100; i++ {
		injector := New()
		Provide(injector, func(i Injector) (*lazyTestHeathcheckerOK, error) {
			return &lazyTestHeathcheckerOK{foobar: lazy}, nil
		})
		Provide(injector, func(i Injector) (*otherHealthChecker, error) {
			return &otherHealthChecker{foobar: other}, nil
		})
		Override(injector, func(i Injector) (*otherHealthChecker, error) {
			return &otherHealthChecker{foobar: third}, nil
		})

		// should find via interface
		svc, err := InvokeAs[Healthchecker](injector)
		if is.EqualValues(&otherHealthChecker{foobar: third}, svc, "iteration %d should have returned the other health checker", i) {
			matches++
		} else {
			failure++
		}
		is.Nil(err)
	}

	is.Equal(100, matches, "all iterations should have returned the other health checker")
	is.Equal(0, failure, "no iterations should have failed")
}

func TestMustInvokeAs(t *testing.T) {
	// @TODO
}
