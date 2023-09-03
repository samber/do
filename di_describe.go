package do

import (
	"bytes"
	"html/template"
	"sort"
	"strings"
)

/////////////////////////////////////////////////////////////////////////////
// 							Templating helpers
/////////////////////////////////////////////////////////////////////////////

func fromTemplate(tpl string, data any) string {
	t := template.Must(template.New("").Parse(tpl))
	var buf bytes.Buffer
	must0(t.Execute(&buf, data)) // 🤮
	return buf.String()
}

/////////////////////////////////////////////////////////////////////////////
// 							Describe services
/////////////////////////////////////////////////////////////////////////////

var describeServiceTemplate = `
Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}
Service: {{.ServiceName}}
Service type: {{.ServiceType}}
Invoked: {{.Invoked}}

Dependencies:{{.Dependencies}}

Dependents:{{.Dependents}}
`

// DescribeService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before DescribeService[T] to ensure that the service is registered.
func DescribeService[T any](i Injector) (output string, ok bool) {
	name := inferServiceName[T]()
	return DescribeNamedService[T](i, name)
}

// DescribeNamedService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before DescribeNamedService[T] to ensure that the service is registered.
func DescribeNamedService[T any](scope Injector, name string) (output string, ok bool) {
	_i := getInjectorOrDefault(scope)

	serviceAny, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return "", false
	}

	service, ok := serviceAny.(Service[T])
	if !ok {
		return "", false
	}

	invoked := ""
	frame, ok := inferServiceStacktrace[T](service)
	if ok {
		invoked = frame.String()
	}

	return fromTemplate(
		describeServiceTemplate,
		map[string]any{
			"ScopeID":      serviceScope.ID(),
			"ScopeName":    serviceScope.Name(),
			"ServiceName":  name,
			"ServiceType":  inferServiceType[T](service),
			"Invoked":      invoked,
			"Dependencies": buildDepsGraph(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependencies"),
			"Dependents":   buildDepsGraph(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependents"),
		},
	), true
}

var describeServiceDepsTemplate = `{{range .}}
* {{.Service}} from scope {{.ScopeName}}{{.Recursive}}{{end}}`

func buildDepsGraph(i Injector, edge EdgeService, mode string) string {
	dependencies, dependents := i.RootScope().dag.explainService(edge.ScopeID, edge.ScopeName, edge.Service)

	deps := dependencies
	if mode == "dependents" {
		deps = dependents
	}

	// order by id to have a deterministic output in unit tests
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Service < deps[j].Service
	})

	return fromTemplate(
		describeServiceDepsTemplate,
		mAp(deps, func(item EdgeService, _ int) map[string]any {
			output := buildDepsGraph(i, item, mode)

			// add tab prefix
			lines := mAp(strings.Split(output, "\n"), func(line string, index int) string {
				if index > 0 {
					return "  " + line
				}
				return line
			})

			return map[string]any{
				"ScopeID":   item.ScopeID,
				"ScopeName": item.ScopeName,
				"Service":   item.Service,
				"Recursive": strings.Join(lines, "\n"),
			}
		}),
	)
}

/////////////////////////////////////////////////////////////////////////////
// 							Describe scopes
/////////////////////////////////////////////////////////////////////////////

var describeInjectorTemplate = `
Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}

DAG:{{.DAG}}
`

const scopePrefixTemplate = "    "

func DescribeInjector(scope Injector) (output string, ok bool) {
	_i := getInjectorOrDefault(scope)

	ancestors := append([]Injector{scope}, castScopesToInjectors(scope.Ancestors())...)
	reverseSlice(ancestors) // root scope first

	return fromTemplate(
		describeInjectorTemplate,
		map[string]any{
			"ScopeID":   _i.ID(),
			"ScopeName": _i.Name(),
			"DAG":       buildInjectorScopes(ancestors, castScopesToInjectors(scope.Children())),
		},
	), true
}

var describeInjectorItemTemplate = `{{range .}}

\_ {{.ScopeName}} (ID: {{.ScopeID}}){{.Services}}
{{.Children}}{{end}}`

// 2 modes are available: looping on ancestors, focused-scope or children
func buildInjectorScopes(ancestors []Injector, children []Injector) string {
	loopingOn := "children" // @TODO: create a real enum
	injectors := children
	if len(ancestors) > 0 {
		injectors = []Injector{ancestors[0]}
		ancestors = ancestors[1:]
		if len(ancestors) == 0 {
			loopingOn = "focused-scope"
		} else {
			loopingOn = "ancestors"
		}
	}

	// order by id to have a deterministic output in unit tests
	sort.Slice(injectors, func(i, j int) bool {
		return injectors[i].ID() < injectors[j].ID()
	})

	seenScope := 0

	return fromTemplate(
		describeInjectorItemTemplate,
		mAp(injectors, func(item Injector, _ int) map[string]any {
			nextChildren := children
			if loopingOn == "children" {
				nextChildren = castScopesToInjectors(item.Children())
			}

			lines := strings.Split(buildInjectorScopes(ancestors, nextChildren), "\n")
			indented := strings.Join(mAp(lines, func(line string, i int) string {
				isFirstLineOfScope := strings.HasPrefix(line, "\\_")
				if isFirstLineOfScope {
					seenScope++
				}

				if loopingOn == "children" && len(nextChildren) > 0 && seenScope < len(nextChildren) {
					return scopePrefixTemplate + "|" + line
				} else if loopingOn == "focused-scope" && seenScope < len(children) {
					return scopePrefixTemplate + "|" + line
				} else if loopingOn == "ancestors" && seenScope < len(injectors) {
					return scopePrefixTemplate + "|" + line
				}

				return scopePrefixTemplate + " " + line
			}), "\n")

			return map[string]any{
				"ScopeID":   item.ID(),
				"ScopeName": item.Name(),
				"Services":  buildInjectorServicesList(item),
				"Children":  indented,
			}
		}),
	)
}

var describeInjectorServicesTemplate = `{{range .}}
    * {{.ServiceType}}{{.ServiceName}}{{.ServiceFeatures}}{{end}}`

func buildInjectorServicesList(i Injector) string {
	services := i.ListProvidedServices()
	services = filter(services, func(item EdgeService, _ int) bool {
		return i.serviceExist(item.Service)
	})

	// order by id to have a deterministic output in unit tests
	sort.Slice(services, func(i, j int) bool {
		return services[i].Service < services[j].Service
	})

	return fromTemplate(
		describeInjectorServicesTemplate,
		mAp(services, func(item EdgeService, _ int) map[string]any {
			prefix := ""
			suffix := ""

			if info, ok := inferServiceInfo(i, item.Service); ok {
				prefix += serviceTypeToIcon[info.serviceType] + " "
				if info.healthchecker {
					suffix += " 🏥"
				}
				if info.shutdowner {
					suffix += " 🙅"
				}
			} else {
				prefix += "❓ " // should never reach this branch
			}

			return map[string]any{
				"ServiceName":     item.Service,
				"ServiceType":     prefix,
				"ServiceFeatures": suffix,
			}
		}),
	)
}

func castScopesToInjectors(scopes []*Scope) []Injector {
	return mAp(scopes, func(item *Scope, _ int) Injector {
		return item
	})
}
