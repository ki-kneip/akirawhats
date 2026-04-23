package api

import (
	"github.com/gin-gonic/gin"
	"kikneip.com/akirawhats/internal/repo"
)

func registrarUser(e gin.IRouter) {
	grp := e.Group("/user")
	grp.POST("", CreateUser)
	repo.UserImpl{}
}
