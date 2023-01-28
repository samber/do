package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTag(t *testing.T) {
	is := assert.New(t)

	type foobar string

	type hello struct {
		world string
	}

	type test struct {
		foobar1 string `do:"baz"`
		foobar2 foobar `do:""`
		foobar3 *hello `do:""`
		foobar4 *hello `do`
		foobar5 string
	}

	injector := New()

	ProvideNamedValue(injector, "baz", "foobar1")
	ProvideValue[foobar](injector, "foobar2")
	Provide(injector, func(i *Injector) (*hello, error) {
		return &hello{"foobar3"}, nil
	})

	_, err := InjectTag(injector, test{})
	is.EqualError(err, "DI: expected a pointer")

	_, err = InjectTag(injector, 42)
	is.EqualError(err, "DI: expected a pointer")

	v, err := InjectTag(injector, &test{})
	is.Nil(err)
	is.Equal("foobar1", v.foobar1)
	is.Equal(foobar("foobar2"), v.foobar2)
	is.Equal("foobar3", v.foobar3.world)
	is.Nil(v.foobar4)
	is.Equal("", v.foobar5)

	type wrongType struct {
		foobar int `do:"baz"`
	}

	_, err = InjectTag(injector, &wrongType{})
	is.EqualError(err, "DI: type mismatch. Expected 'string', got 'int'")
}
