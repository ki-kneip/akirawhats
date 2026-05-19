package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	internal_pkg "kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/model"
)

func registrarAuth(e gin.IRouter, svc internal_pkg.UserService) {
	grp := e.Group("/auth")

	loginRL := newRateLimiter(5, 10)    // burst 5, 10 req/min por IP
	registerRL := newRateLimiter(3, 5)  // burst 3, 5 req/min por IP
	startRateLimitGC(loginRL, registerRL)

	grp.POST("/register", rateLimitMiddleware(registerRL), func(c *gin.Context) {
		var req model.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		user, err := svc.CreateUser(ctx, model.UserDTOPost{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  req.Password,
		})
		if err != nil {
			if errors.Is(err, internal_pkg.ErrAlreadyExists) {
				c.JSON(http.StatusConflict, gin.H{"error": "email já cadastrado"})
				return
			}
			log.Printf("register error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		token, err := issueToken(user.ID, user.Role)
		if err != nil {
			log.Printf("token issue error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusCreated, model.AuthResponse{Token: token, User: *user})
	})

	grp.POST("/login", rateLimitMiddleware(loginRL), func(c *gin.Context) {
		var req model.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		user, err := svc.AuthenticateUser(ctx, req.Email, req.Password)
		if err != nil {
			if errors.Is(err, internal_pkg.ErrInvalidCredentials) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			log.Printf("login error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		token, err := issueToken(user.ID, user.Role)
		if err != nil {
			log.Printf("token issue error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, model.AuthResponse{Token: token, User: *user})
	})
}
