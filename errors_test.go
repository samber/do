package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShutdownErrors_Add(t *testing.T) {
	is := assert.New(t)

	se := newShutdownErrors()
	is.Equal(0, len(*se))
	is.Equal(0, se.Len())

	se.Add("scope-1", "scope-a", "service-a", nil)
	is.Equal(0, len(*se))
	is.Equal(0, se.Len())
	is.EqualValues(&ShutdownErrors{}, se)

	se.Add("scope-2", "scope-b", "service-b", assert.AnError)
	is.Equal(1, len(*se))
	is.Equal(1, se.Len())
	is.EqualValues(&ShutdownErrors{
		{ScopeID: "scope-2", ScopeName: "scope-b", Service: "service-b"}: assert.AnError,
	}, se)
}

func TestShutdownErrors_Error(t *testing.T) {
	is := assert.New(t)

	se := newShutdownErrors()
	is.Equal(0, len(*se))
	is.Equal(0, se.Len())
	is.EqualValues("DI: no shutdown errors", se.Error())

	se.Add("scope-1", "scope-a", "service-a", nil)
	is.Equal(0, len(*se))
	is.Equal(0, se.Len())
	is.EqualValues("DI: no shutdown errors", se.Error())

	se.Add("scope-2", "scope-b", "service-b", assert.AnError)
	is.Equal(1, len(*se))
	is.Equal(1, se.Len())
	is.EqualValues("DI: shutdown errors:\n  - scope-b > service-b: assert.AnError general error for testing", se.Error())
}

func TestMergeShutdownErrors(t *testing.T) {
	is := assert.New(t)

	se1 := newShutdownErrors()
	se2 := newShutdownErrors()
	se3 := newShutdownErrors()

	se1.Add("scope-1", "scope-a", "service-a", assert.AnError)
	se2.Add("scope-2", "scope-b", "service-b", assert.AnError)

	result := mergeShutdownErrors(se1, se2, se3, nil)
	is.Equal(2, result.Len())
	is.EqualValues(
		&ShutdownErrors{
			{ScopeID: "scope-1", ScopeName: "scope-a", Service: "service-a"}: assert.AnError,
			{ScopeID: "scope-2", ScopeName: "scope-b", Service: "service-b"}: assert.AnError,
		},
		result,
	)
}
