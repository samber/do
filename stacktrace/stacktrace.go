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

// type fake struct{}

var (
	// packageName = reflect.TypeOf(fake{}).PkgPath()
	packageName           = "do"
	packageNameStacktrace = packageName + "/stacktrace/"
	packageNameExamples   = packageName + "/examples/"
)

// NewFrameFromCaller creates a new Frame from the current call stack.
// This function walks up the call stack to find the first frame that is not
// in the do package or Go runtime, providing useful debugging information
// about where a service was invoked from.
//
// The function filters out:
//   - Frames in the Go runtime (GOROOT)
//   - Frames in the do package (except examples and tests)
//   - Frames in the stacktrace package
//
// Returns a Frame representing the caller and a boolean indicating success.
// The boolean is false if no suitable frame was found.
//
// This function is used internally by the DI container to track service
// invocation locations for debugging and explanation purposes.
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
		isExamplePkg := strings.Contains(
			file,
			packageNameExamples,
		) // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go") // do not skip frames in tests

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

// NewFrameFromPC creates a new Frame from a program counter (PC) value.
// This function is used to create Frame objects from function pointers,
// typically for tracking where service providers were defined.
//
// Parameters:
//   - pc: The program counter value representing a function
//
// Returns a Frame representing the function and a boolean indicating success.
// The boolean is false if the PC value is invalid.
//
// This function is used internally to track service provider locations
// for debugging and explanation purposes.
func NewFrameFromPC(pc uintptr) (Frame, bool) {
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

// Frame represents a single stack frame with debugging information.
// This struct contains information about a function call location,
// including the file, line number, function name, and program counter.
type Frame struct {
	PC       uintptr
	File     string
	Function string
	Line     int
}

// String returns a formatted string representation of the frame.
// The format is "file:function:line" which is useful for debugging
// and logging purposes.
//
// Returns a string in the format "path/to/file.go:FunctionName:123".
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
