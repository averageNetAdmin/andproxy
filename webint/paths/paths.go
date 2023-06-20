package paths

import (
	"github.com/averageNetAdmin/andproxy/webint/handlers"
	"github.com/gin-gonic/gin"
)

func InitPaths(router *gin.Engine) {
	router.POST("/servers", handlers.CreateServer)
}
