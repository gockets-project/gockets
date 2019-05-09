package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gockets/helpers"
	"gockets/models"
	"gockets/src/services/logger"
)

func NewManagementRouteProtection() func(c *gin.Context) {
	configAllowedHostnames := viper.GetStringSlice("allowedHostnames")
	port := viper.GetInt("port")

	var resolvedHostnames []string

	for _, hostname := range configAllowedHostnames {
		resolvedIp, err := helpers.LookupName(hostname)
		if err != nil {
			continue
		}

		resolvedHostnames = append(resolvedHostnames,
			resolvedIp,
			hostname,
			fmt.Sprintf("%s:%d", resolvedIp, port),
			fmt.Sprintf("%s:%d", hostname, port),
		)
	}

	ll.Log.Debug(resolvedHostnames)

	return func(c *gin.Context) {
		requestHostname := c.Request.Host
		ll.Log.Debug(requestHostname)
		if !helpers.SliceContains(resolvedHostnames, requestHostname) {
			c.AbortWithStatusJSON(403, models.Response{
				Type:    "err",
				Message: "Unauthorized",
			})
		}
		c.Next()
	}
}
