package di

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtilsEmpty(t *testing.T) {
	is := assert.New(t)

	value1 := empty[int]()
	is.Empty(value1)

	value2 := empty[*int]()
	is.Nil(value2)
	is.Empty(value2)
}

func TestUtilsMust(t *testing.T) {
	is := assert.New(t)

	is.Panics(func() {
		must(fmt.Errorf("error"))
	})
	is.NotPanics(func() {
		must(nil)
	})
}
