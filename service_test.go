package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServiceName(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	name := generateServiceName[test]()
	is.Equal("do.test", name)

	name = generateServiceName[*test]()
	is.Equal("*do.test", name)

	name = generateServiceName[int]()
	is.Equal("int", name)
}
