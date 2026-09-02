package stacktrace

import "testing"

func BenchmarkShortFuncName(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = shortFuncName("github.com/samber/do/v2.(*ServiceLazy[int]).getInstance")
	}
}
