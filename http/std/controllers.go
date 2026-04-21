package dohttpstd

import (
	"fmt"
	"net/http"

	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

// Use creates an HTTP handler for the do library's web-based debugging interface
// using the standard Go net/http package. This function returns a handler that
// can be integrated with any standard HTTP server or router.
//
// Parameters:
//   - basePath: The base URL path for the debugging interface
//   - injector: The injector instance to debug
//
// Returns an http.Handler that serves the debugging interface.
//
// The handler sets up the following routes:
//   - GET /: The main debugging interface home page
//   - GET /scope: Scope tree visualization with optional scope_id parameter
//   - GET /service: Service inspection with optional scope_id and service_name parameters
//
// Example:
//
//	// Create the debugging handler
//	debugHandler := std.Use("/debug/di", injector)
//
//	// Mount it in your server
//	mux := http.NewServeMux()
//	mux.Handle("/debug/di/", debugHandler)
//
//	// Your application routes
//	mux.HandleFunc("/api/users", userHandler)
//
//	// Start the server
//	http.ListenAndServe(":8080", mux)
//
// The debugging interface will be available at /debug/di and provides:
//   - Visual representation of scope hierarchy
//   - Service dependency graphs
//   - Service inspection and debugging tools
//   - Navigation between different views
//
// The handler automatically strips the basePath prefix from incoming requests
// to ensure proper routing within the debugging interface.
//
// Security:
// Do not expose this debug UI publicly in production. It reveals internal details
// about your application's DI graph. Protect these routes with authentication
// (for example, Basic Auth) and/or network restrictions (IP allowlist, VPN, etc.).
// Wrap this handler with your auth/middleware before mounting it.
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
		// bearer:disable go_lang_information_leakage
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = w.Write(output)
	if err != nil {
		// bearer:disable go_lang_information_leakage
		http.Error(w, err.Error(), 500)
	}
}
