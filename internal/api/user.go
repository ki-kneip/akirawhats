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

	grp.POST("", func(c *gin.Context) {
		var req model.UserDTOPost
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.CreateUser(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, res)
	})

	grp.GET("", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.GetAllUsers(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	grp.GET("/:id", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.GetUserByID(ctx, c.Param("id"))
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

	grp.PUT("/:id", func(c *gin.Context) {
		var req model.UserDTOPut
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		res, err := svc.UpdateUser(ctx, c.Param("id"), req)
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

	grp.DELETE("/:id", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := svc.DeleteUser(ctx, c.Param("id")); err != nil {
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
