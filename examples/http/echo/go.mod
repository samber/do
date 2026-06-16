module github.com/samber/do/v2/examples/http/echo

go 1.25.0

replace github.com/samber/do/v2 => ../../../

replace github.com/samber/do/http/echo/v2 => ../../../http/echo

require (
	github.com/labstack/echo/v4 v4.15.4
	github.com/samber/do/http/echo/v2 v2.0.0-00010101000000-000000000000
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
)

require (
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/samber/go-type-to-string v1.8.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)
