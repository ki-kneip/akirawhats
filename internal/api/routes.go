package api

import "github.com/gin-gonic/gin"

func RegistrarRotas(e gin.IRouter) {
	api := e.Group("/api")

	registrarUser(api)

	api.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
