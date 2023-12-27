package gin

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

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
