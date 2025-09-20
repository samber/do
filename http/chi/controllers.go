package dochi

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

// Use integrates the do library's web-based debugging interface with a Chi router.
// This function sets up HTTP routes for the debugging UI, allowing you to inspect
// your DI container through a web browser using the Chi framework.
//
// Parameters:
//   - router: The Chi router to add the debugging routes to
//   - basePath: The base URL path for the debugging interface
//   - injector: The injector instance to debug
//
// The function sets up the following routes:
//   - GET {basePath}: The main debugging interface home page
//   - GET {basePath}/scope: Scope tree visualization with optional scope_id parameter
//   - GET {basePath}/service: Service inspection with optional scope_id and service_name parameters
//
// Example:
//
//	r := chi.NewRouter()
//	api := r.Route("/api", nil)
//
//	// Add the debugging interface
//	dochi.Use(r, "/debug/di", injector)
//
//	// Your application routes
//	api.Get("/users", userHandler)
//
// The debugging interface will be available at /debug/di and provides:
//   - Visual representation of scope hierarchy
//   - Service dependency graphs
//   - Service inspection and debugging tools
//   - Navigation between different views
//
// Chi routes are registered with the full basePath, so the debugging interface
// will be available at the exact path specified in basePath.
//
// Security:
// Do not expose these routes publicly in production. They reveal internal details
// about your DI graph. Protect the basePath with authentication (e.g., Basic Auth)
// and/or network restrictions. Apply your auth middleware before calling Use.
func Use(router *chi.Mux, basePath string, injector do.Injector) {
	router.Get(basePath, func(w http.ResponseWriter, r *http.Request) {
		output, err := dohttp.IndexHTML(basePath)
		response(w, output, err)
	})

	router.Get(basePath+"/scope", func(w http.ResponseWriter, r *http.Request) {
		scopeID := r.URL.Query().Get("scope_id")
		if scopeID == "" {
			url := fmt.Sprintf("%s/scope?scope_id=%s", basePath, injector.ID())
			http.Redirect(w, r, url, 302)
			return
		}

		output, err := dohttp.ScopeTreeHTML(basePath, injector, scopeID)
		response(w, output, err)
	})

	router.Get(basePath+"/service", func(w http.ResponseWriter, r *http.Request) {
		scopeID := r.URL.Query().Get("scope_id")
		serviceName := r.URL.Query().Get("service_name")

		if scopeID == "" || serviceName == "" {
			output, err := dohttp.ServiceListHTML(basePath, injector)
			response(w, output, err)
			return
		}

		output, err := dohttp.ServiceHTML(basePath, injector, scopeID, serviceName)
		response(w, output, err)
	})
}

func response(w http.ResponseWriter, data string, err error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(200)
		w.Write([]byte(data))
	}
}
