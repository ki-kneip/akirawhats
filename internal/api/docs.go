package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>AkiraWhats API</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/docs/openapi.yaml",
    dom_id: '#swagger-ui',
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  })
</script>
</body>
</html>`

func registrarDocs(e gin.IRouter) {
	e.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
	e.StaticFile("/docs/openapi.yaml", "./docs/openapi.yaml")
}
