module github.com/samber/do/http/chi/v2

go 1.18

replace github.com/samber/do/v2 => ../../

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
)

require github.com/samber/go-type-to-string v1.1.0 // indirect
