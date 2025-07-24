package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	dohttp "github.com/samber/do/v2/http"
)

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
