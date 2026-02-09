module github.com/samber/do/http/chi/v2

go 1.22

replace github.com/samber/do/v2 => ../../

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
)

require github.com/samber/go-type-to-string v1.8.0 // indirect
