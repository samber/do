package do

import (
	"database/sql"
	"fmt"
)

type dbService struct {
	db *sql.DB
}

func (s *dbService) HealthCheck() error {
	return nil
}

func (s *dbService) Shutdown() error {
	return nil
}

func dbServiceProvider(i Injector) (*dbService, error) {
	return &dbService{db: nil}, nil
}

func ExampleInjector_ListProvidedServices() {
	injector := New()

	ProvideNamedValue(injector, "PG_URI", "postgres://user:pass@host:5432/db")
	services := injector.ListProvidedServices()

	fmt.Println(services)
	// Output:
	// [PG_URI]
}

func ExampleInjector_ListInvokedServices_invoked() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "SERVICE_NAME", func(i Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})
	_, _ = InvokeNamed[test](injector, "SERVICE_NAME")
	services := injector.ListInvokedServices()

	fmt.Println(services)
	// Output:
	// [SERVICE_NAME]
}

func ExampleInjector_ListInvokedServices_notInvoked() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "SERVICE_NAME", func(i Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})
	services := injector.ListInvokedServices()

	fmt.Println(services)
	// Output:
	// []
}

func ExampleInjector_HealthCheck() {
	injector := New()

	Provide(injector, dbServiceProvider)
	health := injector.HealthCheck()

	fmt.Println(health)
	// Output:
	// map[*do.dbService:<nil>]
}

func ExampleInjector_Shutdown() {
	injector := New()

	Provide(injector, dbServiceProvider)
	err := injector.Shutdown()

	fmt.Println(err)
	// Output:
	// <nil>
}

func ExampleInjector_Clone() {
	injector := New()

	ProvideNamedValue(injector, "PG_URI", "postgres://user:pass@host:5432/db")
	injector2 := injector.Clone()
	services := injector2.ListProvidedServices()

	fmt.Println(services)
	// Output:
	// [PG_URI]
}
