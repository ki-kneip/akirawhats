package api

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/whatsapp"
)

func RegistrarRotas(e *gin.Engine, svc internal.UserService, sm *whatsapp.SessionManager) {
	e.Use(cors.New(corsConfig()))

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

func corsConfig() cors.Config {
	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}
	raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	if raw == "" || raw == "*" {
		log.Println("[WARN] CORS_ORIGINS não configurado — permitindo qualquer origem (não recomendado em produção)")
		cfg.AllowAllOrigins = true
		return cfg
	}
	origins := make([]string, 0)
	for _, o := range strings.Split(raw, ",") {
		if s := strings.TrimSpace(o); s != "" {
			origins = append(origins, s)
		}
	}
	cfg.AllowOrigins = origins
	return cfg
}
