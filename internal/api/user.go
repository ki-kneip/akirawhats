package api

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"kikneip.com/akirawhats/internal"
	"kikneip.com/akirawhats/internal/model"
)

func registrarUser(e gin.IRouter) {
	grp := e.Group("/user")
	grp.POST("", func(c *gin.Context) {
		var svc internal.UserService
		var req model.UserDTOPost
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
		defer cancel()

		if err := c.BindJSON(&req); err != nil {
			_ = c.Error(err)
			return
		}
		res, err := svc.CreateUser(ctx, req)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(200, res)
	})
	grp.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		var svc internal.UserService
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
		defer cancel()
		res, err := svc.GetUserByID(ctx, id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(200, res)
	})
	grp.GET("", func(c *gin.Context) {
		var svc internal.UserService
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
		defer cancel()
		res, err := svc.GetAllUsers(ctx)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(200, res)
	})
	grp.DELETE("/:id", func(c *gin.Context) {
		var svc internal.UserService
		id := c.Param("id")
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*5)
		defer cancel()
		err := svc.DeleteUser(ctx, id)
		if err != nil {
			_ = c.Error(err)
			return
		}
		c.JSON(200, gin.H{
			"message": "user deleted",
		})
	})
}
