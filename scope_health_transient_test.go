package do

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type transientHealthBoom struct{}

func (transientHealthBoom) HealthCheck(context.Context) error {
	return errors.New("boom")
}

func TestScope_HealthCheckSkipsTransient(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()
	Provide(i, func(Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "ok"}, nil
	})
	ProvideTransient(i, func(Injector) (transientHealthBoom, error) {
		return transientHealthBoom{}, nil
	})

	_, err := Invoke[*lazyTestHeathcheckerOK](i)
	is.NoError(err)

	res := i.HealthCheckWithContext(context.Background())
	_, hasTransient := res[NameOf[transientHealthBoom]()]
	is.False(hasTransient, "transient must not appear as a false-healthy entry")
	is.Contains(res, NameOf[*lazyTestHeathcheckerOK]())
	is.NoError(res[NameOf[*lazyTestHeathcheckerOK]()])
}
