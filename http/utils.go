package http

import (
	"bytes"
	"text/template"

	"github.com/samber/do/v2"
)

func fromTemplate(tpl string, data any) (string, error) {
	t := template.Must(template.New("").Parse(tpl))
	var buf bytes.Buffer

	err := t.Execute(&buf, data) // 🤮
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func mAp[T any, R any](collection []T, iteratee func(T) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item)
	}

	return result
}

func getScopeByID(injector do.Injector, id string) (do.DescriptionInjectorScope, bool) {
	scopes := getAllScopes(injector)
	for _, scope := range scopes {
		if scope.ScopeID == id {
			return scope, true
		}
	}
	return do.DescriptionInjectorScope{}, false
}

func getAllScopes(injector do.Injector) []do.DescriptionInjectorScope {
	description := do.DescribeInjector(injector)

	return getAllScopesRec(description.DAG)
}

func getAllScopesRec(scopes []do.DescriptionInjectorScope) []do.DescriptionInjectorScope {
	output := []do.DescriptionInjectorScope{}
	for i := range scopes {
		output = append(output, scopes[i])
		output = append(output, getAllScopesRec(scopes[i].Children)...)
	}
	return output
}
