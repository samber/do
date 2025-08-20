package do

import (
	"fmt"
)

type myService struct {
	Name string
}

type IMyService interface {
	GetName() string
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
