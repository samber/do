package dohttp

import (
	"strings"
	"testing"

	do "github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

func TestIndexHTML(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	basePath := "/debug/di"
	html, err := IndexHTML(basePath)
	is.NoError(err)
	is.NotEmpty(html)
	is.Contains(html, "Welcome to do UI")
	is.Contains(html, basePath+"/scope")
	is.Contains(html, basePath+"/service")
}

func TestScopeTreeHTML_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	basePath := "/debug/di"
	root := do.New()
	child := root.Scope("child")

	do.ProvideNamedValue(root, "db-service", "db-value")
	do.ProvideNamedTransient(root, "request-id", func(i do.Injector) (string, error) { return "rid", nil })
	do.ProvideNamedValue(child, "cache-service", 123)

	html, err := ScopeTreeHTML(basePath, root, "")
	is.NoError(err)
	is.NotEmpty(html)
	is.Contains(html, "Scope description")
	// Links to scopes and services should be present
	is.Contains(html, basePath+"/scope")
	is.True(strings.Contains(html, "db-service") || strings.Contains(html, "cache-service"))
}

func TestServiceListHTML_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	basePath := "/debug/di"
	root := do.New()
	_ = root // ensure non-nil
	do.ProvideNamedValue(root, "cfg", "x")

	html, err := ServiceListHTML(basePath, root)
	is.NoError(err)
	is.NotEmpty(html)
	is.Contains(html, "Service description")
	is.Contains(html, basePath+"/service")
}

func TestServiceHTML_Basic(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	basePath := "/debug/di"
	root := do.New()
	do.ProvideNamedValue(root, "cfg", "x")

	html, err := ServiceHTML(basePath, root, root.ID(), "cfg")
	is.NoError(err)
	is.NotEmpty(html)
	is.Contains(html, "Scope id:")
	is.Contains(html, "Service name: cfg")
	// eager value should report eager type
	is.Contains(html, "Service type: eager")
}
