package stacktrace

import (
	"fmt"
	"runtime"
	"strings"
)

///
/// Inspired by palantir/stacktrace repo
/// -> https://github.com/palantir/stacktrace/blob/master/stacktrace.go
/// -> Apache 2.0 LICENSE
///

type fake struct{}

var (
	// packageName = reflect.TypeOf(fake{}).PkgPath()
	packageName           = "samber/do"
	packageNameStacktrace = packageName + "/stacktrace/"
	packageNameExamples   = packageName + "/examples/"
)

func NewFrameFromCaller() (Frame, bool) {
	// find the first frame that is not in this package
	for i := 0; i < 10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = removeGoPath(file)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		function := shortFuncName(f.Name())

		isGoPkg := strings.Contains(file, runtime.GOROOT())                // skip frames in GOROOT
		isDoPkg := strings.Contains(file, packageName)                     // skip frames in this package
		isDoStacktracePkg := strings.Contains(file, packageNameStacktrace) // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)        // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")                    // do not skip frames in tests

		if !isGoPkg && (!isDoPkg || !isDoStacktracePkg || isExamplePkg || isTestPkg) {
			return Frame{
				PC:       pc,
				File:     file,
				Function: function,
				Line:     line,
			}, true
		}
	}

	return Frame{}, false
}

func NewFrameFromPtr(pc uintptr) (Frame, bool) {
	fun := runtime.FuncForPC(pc)
	if fun == nil {
		return Frame{}, false
	}

	function := fun.Name()
	file, line := fun.FileLine(pc)

	file = removeGoPath(file)
	function = shortFuncName(function)

	return Frame{
		PC:       pc,
		File:     file,
		Function: function,
		Line:     line,
	}, true
}

type Frame struct {
	PC       uintptr
	File     string
	Function string
	Line     int
}

func (f Frame) String() string {
	return fmt.Sprintf("%s:%s:%d", f.File, f.Function, f.Line)
}

func shortFuncName(longName string) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}
