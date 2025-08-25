package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

// Use integrates the do library's web-based debugging interface with a Gin router.
// This function sets up HTTP routes for the debugging UI, allowing you to inspect
// your DI container through a web browser.
//
// Parameters:
//   - router: The Gin router group to add the debugging routes to
//   - injector: The injector instance to debug
//
// The function sets up the following routes:
//   - GET /: The main debugging interface home page
//   - GET /scope: Scope tree visualization with optional scope_id parameter
//   - GET /service: Service inspection with optional scope_id and service_name parameters
//
// Example:
//
//	router := gin.Default()
//	api := router.Group("/api")
//	debug := router.Group("/debug/di")
//
//	// Add the debugging interface
//	dogin.Use(debug, injector)
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
// (e.g., Basic Auth) and/or network restrictions, since it can leak internals
// about your application's DI graph. Attach auth middleware to the router group
// before calling Use.
func Use(router *gin.RouterGroup, injector do.Injector) {
	basePathDo := router.BasePath()

	router.Handle("GET", "", func(c *gin.Context) {
		output, err := dohttp.IndexHTML(basePathDo)
		response(c, output, err)
	})

	router.Handle("GET", "/scope", func(c *gin.Context) {
		scopeID := c.Query("scope_id")
		if scopeID == "" {
			url := fmt.Sprintf("%s/scope?scope_id=%s", basePathDo, injector.ID())
			c.Redirect(302, url)
			return
		}

		output, err := dohttp.ScopeTreeHTML(basePathDo, injector, scopeID)
		response(c, output, err)
	})

	router.Handle("GET", "/service", func(c *gin.Context) {
		scopeID := c.Query("scope_id")
		serviceName := c.Query("service_name")

		if scopeID == "" || serviceName == "" {
			output, err := dohttp.ServiceListHTML(basePathDo, injector)
			response(c, output, err)
			return
		}

		output, err := dohttp.ServiceHTML(basePathDo, injector, scopeID, serviceName)
		response(c, output, err)
	})
}

func response(c *gin.Context, data string, err error) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		c.String(500, err.Error())
	} else {
		c.String(200, data)
	}
}
