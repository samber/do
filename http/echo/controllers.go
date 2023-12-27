package gin

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

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
