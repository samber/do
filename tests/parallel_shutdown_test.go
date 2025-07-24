package tests

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/samber/do/v2/tests/fixtures"

	"github.com/stretchr/testify/assert"
)

func TestParallelShutdown(t *testing.T) {
	is := assert.New(t)

	root, driver, passenger := fixtures.GetPackage()
	is.NotPanics(func() {
		_ = do.MustInvoke[*fixtures.Driver](driver)
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-1")
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-2")
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-3")
		_ = root.Shutdown()
	})
}
