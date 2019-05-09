package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gockets/controllers"
	"gockets/middleware"
)

func InitManagementRoutes(r *gin.Engine) {
	mr := r.Group("channel")
	{
		if viper.GetBool("managementProtect") {
			mr.Use(middleware.NewManagementRouteProtection())
		}
		mr.POST("prepare", controllers.PrepareChannel)
		mr.GET("show/", controllers.GetAllChannels)
		mr.GET("show/:publisherToken", controllers.GetChannel)

		pr := mr.Group("publish")
		{
			pr.POST(":publisherToken", controllers.PushToConnection)
			pr.DELETE(":publisherToken", controllers.CloseConnection)
		}
	}
}
