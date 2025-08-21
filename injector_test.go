package do

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetInjectorOrDefault(t *testing.T) {
	// t.Parallel() // parallel forbidden by write on DefaultRootScope
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	is.Equal(DefaultRootScope, getInjectorOrDefault(nil))
	is.NotEqual(DefaultRootScope, getInjectorOrDefault(New()))

	type test struct {
		foobar string
	}

	DefaultRootScope = New()

	Provide(nil, func(i Injector) (*test, error) {
		return &test{foobar: "42"}, nil
	})

	is.Len(DefaultRootScope.self.services, 1)
}
