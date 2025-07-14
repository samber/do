package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServiceName(t *testing.T) {
	is := assert.New(t)

	type MyFunc func(int) int
	type itest interface {}
	type test struct{}

	name := generateServiceName[test]()
	is.Equal("do.test", name)

	name = generateServiceName[*test]()
	is.Equal("*do.test", name)

	name = generateServiceName[itest]()
	is.Equal("do.itest", name)

	name = generateServiceName[*itest]()
	is.Equal("*do.itest", name)

	name = generateServiceName[func(int)int]()
	is.Equal("func(int) int", name)

	name = generateServiceName[MyFunc]()
	is.Equal("do.MyFunc", name)

	name = generateServiceName[int]()
	is.Equal("int", name)

	name = generateServiceName[*int]()
	is.Equal("*int", name)

	name = generateServiceName[[]int]()
	is.Equal("[]int", name)

	name = generateServiceName[map[string]string]()
	is.Equal("map[string]string", name)

	name = generateServiceName[map[string]test]()
	is.Equal("map[string]do.test", name)
}

func TestGenerateServiceNameWithFQSN(t *testing.T) {
	is := assert.New(t)

	type MyFunc func(int) int
	type itest interface {}
	type test struct{}

	name := generateServiceNameWithFQSN[test]()
	is.Equal("github.com/samber/do.test", name)

	name = generateServiceNameWithFQSN[*test]()
	is.Equal("*github.com/samber/do.test", name)

	name = generateServiceNameWithFQSN[itest]()
	is.Equal("github.com/samber/do.itest", name)

	name = generateServiceNameWithFQSN[*itest]()
	is.Equal("*github.com/samber/do.itest", name)

	name = generateServiceNameWithFQSN[func(int)int]()
	is.Equal("func(int) int", name)

	name = generateServiceNameWithFQSN[MyFunc]()
	is.Equal("github.com/samber/do.MyFunc", name)

	name = generateServiceNameWithFQSN[int]()
	is.Equal("int", name)

	name = generateServiceNameWithFQSN[*int]()
	is.Equal("*int", name)

	name = generateServiceNameWithFQSN[[]int]()
	is.Equal("[]int", name)

	name = generateServiceNameWithFQSN[map[string]string]()
	is.Equal("map[string]string", name)
}
