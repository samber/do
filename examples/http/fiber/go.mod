module github.com/samber/do/v2/examples/http/fiber

go 1.18

replace github.com/samber/do/v2 => ../../../

replace github.com/samber/do/http/fiber/v2 => ../../../http/fiber

require (
	github.com/gofiber/fiber/v2 v2.52.11
	github.com/samber/do/http/fiber/v2 v2.0.0-00010101000000-000000000000
	github.com/samber/do/v2 v2.0.0-00010101000000-000000000000
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/samber/go-type-to-string v1.8.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
