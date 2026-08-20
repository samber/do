module github.com/samber/do/v2

go 1.18

//
// Dependencies are excluded from releases. Please check CI.
//

require (
	github.com/samber/go-type-to-string v1.8.0
	github.com/stretchr/testify v1.12.1
	go.uber.org/goleak v1.2.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect
