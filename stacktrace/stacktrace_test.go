package stacktrace

// func a() *Stacktrace {
// 	return b()
// }

// func b() *Stacktrace {
// 	return c()
// }

// func c() *Stacktrace {
// 	return d()
// }

// func d() *Stacktrace {
// 	return e()
// }

// func e() *Stacktrace {
// 	return f()
// }

// func f() *Stacktrace {
// 	return NewStacktrace()
// }

// func TestStacktrace(t *testing.T) {
// 	is := assert.New(t)

// 	st := a()

// 	is.NotNil(st)

// 	if st.frames != nil {
// 		for _, f := range st.frames {
// 			is.True(strings.Contains(f.file, "do/debug/stacktrace_test.go"))
// 		}

// 		is.Len(st.frames, 7, "expected 7 frames")

// 		if len(st.frames) == 7 {
// 			is.Equal("f", (st.frames)[0].function)
// 			is.Equal("e", (st.frames)[1].function)
// 			is.Equal("d", (st.frames)[2].function)
// 			is.Equal("c", (st.frames)[3].function)
// 			is.Equal("b", (st.frames)[4].function)
// 			is.Equal("a", (st.frames)[5].function)
// 			is.Equal("TestStacktrace", (st.frames)[6].function)
// 		}
// 	}
// }
