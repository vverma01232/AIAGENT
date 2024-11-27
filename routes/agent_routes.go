package routes

import (
	"aiagent/controllers"

	"github.com/gin-gonic/gin"
)

func AgentRoutes(router *gin.Engine) {
	router.POST("initializ/v1/ai/generatewithAI", controllers.GeneratewithAIHandler())
}
