package std

import (
	"fmt"
	"net/http"

	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

func Use(basePath string, injector do.Injector) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		output, err := dohttp.IndexHTML(basePath)
		response(w, []byte(output), err)
	})

	mux.HandleFunc("/scope", func(w http.ResponseWriter, r *http.Request) {
		scopeID := r.URL.Query().Get("scope_id")
		if scopeID == "" {
			url := fmt.Sprintf("%s/scope?scope_id=%s", basePath, injector.ID())
			http.Redirect(w, r, url, 302)
			return
		}

		output, err := dohttp.ScopeTreeHTML(basePath, injector, scopeID)
		response(w, []byte(output), err)
	})

	mux.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		scopeID := r.URL.Query().Get("scope_id")
		serviceName := r.URL.Query().Get("service_name")

		if scopeID == "" || serviceName == "" {
			output, err := dohttp.ServiceListHTML(basePath, injector)
			response(w, []byte(output), err)
			return
		}

		output, err := dohttp.ServiceHTML(basePath, injector, scopeID, serviceName)
		response(w, []byte(output), err)
	})

	return http.StripPrefix(basePath, mux)
}

func response(w http.ResponseWriter, output []byte, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = w.Write(output)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// // Cmdline responds with the running program's
// // command line, with arguments separated by NUL bytes.
// // The package initialization registers it as /debug/pprof/cmdline.
// func Cmdline(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	fmt.Fprint(w, strings.Join(os.Args, "\x00"))
// }

// func sleep(r *http.Request, d time.Duration) {
// 	select {
// 	case <-time.After(d):
// 	case <-r.Context().Done():
// 	}
// }

// func durationExceedsWriteTimeout(r *http.Request, seconds float64) bool {
// 	srv, ok := r.Context().Value(http.ServerContextKey).(*http.Server)
// 	return ok && srv.WriteTimeout != 0 && seconds >= srv.WriteTimeout.Seconds()
// }

// func serveError(w http.ResponseWriter, status int, txt string) {
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	w.Header().Set("X-Go-Pprof", "1")
// 	w.Header().Del("Content-Disposition")
// 	w.WriteHeader(status)
// 	fmt.Fprintln(w, txt)
// }

// // Profile responds with the pprof-formatted cpu profile.
// // Profiling lasts for duration specified in seconds GET parameter, or for 30 seconds if not specified.
// // The package initialization registers it as /debug/pprof/profile.
// func Profile(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	sec, err := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
// 	if sec <= 0 || err != nil {
// 		sec = 30
// 	}

// 	if durationExceedsWriteTimeout(r, float64(sec)) {
// 		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
// 		return
// 	}

// 	// Set Content Type assuming StartCPUProfile will work,
// 	// because if it does it starts writing.
// 	w.Header().Set("Content-Type", "application/octet-stream")
// 	w.Header().Set("Content-Disposition", `attachment; filename="profile"`)
// 	if err := pprof.StartCPUProfile(w); err != nil {
// 		// StartCPUProfile failed, so no writes yet.
// 		serveError(w, http.StatusInternalServerError,
// 			fmt.Sprintf("Could not enable CPU profiling: %s", err))
// 		return
// 	}
// 	sleep(r, time.Duration(sec)*time.Second)
// 	pprof.StopCPUProfile()
// }

// // Trace responds with the execution trace in binary form.
// // Tracing lasts for duration specified in seconds GET parameter, or for 1 second if not specified.
// // The package initialization registers it as /debug/pprof/trace.
// func Trace(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	sec, err := strconv.ParseFloat(r.FormValue("seconds"), 64)
// 	if sec <= 0 || err != nil {
// 		sec = 1
// 	}

// 	if durationExceedsWriteTimeout(r, sec) {
// 		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
// 		return
// 	}

// 	// Set Content Type assuming trace.Start will work,
// 	// because if it does it starts writing.
// 	w.Header().Set("Content-Type", "application/octet-stream")
// 	w.Header().Set("Content-Disposition", `attachment; filename="trace"`)
// 	if err := trace.Start(w); err != nil {
// 		// trace.Start failed, so no writes yet.
// 		serveError(w, http.StatusInternalServerError,
// 			fmt.Sprintf("Could not enable tracing: %s", err))
// 		return
// 	}
// 	sleep(r, time.Duration(sec*float64(time.Second)))
// 	trace.Stop()
// }

// // Symbol looks up the program counters listed in the request,
// // responding with a table mapping program counters to function names.
// // The package initialization registers it as /debug/pprof/symbol.
// func Symbol(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

// 	// We have to read the whole POST body before
// 	// writing any output. Buffer the output here.
// 	var buf bytes.Buffer

// 	// We don't know how many symbols we have, but we
// 	// do have symbol information. Pprof only cares whether
// 	// this number is 0 (no symbols available) or > 0.
// 	fmt.Fprintf(&buf, "num_symbols: 1\n")

// 	var b *bufio.Reader
// 	if r.Method == "POST" {
// 		b = bufio.NewReader(r.Body)
// 	} else {
// 		b = bufio.NewReader(strings.NewReader(r.URL.RawQuery))
// 	}

// 	for {
// 		word, err := b.ReadSlice('+')
// 		if err == nil {
// 			word = word[0 : len(word)-1] // trim +
// 		}
// 		pc, _ := strconv.ParseUint(string(word), 0, 64)
// 		if pc != 0 {
// 			f := runtime.FuncForPC(uintptr(pc))
// 			if f != nil {
// 				fmt.Fprintf(&buf, "%#x %s\n", pc, f.Name())
// 			}
// 		}

// 		// Wait until here to check for err; the last
// 		// symbol will have an err because it doesn't end in +.
// 		if err != nil {
// 			if err != io.EOF {
// 				fmt.Fprintf(&buf, "reading request: %v\n", err)
// 			}
// 			break
// 		}
// 	}

// 	w.Write(buf.Bytes())
// }

// // Handler returns an HTTP handler that serves the named profile.
// // Available profiles can be found in [runtime/pprof.Profile].
// func Handler(name string) http.Handler {
// 	return handler(name)
// }

// type handler string

// func (name handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	p := pprof.Lookup(string(name))
// 	if p == nil {
// 		serveError(w, http.StatusNotFound, "Unknown profile")
// 		return
// 	}
// 	if sec := r.FormValue("seconds"); sec != "" {
// 		name.serveDeltaProfile(w, r, p, sec)
// 		return
// 	}
// 	gc, _ := strconv.Atoi(r.FormValue("gc"))
// 	if name == "heap" && gc > 0 {
// 		runtime.GC()
// 	}
// 	debug, _ := strconv.Atoi(r.FormValue("debug"))
// 	if debug != 0 {
// 		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
// 	} else {
// 		w.Header().Set("Content-Type", "application/octet-stream")
// 		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
// 	}
// 	p.WriteTo(w, debug)
// }

// func (name handler) serveDeltaProfile(w http.ResponseWriter, r *http.Request, p *pprof.Profile, secStr string) {
// 	sec, err := strconv.ParseInt(secStr, 10, 64)
// 	if err != nil || sec <= 0 {
// 		serveError(w, http.StatusBadRequest, `invalid value for "seconds" - must be a positive integer`)
// 		return
// 	}
// 	if !profileSupportsDelta[name] {
// 		serveError(w, http.StatusBadRequest, `"seconds" parameter is not supported for this profile type`)
// 		return
// 	}
// 	// 'name' should be a key in profileSupportsDelta.
// 	if durationExceedsWriteTimeout(r, float64(sec)) {
// 		serveError(w, http.StatusBadRequest, "profile duration exceeds server's WriteTimeout")
// 		return
// 	}
// 	debug, _ := strconv.Atoi(r.FormValue("debug"))
// 	if debug != 0 {
// 		serveError(w, http.StatusBadRequest, "seconds and debug params are incompatible")
// 		return
// 	}
// 	p0, err := collectProfile(p)
// 	if err != nil {
// 		serveError(w, http.StatusInternalServerError, "failed to collect profile")
// 		return
// 	}

// 	t := time.NewTimer(time.Duration(sec) * time.Second)
// 	defer t.Stop()

// 	select {
// 	case <-r.Context().Done():
// 		err := r.Context().Err()
// 		if err == context.DeadlineExceeded {
// 			serveError(w, http.StatusRequestTimeout, err.Error())
// 		} else { // TODO: what's a good status code for canceled requests? 400?
// 			serveError(w, http.StatusInternalServerError, err.Error())
// 		}
// 		return
// 	case <-t.C:
// 	}

// 	p1, err := collectProfile(p)
// 	if err != nil {
// 		serveError(w, http.StatusInternalServerError, "failed to collect profile")
// 		return
// 	}
// 	ts := p1.TimeNanos
// 	dur := p1.TimeNanos - p0.TimeNanos

// 	p0.Scale(-1)

// 	p1, err = profile.Merge([]*profile.Profile{p0, p1})
// 	if err != nil {
// 		serveError(w, http.StatusInternalServerError, "failed to compute delta")
// 		return
// 	}

// 	p1.TimeNanos = ts // set since we don't know what profile.Merge set for TimeNanos.
// 	p1.DurationNanos = dur

// 	w.Header().Set("Content-Type", "application/octet-stream")
// 	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-delta"`, name))
// 	p1.Write(w)
// }

// func collectProfile(p *pprof.Profile) (*profile.Profile, error) {
// 	var buf bytes.Buffer
// 	if err := p.WriteTo(&buf, 0); err != nil {
// 		return nil, err
// 	}
// 	ts := time.Now().UnixNano()
// 	p0, err := profile.Parse(&buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	p0.TimeNanos = ts
// 	return p0, nil
// }

// var profileSupportsDelta = map[handler]bool{
// 	"allocs":       true,
// 	"block":        true,
// 	"goroutine":    true,
// 	"heap":         true,
// 	"mutex":        true,
// 	"threadcreate": true,
// }

// var profileDescriptions = map[string]string{
// 	"allocs":       "A sampling of all past memory allocations",
// 	"block":        "Stack traces that led to blocking on synchronization primitives",
// 	"cmdline":      "The command line invocation of the current program",
// 	"goroutine":    "Stack traces of all current goroutines. Use debug=2 as a query parameter to export in the same format as an unrecovered panic.",
// 	"heap":         "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
// 	"mutex":        "Stack traces of holders of contended mutexes",
// 	"profile":      "CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.",
// 	"threadcreate": "Stack traces that led to the creation of new OS threads",
// 	"trace":        "A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.",
// }

// type profileEntry struct {
// 	Name  string
// 	Href  string
// 	Desc  string
// 	Count int
// }
