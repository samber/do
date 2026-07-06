package do

import "testing"

func BenchmarkOrderedUniq(b *testing.B) {
	// Mirrors the typical caller (Scope.ListProvidedServices/ListInvokedServices):
	// mostly-unique service names aggregated across ancestor scopes, with only a
	// handful of names shadowed between scopes.
	in := make([]int, 200)
	for i := range in {
		in[i] = i
	}
	for i := 190; i < 200; i++ {
		in[i] = i - 190
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = orderedUniq(in)
	}
}

func BenchmarkNewUUID(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _ = newUUID()
	}
}
