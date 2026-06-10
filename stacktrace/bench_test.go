package stacktrace

import (
	"reflect"
	"testing"
)

func BenchmarkNewFrameFromCaller(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_, _ = NewFrameFromCaller()
	}
}

func BenchmarkNewFrameFromPC(b *testing.B) {
	pc := reflect.ValueOf(NewFrameFromPC).Pointer()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = NewFrameFromPC(pc)
	}
}

func BenchmarkRemoveGoPath(b *testing.B) {
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		_ = removeGoPath("/home/user/go/src/github.com/samber/do/di.go")
	}
}
