module github.com/samber/do/v2/examples/http/std

go 1.18

replace github.com/samber/do/v2 => ../../../

replace github.com/samber/do/v2/http/std => ../../../http/std

require (
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
	github.com/samber/do/v2/http/std v0.0.0-00010101000000-000000000000
)

require github.com/samber/go-type-to-string v1.1.0 // indirect
