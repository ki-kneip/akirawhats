package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/whatsapp"
)

func RegistrarRotas(e *gin.Engine, svc internal.UserService, sm *whatsapp.SessionManager) {
	e.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	registrarDocs(e)

	apiGroup := e.Group("/api")
	registrarAuth(apiGroup, svc)

	protected := apiGroup.Group("", jwtMiddleware())
	registrarUser(protected, svc)
	registrarWhatsApp(protected, sm)
	protected.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
}
