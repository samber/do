package do

import (
	"fmt"
)

func ExampleNameOf() {
	type exampleService struct {
		Name string
	}

	name := NameOf[*exampleService]()

	fmt.Println(name)
	// Output: *github.com/samber/do/v2.exampleService
}

func ExampleProvide() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := Invoke[*exampleService](injector)

	fmt.Println(err)
	fmt.Println(service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleProvideNamed() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := InvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(err)
	fmt.Println(service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleProvideValue() {
	type exampleService struct {
		Name string
	}

	injector := New()

	service := &exampleService{Name: "eager-service"}
	ProvideValue(injector, service)
	retrieved, err := Invoke[*exampleService](injector)

	fmt.Println(err)
	fmt.Println(retrieved.Name)
	// Output:
	// <nil>
	// eager-service
}

func ExampleProvideNamedValue() {
	type exampleService struct {
		Name string
	}

	injector := New()

	service := &exampleService{Name: "named-eager-service"}
	ProvideNamedValue(injector, "my-eager-service", service)
	retrieved, err := InvokeNamed[*exampleService](injector, "my-eager-service")

	fmt.Println(err)
	fmt.Println(retrieved.Name)
	// Output:
	// <nil>
	// named-eager-service
}

func ExampleProvideTransient() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideTransient(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service1, _ := Invoke[*exampleService](injector)
	service2, _ := Invoke[*exampleService](injector)

	fmt.Println(service1 != service2)
	// Output: true
}

func ExampleProvideNamedTransient() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamedTransient(injector, "transient-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service1, _ := InvokeNamed[*exampleService](injector, "transient-service")
	service2, _ := InvokeNamed[*exampleService](injector, "transient-service")

	fmt.Println(service1 != service2)
	// Output: true
}

func ExampleOverride() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	Override(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "overridden-service"}, nil
	})
	service, _ := Invoke[*exampleService](injector)

	fmt.Println(service.Name)
	// Output: overridden-service
}

func ExampleOverrideNamed() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	OverrideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "overridden-named-service"}, nil
	})
	service, _ := InvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(service.Name)
	// Output: overridden-named-service
}

func ExampleOverrideValue() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	OverrideValue(injector, &exampleService{Name: "overridden-value-service"})
	service, _ := Invoke[*exampleService](injector)

	fmt.Println(service.Name)
	// Output: overridden-value-service
}

func ExampleOverrideNamedValue() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	OverrideNamedValue(injector, "my-service", &exampleService{Name: "overridden-named-value-service"})
	service, _ := InvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(service.Name)
	// Output: overridden-named-value-service
}

func ExampleOverrideTransient() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	OverrideTransient(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "overridden-service"}, nil
	})
	service1, _ := Invoke[*exampleService](injector)
	service2, _ := Invoke[*exampleService](injector)

	fmt.Println(service1 != service2)
	// Output: true
}

func ExampleOverrideNamedTransient() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	OverrideNamedTransient(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "overridden-service"}, nil
	})
	service1, _ := InvokeNamed[*exampleService](injector, "my-service")
	service2, _ := InvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(service1 != service2)
	// Output: true
}

func ExampleInvoke() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := Invoke[*exampleService](injector)

	fmt.Println(err)
	fmt.Println(service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleInvokeNamed() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := InvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(err)
	fmt.Println(service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleMustInvoke() {
	type exampleService struct {
		Name string
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service := MustInvoke[*exampleService](injector)

	fmt.Println(service.Name)
	// Output: test-service
}

func ExampleMustInvokeNamed() {
	type exampleService struct {
		Name string
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service := MustInvokeNamed[*exampleService](injector, "my-service")

	fmt.Println(service.Name)
	// Output: test-service
}

func ExampleInvokeStruct() {
	type exampleService struct {
		Name string
	}
	type exampleStructService struct {
		Service *exampleService `do:""`
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := InvokeStruct[exampleStructService](injector)

	fmt.Println(err)
	fmt.Println(service.Service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleInvokeStruct_named() {
	type exampleService struct {
		Name string
	}
	type exampleStructStruct struct {
		Service *exampleService `do:"my-service"`
	}

	injector := New()

	ProvideNamed(injector, "my-service", func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service, err := InvokeStruct[exampleStructStruct](injector)

	fmt.Println(err)
	fmt.Println(service.Service.Name)
	// Output:
	// <nil>
	// test-service
}

func ExampleMustInvokeStruct() {
	type exampleService struct {
		Name string
	}
	type exampleStructService struct {
		Service *exampleService `do:""`
	}

	injector := New()

	Provide(injector, func(i Injector) (*exampleService, error) {
		return &exampleService{Name: "test-service"}, nil
	})
	service := MustInvokeStruct[exampleStructService](injector)

	fmt.Println(service.Service.Name)
	// Output: test-service
}

type IMyService interface {
	GetName() string
}

type myService struct {
	Name string
}

func (s *myService) GetName() string {
	return s.Name
}

func (s *myService) HealthCheck() error {
	return nil
}

func (s *myService) Shutdown() error {
	return nil
}

func aliasMyServiceProvider(i Injector) (*myService, error) {
	return &myService{Name: "alias-test-service"}, nil
}

func ExampleAs() {
	injector := New()

	Provide(injector, aliasMyServiceProvider)
	err := As[*myService, IMyService](injector)

	fmt.Println(err)

	// Now both work:
	service1, _ := Invoke[*myService](injector)
	service2, _ := Invoke[IMyService](injector)

	fmt.Println(service1.GetName())
	fmt.Println(service2.GetName())
	// Output:
	// <nil>
	// alias-test-service
	// alias-test-service
}

func ExampleMustAs() {
	injector := New()

	Provide(injector, aliasMyServiceProvider)
	MustAs[*myService, IMyService](injector)

	// Now both work:
	service1 := MustInvoke[*myService](injector)
	service2 := MustInvoke[IMyService](injector)

	fmt.Println(service1.GetName())
	fmt.Println(service2.GetName())
	// Output:
	// alias-test-service
	// alias-test-service
}

func ExampleAsNamed() {
	injector := New()

	ProvideNamed(injector, "my-service", aliasMyServiceProvider)
	err := AsNamed[*myService, IMyService](injector, "my-service", "my-interface")

	fmt.Println(err)

	// Retrieve using the alias name
	service, _ := InvokeNamed[IMyService](injector, "my-interface")
	fmt.Println(service.GetName())
	// Output:
	// <nil>
	// alias-test-service
}

func ExampleMustAsNamed() {
	injector := New()

	ProvideNamed(injector, "my-service", aliasMyServiceProvider)
	MustAsNamed[*myService, IMyService](injector, "my-service", "my-interface")

	// Retrieve using the alias name
	service := MustInvokeNamed[IMyService](injector, "my-interface")
	fmt.Println(service.GetName())
	// Output: alias-test-service
}

func ExampleInvokeAs() {
	injector := New()

	Provide(injector, aliasMyServiceProvider)
	service, err := InvokeAs[IMyService](injector)

	fmt.Println(err)
	fmt.Println(service.GetName())
	// Output:
	// <nil>
	// alias-test-service
}

func ExampleMustInvokeAs() {
	injector := New()

	Provide(injector, aliasMyServiceProvider)
	service := MustInvokeAs[IMyService](injector)

	fmt.Println(service.GetName())
	// Output: alias-test-service
}
