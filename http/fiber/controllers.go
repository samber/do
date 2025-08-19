package dofiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

// Use integrates the do library's web-based debugging interface with a Fiber router.
// This function sets up HTTP routes for the debugging UI, allowing you to inspect
// your DI container through a web browser using the Fiber framework.
//
// Parameters:
//   - router: The Fiber router to add the debugging routes to
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
//	app := fiber.New()
//	api := app.Group("/api")
//	debug := app.Group("/debug/di")
//
//	// Add the debugging interface
//	dofiber.Use(debug, "/debug/di", injector)
//
//	// Your application routes
//	api.Get("/users", userHandler)
//
// The debugging interface will be available at /debug/di and provides:
//   - Visual representation of scope hierarchy
//   - Service dependency graphs
//   - Service inspection and debugging tools
//   - Navigation between different views
func Use(router fiber.Router, basePath string, injector do.Injector) {
	router.Get("", func(c *fiber.Ctx) error {
		output, err := dohttp.IndexHTML(basePath)
		if err != nil {
			return err
		}

		response(c, output)
		return nil
	})

	router.Get("/scope", func(c *fiber.Ctx) error {
		scopeID := c.Query("scope_id")
		if scopeID == "" {
			url := fmt.Sprintf("%s/scope?scope_id=%s", basePath, injector.ID())
			c.Redirect(url, 302)
			return nil
		}

		output, err := dohttp.ScopeTreeHTML(basePath, injector, scopeID)
		if err != nil {
			return err
		}

		response(c, output)
		return nil
	})

	router.Get("/service", func(c *fiber.Ctx) error {
		scopeID := c.Query("scope_id")
		serviceName := c.Query("service_name")

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

func response(c *fiber.Ctx, data string) {
	c.Response().Header.Set("Content-Type", "text/html; charset=utf-8")
	c.Status(200).SendString(data)
}
