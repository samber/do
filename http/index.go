package http

func IndexHTML(basePath string) (string, error) {
	return fromTemplate(
		`<!DOCTYPE html>
<html>
	<head>
		<title>Dependency injection UI - samber/do</title>
	</head>
	<body>
		<h1>Welcome to do UI ✌️</h1>
		
		<h2>Introduction</h2>
		<ul>
			<li><a href="https://github.com/samber/do" target="_blank">Repository</a></li>
			<li><a href="https://github.com/samber/do/issues" target="_blank">New issue</a></li>
			<li><a href="https://do.samber.dev" target="_blank">Documentation</a></li>
			<li><a href="https://pkg.go.dev/github.com/samber/do/v2" target="_blank">Godoc</a></li>
			<li><a href="https://github.com/samber/do/releases" target="_blank">Changelog</a></li>
			<li><a href="https://github.com/samber/do/blob/master/LICENSE" target="_blank">License</a></li>
		</ul>

		<h2>Getting started</h2>
		<ul>
			<li><a href="{{.BasePath}}/scope">Inspect scopes</a></li>
			<li><a href="{{.BasePath}}/service">Inspect services</a></li>
		</ul>
	</body>
</html>`,
		map[string]any{
			"BasePath": basePath,
		},
	)
}
