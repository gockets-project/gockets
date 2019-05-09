package routes

import (
	"github.com/gin-gonic/gin"
	"gockets/controllers"
)

func InitWSRoutes(r *gin.Engine) {
	cr := r.Group("channel/subscribe")
	{
		cr.GET(":subscriberToken", controllers.CreateConnection)
	}
}
