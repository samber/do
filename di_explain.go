package do

import (
	"bytes"
	"html/template"
	"sort"
	"strings"
	"time"

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
// 							Explain services
/////////////////////////////////////////////////////////////////////////////

const explainServiceTemplate = `
Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}

Service name: {{.ServiceName}}
Service type: {{.ServiceType}}{{if .ServiceBuildTime}}
Service build time: {{.ServiceBuildTime}}{{end}}
Invoked: {{.Invoked}}

Dependencies:
{{.Dependencies}}

Dependents:
{{.Dependents}}
`

// @TODO: add service type icon (lazy, eager, transient)
const explainServiceDependencyTemplate = `* {{.Service}} from scope {{.ScopeName}}{{.Recursive}}`

type ExplainServiceOutput struct {
	ScopeID          string                           `json:"scope_id"`
	ScopeName        string                           `json:"scope_name"`
	ServiceName      string                           `json:"service_name"`
	ServiceType      ServiceType                      `json:"service_type"`
	ServiceBuildTime time.Duration                    `json:"service_build_time,omitempty"`
	Invoked          *stacktrace.Frame                `json:"invoked"`
	Dependencies     []ExplainServiceDependencyOutput `json:"dependencies"`
	Dependents       []ExplainServiceDependencyOutput `json:"dependents"`
}

func (sd *ExplainServiceOutput) String() string {
	invoked := ""
	if sd.Invoked != nil {
		invoked = sd.Invoked.String()
	}

	buildTime := ""
	if sd.ServiceBuildTime > 0 {
		buildTime = sd.ServiceBuildTime.String()
	}

	return fromTemplate(
		explainServiceTemplate,
		map[string]string{
			"ScopeID":          sd.ScopeID,
			"ScopeName":        sd.ScopeName,
			"ServiceName":      sd.ServiceName,
			"ServiceType":      string(sd.ServiceType),
			"ServiceBuildTime": buildTime,
			"Invoked":          invoked,
			"Dependencies": strings.Join(
				mAp(sd.Dependencies, func(item ExplainServiceDependencyOutput, _ int) string {
					return item.String()
				}),
				"\n",
			),
			"Dependents": strings.Join(
				mAp(sd.Dependents, func(item ExplainServiceDependencyOutput, _ int) string {
					return item.String()
				}),
				"\n",
			),
		},
	)
}

type ExplainServiceDependencyOutput struct {
	ScopeID   string                           `json:"scope_id"`
	ScopeName string                           `json:"scope_name"`
	Service   string                           `json:"service"`
	Recursive []ExplainServiceDependencyOutput `json:"recursive"`
}

func (sdd *ExplainServiceDependencyOutput) String() string {
	lines := flatten(
		mAp(sdd.Recursive, func(item ExplainServiceDependencyOutput, _ int) []string {
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
		explainServiceDependencyTemplate,
		map[string]string{
			"ScopeID":   sdd.ScopeID,
			"ScopeName": sdd.ScopeName,
			"Service":   sdd.Service,
			"Recursive": recursive,
		},
	)
}

// ExplainService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before ExplainService[T] to ensure that the service is registered.
func ExplainService[T any](i Injector) (description ExplainServiceOutput, ok bool) {
	name := inferServiceName[T]()
	return ExplainNamedService(i, name)
}

// ExplainNamedService returns a human readable description of the service.
// It returns false if the service is not found.
// Please call Invoke[T] before ExplainNamedService[T] to ensure that the service is registered.
func ExplainNamedService(scope Injector, name string) (description ExplainServiceOutput, ok bool) {
	_i := getInjectorOrDefault(scope)

	serviceAny, serviceScope, ok := _i.serviceGetRec(name)
	if !ok {
		return ExplainServiceOutput{}, false
	}

	service, ok := serviceAny.(ServiceAny)
	if !ok {
		return ExplainServiceOutput{}, false
	}

	var invoked *stacktrace.Frame
	frame, ok := inferServiceProviderStacktrace(service)
	if ok {
		invoked = &frame
	}

	var buildTime time.Duration
	if lazy, ok := serviceAny.(serviceBuildTime); ok {
		buildTime, _ = lazy.getBuildTime()
	}

	return ExplainServiceOutput{
		ScopeID:          serviceScope.ID(),
		ScopeName:        serviceScope.Name(),
		ServiceName:      name,
		ServiceType:      service.getType(),
		ServiceBuildTime: buildTime,
		Invoked:          invoked,
		Dependencies:     newExplainServiceDependencies(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependencies"),
		Dependents:       newExplainServiceDependencies(_i, newEdgeService(_i.ID(), _i.Name(), name), "dependents"),
	}, true
}

func newExplainServiceDependencies(i Injector, edge EdgeService, mode string) []ExplainServiceDependencyOutput {
	dependencies, dependents := i.RootScope().dag.explainService(edge.ScopeID, edge.ScopeName, edge.Service)

	deps := dependencies
	if mode == "dependents" {
		deps = dependents
	}

	// order by id to have a deterministic output in unit tests
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Service < deps[j].Service
	})

	return mAp(deps, func(item EdgeService, _ int) ExplainServiceDependencyOutput {
		recursive := newExplainServiceDependencies(i, item, mode)

		// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
		return ExplainServiceDependencyOutput{
			ScopeID:   item.ScopeID,
			ScopeName: item.ScopeName,
			Service:   item.Service,
			Recursive: recursive,
		}
	})
}

/////////////////////////////////////////////////////////////////////////////
// 							Explain scopes
/////////////////////////////////////////////////////////////////////////////

const explainInjectorTemplate = `Scope ID: {{.ScopeID}}
Scope name: {{.ScopeName}}

DAG:
{{.DAG}}
`
const explainInjectorScopeTemplate = `{{.ScopeName}} (ID: {{.ScopeID}}){{.Services}}{{.Children}}`
const explainInjectorServiceTemplate = ` * {{.ServiceType}}{{.ServiceName}}{{.ServiceFeatures}}`

type ExplainInjectorOutput struct {
	ScopeID   string                       `json:"scope_id"`
	ScopeName string                       `json:"scope_name"`
	DAG       []ExplainInjectorScopeOutput `json:"dag"`
}

func (id *ExplainInjectorOutput) String() string {
	dag := mergeScopes(&id.DAG)
	if strings.HasPrefix(dag, " |\n |\n") {
		dag = dag[3:]
	}

	return fromTemplate(
		explainInjectorTemplate,
		map[string]string{
			"ScopeID":   id.ScopeID,
			"ScopeName": id.ScopeName,
			"DAG":       dag,
		},
	)
}

func mergeScopes(scopes *[]ExplainInjectorScopeOutput) string {
	nbrScopes := len(*scopes)

	const prefixScope = ` |`
	const prefixLastScopeHeader = `  \_ `
	const prefixLastScopeContent = `     `
	const prefixNotLastScopeHeader = ` |\_ `
	const prefixNotLastScopeContent = ` |   `

	return strings.Join(
		mAp(*scopes, func(item ExplainInjectorScopeOutput, i int) string {
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

type ExplainInjectorScopeOutput struct {
	ScopeID   string                         `json:"scope_id"`
	ScopeName string                         `json:"scope_name"`
	Scope     Injector                       `json:"scope"`
	Services  []ExplainInjectorServiceOutput `json:"services"`
	Children  []ExplainInjectorScopeOutput   `json:"children"`

	IsAncestor bool `json:"is_ancestor"`
	IsChildren bool `json:"is_children"`
}

func (ids *ExplainInjectorScopeOutput) String() string {
	services := strings.Join(
		mAp(ids.Services, func(item ExplainInjectorServiceOutput, _ int) string {
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
		explainInjectorScopeTemplate,
		map[string]string{
			"ScopeID":   ids.ScopeID,
			"ScopeName": ids.ScopeName,
			"Services":  services,
			"Children":  children,
		},
	)
}

type ExplainInjectorServiceOutput struct {
	ServiceName      string        `json:"service_name"`
	ServiceType      ServiceType   `json:"service_type"`
	ServiceTypeIcon  string        `json:"service_type_icon"`
	ServiceBuildTime time.Duration `json:"service_build_time,omitempty"`
	IsHealthchecker  bool          `json:"is_healthchecker"`
	IsShutdowner     bool          `json:"is_shutdowner"`
}

func (idss *ExplainInjectorServiceOutput) String() string {
	prefix := ""
	suffix := ""

	if idss.ServiceType != empty[ServiceType]() {
		// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
		prefix += idss.ServiceTypeIcon + " "

		if idss.IsHealthchecker {
			suffix += " 🫀"
		}

		if idss.IsShutdowner {
			suffix += " 🙅"
		}

		// if idss.ServiceBuildTime > 0 {
		// 	suffix += fmt.Sprintf(" (build time: %s)", idss.ServiceBuildTime.String())
		// }
	} else {
		prefix += "❓ " // should never reach this branch
	}

	return fromTemplate(
		explainInjectorServiceTemplate,
		map[string]string{
			"ServiceName":     idss.ServiceName,
			"ServiceType":     prefix,
			"ServiceFeatures": suffix,
		},
	)
}

// ExplainInjector returns a human readable description of the injector, with services and scope tree.
func ExplainInjector(scope Injector) ExplainInjectorOutput {
	_i := getInjectorOrDefault(scope)

	ancestors := append([]Injector{_i}, castScopesToInjectors(_i.Ancestors())...)
	reverseSlice(ancestors) // root scope first

	return ExplainInjectorOutput{
		ScopeID:   _i.ID(),
		ScopeName: _i.Name(),
		DAG:       newExplainInjectorScopes(ancestors, castScopesToInjectors(_i.Children())),
	}
}

// 2 modes are available: looping on ancestors, focused-scope or children
func newExplainInjectorScopes(ancestors []Injector, children []Injector) []ExplainInjectorScopeOutput {
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

	return mAp(injectors, func(item Injector, _ int) ExplainInjectorScopeOutput {
		nextChildren := children
		if loopingOn == "children" {
			nextChildren = castScopesToInjectors(item.Children())
		}

		return ExplainInjectorScopeOutput{
			ScopeID:   item.ID(),
			ScopeName: item.Name(),
			Scope:     item,
			Services:  newExplainInjectorServices(item),
			Children:  newExplainInjectorScopes(ancestors, nextChildren),

			IsAncestor: loopingOn == "ancestors",
			IsChildren: loopingOn == "children",
		}
	})
}

func newExplainInjectorServices(i Injector) []ExplainInjectorServiceOutput {
	services := i.ListProvidedServices()
	services = filter(services, func(item EdgeService, _ int) bool {
		return i.serviceExist(item.Service)
	})

	// order by id to have a deterministic output in unit tests
	sort.Slice(services, func(i, j int) bool {
		return services[i].Service < services[j].Service
	})

	return mAp(services, func(item EdgeService, _ int) ExplainInjectorServiceOutput {
		var serviceType ServiceType
		var serviceTypeIcon string
		var serviceBuildTime time.Duration
		var isHealthchecker bool
		var isShutdowner bool

		if info, ok := inferServiceInfo(i, item.Service); ok {
			// @TODO: differenciate status of lazy services (built, not built). Such as: "😴 (✅)"
			serviceType = info.serviceType
			serviceTypeIcon = serviceTypeToIcon[info.serviceType]
			serviceBuildTime = info.serviceBuildTime
			isHealthchecker = info.healthchecker
			isShutdowner = info.shutdowner
		}

		return ExplainInjectorServiceOutput{
			ServiceName:      item.Service,
			ServiceType:      serviceType,
			ServiceTypeIcon:  serviceTypeIcon,
			ServiceBuildTime: serviceBuildTime,
			IsHealthchecker:  isHealthchecker,
			IsShutdowner:     isShutdowner,
		}
	})
}

func castScopesToInjectors(scopes []*Scope) []Injector {
	return mAp(scopes, func(item *Scope, _ int) Injector {
		return item
	})
}
