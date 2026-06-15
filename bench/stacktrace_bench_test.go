package bench

import (
	"reflect"
	"testing"

	"github.com/samber/do/v2/stacktrace"
)

func BenchmarkNewFrameFromCaller(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _ = stacktrace.NewFrameFromCaller()
	}
}

func BenchmarkNewFrameFromPC(b *testing.B) {
	pc := reflect.ValueOf(stacktrace.NewFrameFromPC).Pointer()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = stacktrace.NewFrameFromPC(pc)
	}
}
