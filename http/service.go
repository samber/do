package http

import (
	"github.com/samber/do/v2"
)

func ServiceHTML(basePath string, injector do.Injector, scopeID string, serviceName string) (string, error) {
	scope, ok := getScopeByID(injector, scopeID)
	if !ok {
		return ServiceListHTML(basePath, injector)
	}

	service, ok := do.DescribeNamedService(scope.Scope, serviceName)
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

func serviceToHTML(basePath string, services []do.DescriptionServiceDependency) string {
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
			"Services": mAp(services, func(service do.DescriptionServiceDependency) map[string]any {
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
			"Scopes": mAp(scopes, func(item do.DescriptionInjectorScope) string {
				return serviceListScopeToHTML(basePath, item)
			}),
		},
	)
}

func serviceListScopeToHTML(basePath string, description do.DescriptionInjectorScope) string {
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
			"Services": mAp(description.Services, func(item do.DescriptionInjectorService) string {
				return serviceListServiceToHTML(basePath, description.ScopeID, item)
			}),
		},
	)
	return html
}

func serviceListServiceToHTML(basePath string, scopeID string, description do.DescriptionInjectorService) string {
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
