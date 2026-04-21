package dohttp

import (
	"github.com/samber/do/v2"
)

// ServiceHTML generates an HTML page that displays detailed information about a specific service.
// This function creates a comprehensive service inspection page showing the service's scope,
// type, build time, invocation location, dependencies, and dependents.
//
// Parameters:
//   - basePath: The base URL path for the web interface
//   - injector: The injector containing the service
//   - scopeID: The ID of the scope containing the service
//   - serviceName: The name of the service to inspect
//
// Returns the HTML content as a string and any error that occurred during generation.
//
// The generated page includes:
//   - Service metadata (scope, type, build time, invocation location)
//   - List of dependencies with clickable links
//   - List of dependents with clickable links
//   - Navigation to other views
//
// If the scope or service is not found, it falls back to the service list page.
//
// Example:
//
//	html, err := http.ServiceHTML("/debug/di", injector, "root", "database")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Fprint(w, html)
func ServiceHTML(basePath string, injector do.Injector, scopeID string, serviceName string) (string, error) {
	scope, ok := getScopeByID(injector, scopeID)
	if !ok {
		return ServiceListHTML(basePath, injector)
	}

	service, ok := do.ExplainNamedService(scope.Scope, serviceName)
	if !ok {
		return ServiceListHTML(basePath, injector)
	}

	invoked := ""
	if service.Invoked != nil {
		invoked = service.Invoked.String()
	}

	return fromTemplate(
		`<!DOCTYPE html>
<html>
<head>
	<title>Inspect service - samber/do</title>
	<style>
	</style>
</head>
<body>
	<h1>Services by scope</h1>
	<small>
		Menu:
		<a href="{{.BasePath}}">Home</a>
		-
		<a href="{{$.BasePath}}/scope">Scopes</a>
		-
		<a href="{{$.BasePath}}/service">Services</a>
	</small>

	<p>
		Scope id: {{.ScopeID}}
		<br>
		Scope name: {{.ScopeName}}
		<br>
		Service name: {{.ServiceName}}
		<br>
		Service type: {{.ServiceType}}
		{{if .ServiceBuildTime}}
			<br>
			Service build time: {{.ServiceBuildTime}}
		{{end}}
		<br>
		Invoked at: {{.Invoked}}
	</p>

	<h2>Dependencies:</h2>
	{{.Dependencies}}

	<h2>Dependents:</h2>
	{{.Dependents}}
</body>
</html>`,
		map[string]any{
			"BasePath":         basePath,
			"ScopeID":          service.ScopeID,
			"ScopeName":        service.ScopeName,
			"ServiceName":      service.ServiceName,
			"ServiceType":      service.ServiceType,
			"ServiceBuildTime": service.ServiceBuildTime,
			"Invoked":          invoked,
			"Dependencies":     serviceToHTML(basePath, service.Dependencies),
			"Dependents":       serviceToHTML(basePath, service.Dependents),
		},
	)
}

// serviceToHTML converts a list of service dependencies to HTML representation.
// This function generates clickable links for each service in the dependency list,
// allowing users to navigate to detailed views of related services.
//
// Parameters:
//   - basePath: The base URL path for the web interface
//   - services: List of service dependency outputs to convert
//
// Returns the HTML string representation of the service list.
//
// Each service is rendered as a clickable link that navigates to the service
// detail page, with recursive dependencies shown as nested lists.
func serviceToHTML(basePath string, services []do.ExplainServiceDependencyOutput) string {
	output, _ := fromTemplate(
		`
		<ul class="services">
			{{range .Services}}
				<li class="service">
					<a href="{{$.BasePath}}/service?scope_id={{.ScopeID}}&service_name={{.ServiceName}}">
						{{.ServiceName}}
					</a>
					{{if .Recursive}}
						{{.Recursive}}
					{{end}}
				</li>
			{{end}}
		</ul>
	`,
		map[string]any{
			"BasePath": basePath,
			"Services": mAp(services, func(service do.ExplainServiceDependencyOutput) map[string]any {
				return map[string]any{
					"ScopeID":     service.ScopeID,
					"ServiceName": service.Service,
					"Recursive":   serviceToHTML(basePath, service.Recursive),
				}
			}),
		},
	)
	return output
}

// ServiceListHTML generates an HTML page that displays a list of all services across all scopes.
// This function creates a comprehensive service listing page showing all services
// organized by their respective scopes.
//
// Parameters:
//   - basePath: The base URL path for the web interface
//   - injector: The injector containing the services to list
//
// Returns the HTML content as a string and any error that occurred during generation.
//
// The generated page includes:
//   - List of all scopes in the injector hierarchy
//   - Services within each scope with clickable links
//   - Navigation to other views
//   - Service type indicators and capabilities
//
// Example:
//
//	html, err := http.ServiceListHTML("/debug/di", injector)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Fprint(w, html)
func ServiceListHTML(basePath string, injector do.Injector) (string, error) {
	scopes := getAllScopes(injector)

	return fromTemplate(
		`<!DOCTYPE html>
<html>
	<head>
		<title>Inspect service - samber/do</title>
		<style>
		.scope {
			padding-top: 10px;
			padding-bottom: 10px;
		}
		</style>
	</head>
	<body>
		<h1>Service description</h1>
		<small>
			Menu:
			<a href="{{.BasePath}}">Home</a>
			-
			<a href="{{$.BasePath}}/scope">Scopes</a>
			-
			<a href="{{$.BasePath}}/service">Services</a>
		</small>

		<ul class="scopes">
			{{range .Scopes}}
				<li class="scope">
					{{.}}
				</li>
			{{end}}
		</ul>
	</body>
</html>`,
		map[string]any{
			"BasePath": basePath,
			"Scopes": mAp(scopes, func(item do.ExplainInjectorScopeOutput) string {
				return serviceListScopeToHTML(basePath, item)
			}),
		},
	)
}

func serviceListScopeToHTML(basePath string, description do.ExplainInjectorScopeOutput) string {
	html, _ := fromTemplate(
		`
			Scope:
			<a href="{{$.BasePath}}/scope?scope_id={{.ScopeID}}">
				{{.ScopeName}}
			</a>

			{{if .Services}}
				<ul class="services">
					{{range .Services}}
						<li class="service">
							{{.}}
						</li>
					{{end}}
				</ul>
			{{end}}
		`,
		map[string]any{
			"BasePath":  basePath,
			"ScopeID":   description.ScopeID,
			"ScopeName": description.ScopeName,
			"Services": mAp(description.Services, func(item do.ExplainInjectorServiceOutput) string {
				return serviceListServiceToHTML(basePath, description.ScopeID, item)
			}),
		},
	)
	return html
}

func serviceListServiceToHTML(basePath string, scopeID string, description do.ExplainInjectorServiceOutput) string {
	featuresIcons := ""

	if description.IsHealthchecker {
		featuresIcons += " 🫀"
	}

	if description.IsShutdowner {
		featuresIcons += " 🙅"
	}

	html, _ := fromTemplate(
		`
			{{.ServiceTypeIcon}}
			<a href="{{$.BasePath}}/service?scope_id={{.ScopeID}}&service_name={{.ServiceName}}">
				{{.ServiceName}}
			</a>
			{{.FeaturesIcons}}
		`,
		map[string]any{
			"BasePath":        basePath,
			"ScopeID":         scopeID,
			"ServiceName":     description.ServiceName,
			"ServiceTypeIcon": description.ServiceTypeIcon,
			"FeaturesIcons":   featuresIcons,
		},
	)
	return html
}
