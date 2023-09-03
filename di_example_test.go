package do

import (
	"fmt"
)

func ExampleProvide() {
	injector := New()

	type test struct {
		foobar string
	}

	Provide(injector, func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value, err := Invoke[*test](injector)

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleProvideNamed() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "my_service", func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value, err := InvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleProvideValue() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideValue(injector, &test{foobar: "foobar"})
	value, err := Invoke[*test](injector)

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleProvideNamedValue() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamedValue(injector, "my_service", &test{foobar: "foobar"})
	value, err := InvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleOverride() {
	injector := New()

	type test struct {
		foobar string
	}

	Provide(injector, func(i Injector) (*test, error) {
		return &test{foobar: "foobar1"}, nil
	})
	Override(injector, func(i Injector) (*test, error) {
		return &test{foobar: "foobar2"}, nil
	})
	value, err := Invoke[*test](injector)

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar2}
	// <nil>
}

func ExampleOverrideNamed() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "my_service", func(i Injector) (*test, error) {
		return &test{foobar: "foobar1"}, nil
	})
	OverrideNamed(injector, "my_service", func(i Injector) (*test, error) {
		return &test{foobar: "foobar2"}, nil
	})
	value, err := InvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar2}
	// <nil>
}

func ExampleOverrideNamedValue() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamedValue(injector, "my_service", &test{foobar: "foobar1"})
	OverrideNamedValue(injector, "my_service", &test{foobar: "foobar2"})
	value, err := InvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar2}
	// <nil>
}

func ExampleInvoke() {
	injector := New()

	type test struct {
		foobar string
	}

	Provide(injector, func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value, err := Invoke[*test](injector)

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleMustInvoke() {
	injector := New()

	type test struct {
		foobar string
	}

	Provide(injector, func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value := MustInvoke[*test](injector)

	fmt.Println(value)
	// Output:
	// &{foobar}
}

func ExampleInvokeNamed() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "my_service", func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value, err := InvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	fmt.Println(err)
	// Output:
	// &{foobar}
	// <nil>
}

func ExampleMustInvokeNamed() {
	injector := New()

	type test struct {
		foobar string
	}

	ProvideNamed(injector, "my_service", func(i Injector) (*test, error) {
		return &test{foobar: "foobar"}, nil
	})
	value := MustInvokeNamed[*test](injector, "my_service")

	fmt.Println(value)
	// Output:
	// &{foobar}
}
