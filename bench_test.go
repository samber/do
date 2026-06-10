package do

import (
	"fmt"
	"testing"
)

//
// Benchmark fixtures
//

type benchSvcA struct{}

type benchSvcB struct{}

type benchSvcC struct{}

type benchSvcTop struct{}

type benchIface interface {
	benchMethod()
}

type benchSvcImpl struct{}

func (s *benchSvcImpl) benchMethod() {}

type benchHealthchecker struct{}

func (s *benchHealthchecker) HealthCheck() error { return nil }

type benchShutdowner struct{}

func (s *benchShutdowner) Shutdown() {}

//
// Invocation
//

func BenchmarkInvokeLazySteadyState(b *testing.B) {
	injector := New()
	Provide(injector, func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = MustInvoke[*benchSvcA](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Invoke[*benchSvcA](injector)
	}
}

func BenchmarkInvokeLazyParallel(b *testing.B) {
	injector := New()
	Provide(injector, func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = MustInvoke[*benchSvcA](injector)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = Invoke[*benchSvcA](injector)
		}
	})
}

func BenchmarkInvokeNamed(b *testing.B) {
	injector := New()
	ProvideNamed(injector, "bench-svc", func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = MustInvokeNamed[*benchSvcA](injector, "bench-svc")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = InvokeNamed[*benchSvcA](injector, "bench-svc")
	}
}

func BenchmarkInvokeTransient(b *testing.B) {
	injector := New()
	ProvideTransient(injector, func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Invoke[*benchSvcA](injector)
	}
}

func BenchmarkInvokeAlias(b *testing.B) {
	injector := New()
	Provide(injector, func(i Injector) (*benchSvcImpl, error) {
		return &benchSvcImpl{}, nil
	})
	if err := As[*benchSvcImpl, benchIface](injector); err != nil {
		b.Fatal(err)
	}
	_ = MustInvoke[benchIface](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Invoke[benchIface](injector)
	}
}

func BenchmarkInvokeAs(b *testing.B) {
	injector := New()
	for j := 0; j < 19; j++ {
		ProvideNamedValue(injector, fmt.Sprintf("bench-filler-%d", j), &benchSvcA{})
	}
	Provide(injector, func(i Injector) (*benchSvcImpl, error) {
		return &benchSvcImpl{}, nil
	})
	_ = MustInvokeAs[benchIface](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = InvokeAs[benchIface](injector)
	}
}

func BenchmarkInvokeNestedScopes(b *testing.B) {
	injector := New()
	Provide(injector, func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})

	scope := injector.Scope("level-1")
	for j := 2; j <= 5; j++ {
		scope = scope.Scope(fmt.Sprintf("level-%d", j))
	}
	_ = MustInvoke[*benchSvcA](scope)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Invoke[*benchSvcA](scope)
	}
}

func BenchmarkProviderChainTransient(b *testing.B) {
	injector := New()
	Provide(injector, func(i Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	Provide(injector, func(i Injector) (*benchSvcB, error) {
		return &benchSvcB{}, nil
	})
	Provide(injector, func(i Injector) (*benchSvcC, error) {
		return &benchSvcC{}, nil
	})
	ProvideTransient(injector, func(i Injector) (*benchSvcTop, error) {
		_ = MustInvoke[*benchSvcA](i)
		_ = MustInvoke[*benchSvcB](i)
		_ = MustInvoke[*benchSvcC](i)
		return &benchSvcTop{}, nil
	})
	_ = MustInvoke[*benchSvcTop](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = Invoke[*benchSvcTop](injector)
	}
}

//
// Registration
//

func BenchmarkProvide(b *testing.B) {
	names := make([]string, 100)
	for j := 0; j < 100; j++ {
		names[j] = fmt.Sprintf("bench-svc-%d", j)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		injector := New()
		for j := 0; j < 100; j++ {
			ProvideNamedValue(injector, names[j], &benchSvcA{})
		}
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = New()
	}
}

//
// Lifecycle
//

func BenchmarkHealthCheck(b *testing.B) {
	injector := New()
	for j := 0; j < 50; j++ {
		name := fmt.Sprintf("bench-hc-%d", j)
		ProvideNamed(injector, name, func(i Injector) (*benchHealthchecker, error) {
			return &benchHealthchecker{}, nil
		})
		_ = MustInvokeNamed[*benchHealthchecker](injector, name)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = injector.HealthCheck()
	}
}

func BenchmarkShutdown(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		b.StopTimer()

		// Build a 20-service dependency chain, fully instantiated.
		injector := New()
		ProvideNamed(injector, "bench-chain-0", func(i Injector) (*benchShutdowner, error) {
			return &benchShutdowner{}, nil
		})
		for j := 1; j < 20; j++ {
			name := fmt.Sprintf("bench-chain-%d", j)
			dependency := fmt.Sprintf("bench-chain-%d", j-1)
			ProvideNamed(injector, name, func(i Injector) (*benchShutdowner, error) {
				_ = MustInvokeNamed[*benchShutdowner](i, dependency)
				return &benchShutdowner{}, nil
			})
		}
		_ = MustInvokeNamed[*benchShutdowner](injector, "bench-chain-19")

		b.StartTimer()
		_ = injector.Shutdown()
	}
}

//
// Introspection
//

func BenchmarkListProvidedServices(b *testing.B) {
	injector := New()
	var scope Injector = injector
	for level := 0; level < 3; level++ {
		for j := 0; j < 10; j++ {
			ProvideNamedValue(scope, fmt.Sprintf("bench-svc-%d-%d", level, j), &benchSvcA{})
		}
		scope = scope.Scope(fmt.Sprintf("level-%d", level+1))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = scope.ListProvidedServices()
	}
}

func BenchmarkAncestors(b *testing.B) {
	injector := New()
	scope := injector.Scope("level-1")
	for j := 2; j <= 10; j++ {
		scope = scope.Scope(fmt.Sprintf("level-%d", j))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = scope.Ancestors()
	}
}
