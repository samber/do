package do

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandleProviderPanic(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	is.NotPanics(func() {
		// ok
		svc, err := handleProviderPanic(func(i Injector) (int, error) {
			return 42, nil
		}, nil)
		is.Nil(err)
		is.Equal(42, svc)

		// should return an error
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			return 0, assert.AnError
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)

		// shoud not return 42
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			return 42, assert.AnError
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)

		// panics with string
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			panic("aïe")
		}, nil)
		is.EqualError(err, "DI: aïe")
		is.Equal(0, svc)

		// panics with error
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			panic(assert.AnError)
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)
	})
}
