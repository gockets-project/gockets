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
		sr := mr.Group("show")
		{
			sr.GET("", controllers.GetAllChannels)
			sr.GET(":publisherToken", controllers.GetChannel)
		}

		pr := mr.Group("publish")
		{
			pr.POST(":publisherToken", controllers.PushToConnection)
			pr.DELETE(":publisherToken", controllers.CloseConnection)
		}

		er := mr.Group("edit")
		{
			er.PATCH(":publisherToken", controllers.EditChannel)
		}
	}
}
