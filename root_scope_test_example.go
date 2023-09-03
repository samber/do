package do

// import "fmt"

// func ExampleNew() {
// 	injector := New()

// 	ProvideNamedValue(injector, "PG_URI", "postgres://user:pass@host:5432/db")
// 	uri, err := InvokeNamed[string](injector, "PG_URI")

// 	fmt.Println(uri)
// 	fmt.Println(err)
// 	// Output:
// 	// postgres://user:pass@host:5432/db
// 	// <nil>
// }

// func ExampleDefaultRootScope() {
// 	ProvideNamedValue(nil, "PG_URI", "postgres://user:pass@host:5432/db")
// 	uri, err := InvokeNamed[string](nil, "PG_URI")

// 	fmt.Println(uri)
// 	fmt.Println(err)
// 	// Output:
// 	// postgres://user:pass@host:5432/db
// 	// <nil>
// }

// func ExampleNewWithOpts() {
// 	injector := NewWithOpts(&InjectorOpts{
// 		HookAfterShutdown: func(scope *Scope, serviceName string) {
// 			fmt.Printf("service shutdown: %s\n", serviceName)
// 		},
// 	})

// 	ProvideNamed(injector, "PG_URI", func(i Injector) (string, error) {
// 		return "postgres://user:pass@host:5432/db", nil
// 	})
// 	MustInvokeNamed[string](injector, "PG_URI")
// 	err := injector.Shutdown()
// 	fmt.Println(err)

// 	// Output:
// 	// service shutdown: PG_URI
// 	// <nil>
// }
