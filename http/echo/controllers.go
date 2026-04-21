package doecho

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

// Use integrates the do library's web-based debugging interface with an Echo router.
// This function sets up HTTP routes for the debugging UI, allowing you to inspect
// your DI container through a web browser using the Echo framework.
//
// Parameters:
//   - router: The Echo router group to add the debugging routes to
//   - basePath: The base URL path for the debugging interface
//   - injector: The injector instance to debug
//
// The function sets up the following routes:
//   - GET /: The main debugging interface home page
//   - GET /scope: Scope tree visualization with optional scope_id parameter
//   - GET /service: Service inspection with optional scope_id and service_name parameters
//
// Example:
//
//	e := echo.New()
//	api := e.Group("/api")
//	debug := e.Group("/debug/di")
//
//	// Add the debugging interface
//	doecho.Use(debug, "/debug/di", injector)
//
//	// Your application routes
//	api.GET("/users", userHandler)
//
// The debugging interface will be available at /debug/di and provides:
//   - Visual representation of scope hierarchy
//   - Service dependency graphs
//   - Service inspection and debugging tools
//   - Navigation between different views
//
// Security:
// Do not expose this group publicly in production. Protect it with authentication
// (e.g., Basic Auth) and/or network restrictions, since it reveals internals
// about your DI graph. Attach auth middleware to the Echo group before Use.
func Use(router *echo.Group, basePath string, injector do.Injector) {
	router.GET("", func(c echo.Context) error {
		output, err := dohttp.IndexHTML(basePath)
		if err != nil {
			return err
		}

		response(c, output)
		return nil
	})

	router.GET("/scope", func(c echo.Context) error {
		scopeID := c.QueryParam("scope_id")
		if scopeID == "" {
			url := fmt.Sprintf("%s/scope?scope_id=%s", basePath, injector.ID())
			c.Redirect(302, url)
			return nil
		}

		output, err := dohttp.ScopeTreeHTML(basePath, injector, scopeID)
		if err != nil {
			return err
		}

		response(c, output)
		return nil
	})

	router.GET("/service", func(c echo.Context) error {
		scopeID := c.QueryParam("scope_id")
		serviceName := c.QueryParam("service_name")

		if scopeID == "" || serviceName == "" {
			output, err := dohttp.ServiceListHTML(basePath, injector)
			if err != nil {
				return err
			}

			response(c, output)
			return nil
		}

		output, err := dohttp.ServiceHTML(basePath, injector, scopeID, serviceName)
		if err != nil {
			return err
		}

		response(c, output)
		return nil
	})
}

func response(c echo.Context, data string) {
	// c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.HTML(200, data)
}
