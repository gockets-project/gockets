package routes

import (
	"github.com/gin-gonic/gin"
	"gockets/models"
)

func InitRoutes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, models.Response{
			Type:    "err",
			Message: "Endpoint not found",
		})
	})

	InitWSRoutes(r)
	InitManagementRoutes(r)

	return r
}
