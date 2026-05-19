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

func registrarUser(e gin.IRouter, svc internal_pkg.UserService) {
	grp := e.Group("/user")

	// Admin-only: list all users
	grp.GET("", adminOnly(), func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.GetAllUsers(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Returns the current authenticated user's profile.
	grp.GET("/me", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.GetUserByID(ctx, getUserID(c))
		if err != nil {
			if errors.Is(err, internal_pkg.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	grp.PUT("/me", func(c *gin.Context) {
		var req model.UserDTOPut
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.UpdateUser(ctx, getUserID(c), req)
		if err != nil {
			if errors.Is(err, internal_pkg.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	grp.PUT("/me/password", func(c *gin.Context) {
		var req model.ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := svc.ChangePassword(ctx, getUserID(c), req.CurrentPassword, req.NewPassword); err != nil {
			if errors.Is(err, internal_pkg.ErrInvalidCredentials) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "senha atual incorreta"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "senha alterada"})
	})

	grp.DELETE("/me", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := svc.DeleteUser(ctx, getUserID(c)); err != nil {
			if errors.Is(err, internal_pkg.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
	})
}
