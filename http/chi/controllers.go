package gin

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

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
