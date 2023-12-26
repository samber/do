package do

import (
	"bytes"
	"html/template"
	"sort"
	"strings"

	"github.com/samber/do/v2/stacktrace"
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

const describeServiceTemplate = `
Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}

Service name: {{.ServiceName}}
Service type: {{.ServiceType}}
Invoked: {{.Invoked}}

Dependencies:
{{.Dependencies}}

Dependents:
{{.Dependents}}
`

// @TODO: add service type icon (lazy, eager, transient)
const describeServiceDependenciesTemplate = `* {{.Service}} from scope {{.ScopeName}}{{.Recursive}}`

type DescriptionService struct {
	ScopeID      string                         `json:"scope_id"`
	ScopeName    string                         `json:"scope_name"`
	ServiceName  string                         `json:"service_name"`
	ServiceType  ServiceType                    `json:"service_type"`
	Invoked      *stacktrace.Frame              `json:"invoked"`
	Dependencies []DescriptionServiceDependency `json:"dependencies"`
	Dependents   []DescriptionServiceDependency `json:"dependents"`
}

func (sd *DescriptionService) String() string {
	invoked := ""
	if sd.Invoked != nil {
		invoked = sd.Invoked.String()
	}

	return fromTemplate(
		describeServiceTemplate,
		map[string]string{
			"ScopeID":     sd.ScopeID,
			"ScopeName":   sd.ScopeName,
			"ServiceName": sd.ServiceName,
			"ServiceType": string(sd.ServiceType),
			"Invoked":     invoked,
			"Dependencies": strings.Join(
				mAp(sd.Dependencies, func(item DescriptionServiceDependency, _ int) string {
					return item.String()
				}),
				"\n",
			),
			"Dependents": strings.Join(
				mAp(sd.Dependents, func(item DescriptionServiceDependency, _ int) string {
					return item.String()
				}),
				"\n",
			),
		},
	)
}

type DescriptionServiceDependency struct {
	ScopeID   string                         `json:"scope_id"`
	ScopeName string                         `json:"scope_name"`
	Service   string                         `json:"service"`
	Recursive []DescriptionServiceDependency `json:"recursive"`
}

func (sdd *DescriptionServiceDependency) String() string {
	lines := flatten(
		mAp(sdd.Recursive, func(item DescriptionServiceDependency, _ int) []string {
			return mAp(
				strings.Split(item.String(), "\n"),
				func(line string, _ int) string {
					return "  " + line
				},
			)
		}),
	)

	recursive := strings.Join(lines, "\n")
	if len(lines) > 0 {
		recursive = "\n" + recursive
	}

	return fromTemplate(
		describeServiceDependenciesTemplate,
		map[string]string{
			"ScopeID":   sdd.ScopeID,
			"ScopeName": sdd.ScopeName,
			"Service":   sdd.Service,
			"Recursive": recursive,
		},
	)
}

// DescribeService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before DescribeService[T] to ensure that the service is registered.
func DescribeService[T any](i Injector) (output DescriptionService, ok bool) {
	name := inferServiceName[T]()
	return DescribeNamedService(i, name)
}

// DescribeNamedService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before DescribeNamedService[T] to ensure that the service is registered.
func DescribeNamedService(scope Injector, name string) (output DescriptionService, ok bool) {
	_i := getInjectorOrDefault(scope)

	serviceAny, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return DescriptionService{}, false
	}

	service, ok := serviceAny.(ServiceAny)
	if !ok {
		return DescriptionService{}, false
	}

	var invoked *stacktrace.Frame
	frame, ok := inferServiceProviderStacktrace(service)
	if ok {
		invoked = &frame
	}

	return DescriptionService{
		ScopeID:      serviceScope.ID(),
		ScopeName:    serviceScope.Name(),
		ServiceName:  name,
		ServiceType:  service.getType(),
		Invoked:      invoked,
		Dependencies: newDescriptionServiceDependencies(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependencies"),
		Dependents:   newDescriptionServiceDependencies(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependents"),
	}, true
}

func newDescriptionServiceDependencies(i Injector, edge EdgeService, mode string) []DescriptionServiceDependency {
	dependencies, dependents := i.RootScope().dag.explainService(edge.ScopeID, edge.ScopeName, edge.Service)

	deps := dependencies
	if mode == "dependents" {
		deps = dependents
	}

	// order by id to have a deterministic output in unit tests
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Service < deps[j].Service
	})

	return mAp(deps, func(item EdgeService, _ int) DescriptionServiceDependency {
		recursive := newDescriptionServiceDependencies(i, item, mode)

		// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
		return DescriptionServiceDependency{
			ScopeID:   item.ScopeID,
			ScopeName: item.ScopeName,
			Service:   item.Service,
			Recursive: recursive,
		}
	})
}

/////////////////////////////////////////////////////////////////////////////
// 							Describe scopes
/////////////////////////////////////////////////////////////////////////////

const describeInjectorTemplate = `Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}

DAG:
{{.DAG}}
`
const describeInjectorScopeTemplate = `{{.ScopeName}} (ID: {{.ScopeID}}){{.Services}}{{.Children}}`
const describeInjectorServiceTemplate = ` * {{.ServiceType}}{{.ServiceName}}{{.ServiceFeatures}}`

type InjectorDescription struct {
	ScopeID   string                     `json:"scope_id"`
	ScopeName string                     `json:"scope_name"`
	DAG       []DescriptionInjectorScope `json:"dag"`
}

func (id *InjectorDescription) String() string {
	dag := mergeScopes(&id.DAG)
	if strings.HasPrefix(dag, " |\n |\n") {
		dag = dag[3:]
	}

	return fromTemplate(
		describeInjectorTemplate,
		map[string]string{
			"ScopeID":   id.ScopeID,
			"ScopeName": id.ScopeName,
			"DAG":       dag,
		},
	)
}

func mergeScopes(scopes *[]DescriptionInjectorScope) string {
	nbrScopes := len(*scopes)

	const prefixScope = ` |`
	const prefixLastScopeHeader = `  \_ `
	const prefixLastScopeContent = `     `
	const prefixNotLastScopeHeader = ` |\_ `
	const prefixNotLastScopeContent = ` |   `

	return strings.Join(
		mAp(*scopes, func(item DescriptionInjectorScope, i int) string {
			isLastScope := i == nbrScopes-1

			lines := strings.Split(item.String(), "\n")
			lines = mAp(lines, func(line string, j int) string {
				if isLastScope && j == 0 {
					return prefixLastScopeHeader + line
				} else if isLastScope {
					return prefixLastScopeContent + line
				} else if j == 0 {
					return prefixNotLastScopeHeader + line
				} else {
					return prefixNotLastScopeContent + line
				}
			})

			lines = append([]string{prefixScope, prefixScope}, lines...)
			return strings.Join(lines, "\n")
		}),
		"\n",
	)
}

type DescriptionInjectorScope struct {
	ScopeID   string                       `json:"scope_id"`
	ScopeName string                       `json:"scope_name"`
	Services  []DescriptionInjectorService `json:"services"`
	Children  []DescriptionInjectorScope   `json:"children"`

	IsAncestor bool `json:"is_ancestor"`
	IsChildren bool `json:"is_children"`
}

func (ids *DescriptionInjectorScope) String() string {
	services := strings.Join(
		mAp(ids.Services, func(item DescriptionInjectorService, _ int) string {
			return item.String()
		}),
		"\n",
	)
	if len(ids.Services) > 0 {
		services = "\n" + services
	}

	children := mergeScopes(&ids.Children)
	if len(ids.Children) > 0 {
		children = "\n" + children
	}

	return fromTemplate(
		describeInjectorScopeTemplate,
		map[string]string{
			"ScopeID":   ids.ScopeID,
			"ScopeName": ids.ScopeName,
			"Services":  services,
			"Children":  children,
		},
	)
}

type DescriptionInjectorService struct {
	ServiceName     string      `json:"service_name"`
	ServiceType     ServiceType `json:"service_type"`
	IsHealthchecker bool        `json:"is_healthchecker"`
	IsShutdowner    bool        `json:"is_shutdowner"`
}

func (idss *DescriptionInjectorService) String() string {
	prefix := ""
	suffix := ""

	if idss.ServiceType != empty[ServiceType]() {
		// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
		prefix += serviceTypeToIcon[idss.ServiceType] + " "

		if idss.IsHealthchecker {
			suffix += " 🏥"
		}
		if idss.IsShutdowner {
			suffix += " 🙅"
		}
	} else {
		prefix += "❓ " // should never reach this branch
	}

	return fromTemplate(
		describeInjectorServiceTemplate,
		map[string]string{
			"ServiceName":     idss.ServiceName,
			"ServiceType":     prefix,
			"ServiceFeatures": suffix,
		},
	)
}

// DescribeInjector returns a human readable description of the injector, with services and scope tree.
func DescribeInjector(scope Injector) InjectorDescription {
	_i := getInjectorOrDefault(scope)

	ancestors := append([]Injector{_i}, castScopesToInjectors(_i.Ancestors())...)
	reverseSlice(ancestors) // root scope first

	return InjectorDescription{
		ScopeID:   _i.ID(),
		ScopeName: _i.Name(),
		DAG:       newDescriptionInjectorScopes(ancestors, castScopesToInjectors(_i.Children())),
	}
}

// 2 modes are available: looping on ancestors, focused-scope or children
func newDescriptionInjectorScopes(ancestors []Injector, children []Injector) []DescriptionInjectorScope {
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

	return mAp(injectors, func(item Injector, _ int) DescriptionInjectorScope {
		nextChildren := children
		if loopingOn == "children" {
			nextChildren = castScopesToInjectors(item.Children())
		}

		return DescriptionInjectorScope{
			ScopeID:   item.ID(),
			ScopeName: item.Name(),
			Services:  newDescriptionInjectorServices(item),
			Children:  newDescriptionInjectorScopes(ancestors, nextChildren),

			IsAncestor: loopingOn == "ancestors",
			IsChildren: loopingOn == "children",
		}
	})
}

func newDescriptionInjectorServices(i Injector) []DescriptionInjectorService {
	services := i.ListProvidedServices()
	services = filter(services, func(item EdgeService, _ int) bool {
		return i.serviceExist(item.Service)
	})

	// order by id to have a deterministic output in unit tests
	sort.Slice(services, func(i, j int) bool {
		return services[i].Service < services[j].Service
	})

	return mAp(services, func(item EdgeService, _ int) DescriptionInjectorService {
		var serviceType ServiceType
		var isHealthchecker bool
		var isShutdowner bool

		if info, ok := inferServiceInfo(i, item.Service); ok {
			// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
			serviceType = info.serviceType
			isHealthchecker = info.healthchecker
			isShutdowner = info.shutdowner
		}

		return DescriptionInjectorService{
			ServiceName:     item.Service,
			ServiceType:     serviceType,
			IsHealthchecker: isHealthchecker,
			IsShutdowner:    isShutdowner,
		}
	})
}

func castScopesToInjectors(scopes []*Scope) []Injector {
	return mAp(scopes, func(item *Scope, _ int) Injector {
		return item
	})
}
