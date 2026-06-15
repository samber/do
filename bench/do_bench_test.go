package bench

import (
	"fmt"
	"testing"

	"github.com/samber/do/v2"
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
	injector := do.New()
	do.Provide(injector, func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = do.MustInvoke[*benchSvcA](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.Invoke[*benchSvcA](injector)
	}
}

func BenchmarkInvokeLazyParallel(b *testing.B) {
	injector := do.New()
	do.Provide(injector, func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = do.MustInvoke[*benchSvcA](injector)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = do.Invoke[*benchSvcA](injector)
		}
	})
}

func BenchmarkInvokeNamed(b *testing.B) {
	injector := do.New()
	do.ProvideNamed(injector, "bench-svc", func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	_ = do.MustInvokeNamed[*benchSvcA](injector, "bench-svc")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.InvokeNamed[*benchSvcA](injector, "bench-svc")
	}
}

func BenchmarkInvokeTransient(b *testing.B) {
	injector := do.New()
	do.ProvideTransient(injector, func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.Invoke[*benchSvcA](injector)
	}
}

func BenchmarkInvokeAlias(b *testing.B) {
	injector := do.New()
	do.Provide(injector, func(i do.Injector) (*benchSvcImpl, error) {
		return &benchSvcImpl{}, nil
	})
	if err := do.As[*benchSvcImpl, benchIface](injector); err != nil {
		b.Fatal(err)
	}
	_ = do.MustInvoke[benchIface](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.Invoke[benchIface](injector)
	}
}

func BenchmarkInvokeAs(b *testing.B) {
	injector := do.New()
	for j := 0; j < 19; j++ {
		do.ProvideNamedValue(injector, fmt.Sprintf("bench-filler-%d", j), &benchSvcA{})
	}
	do.Provide(injector, func(i do.Injector) (*benchSvcImpl, error) {
		return &benchSvcImpl{}, nil
	})
	_ = do.MustInvokeAs[benchIface](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.InvokeAs[benchIface](injector)
	}
}

func BenchmarkInvokeNestedScopes(b *testing.B) {
	injector := do.New()
	do.Provide(injector, func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})

	scope := injector.Scope("level-1")
	for j := 2; j <= 5; j++ {
		scope = scope.Scope(fmt.Sprintf("level-%d", j))
	}
	_ = do.MustInvoke[*benchSvcA](scope)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.Invoke[*benchSvcA](scope)
	}
}

func BenchmarkProviderChainTransient(b *testing.B) {
	injector := do.New()
	do.Provide(injector, func(i do.Injector) (*benchSvcA, error) {
		return &benchSvcA{}, nil
	})
	do.Provide(injector, func(i do.Injector) (*benchSvcB, error) {
		return &benchSvcB{}, nil
	})
	do.Provide(injector, func(i do.Injector) (*benchSvcC, error) {
		return &benchSvcC{}, nil
	})
	do.ProvideTransient(injector, func(i do.Injector) (*benchSvcTop, error) {
		_ = do.MustInvoke[*benchSvcA](i)
		_ = do.MustInvoke[*benchSvcB](i)
		_ = do.MustInvoke[*benchSvcC](i)
		return &benchSvcTop{}, nil
	})
	_ = do.MustInvoke[*benchSvcTop](injector)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = do.Invoke[*benchSvcTop](injector)
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
		injector := do.New()
		for j := 0; j < 100; j++ {
			do.ProvideNamedValue(injector, names[j], &benchSvcA{})
		}
	}
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = do.New()
	}
}

//
// Lifecycle
//

func BenchmarkHealthCheck(b *testing.B) {
	injector := do.New()
	for j := 0; j < 50; j++ {
		name := fmt.Sprintf("bench-hc-%d", j)
		do.ProvideNamed(injector, name, func(i do.Injector) (*benchHealthchecker, error) {
			return &benchHealthchecker{}, nil
		})
		_ = do.MustInvokeNamed[*benchHealthchecker](injector, name)
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
		injector := do.New()
		do.ProvideNamed(injector, "bench-chain-0", func(i do.Injector) (*benchShutdowner, error) {
			return &benchShutdowner{}, nil
		})
		for j := 1; j < 20; j++ {
			name := fmt.Sprintf("bench-chain-%d", j)
			dependency := fmt.Sprintf("bench-chain-%d", j-1)
			do.ProvideNamed(injector, name, func(i do.Injector) (*benchShutdowner, error) {
				_ = do.MustInvokeNamed[*benchShutdowner](i, dependency)
				return &benchShutdowner{}, nil
			})
		}
		_ = do.MustInvokeNamed[*benchShutdowner](injector, "bench-chain-19")

		b.StartTimer()
		_ = injector.Shutdown()
	}
}

//
// Introspection
//

func BenchmarkListProvidedServices(b *testing.B) {
	injector := do.New()
	var scope do.Injector = injector
	for level := 0; level < 3; level++ {
		for j := 0; j < 10; j++ {
			do.ProvideNamedValue(scope, fmt.Sprintf("bench-svc-%d-%d", level, j), &benchSvcA{})
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
	injector := do.New()
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
