module github.com/samber/do/v2/examples/http/chi

go 1.18

replace github.com/samber/do/v2 => ../../../

replace github.com/samber/do/v2/http/chi => ../../../http/chi

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
	github.com/samber/do/v2/http/chi v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-chi/chi v1.5.5 // indirect
	github.com/samber/go-type-to-string v1.1.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
)
