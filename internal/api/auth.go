package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	internal_pkg "kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/model"
)

func registrarAuth(e gin.IRouter, svc internal_pkg.UserService) {
	grp := e.Group("/auth")

	grp.POST("/register", func(c *gin.Context) {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		token, err := issueToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		c.JSON(http.StatusCreated, model.AuthResponse{Token: token, User: *user})
	})

	grp.POST("/login", func(c *gin.Context) {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		token, err := issueToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		c.JSON(http.StatusOK, model.AuthResponse{Token: token, User: *user})
	})
}
