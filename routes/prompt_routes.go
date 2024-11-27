package routes

import (
	"aiagent/config"
	"aiagent/controllers"

	"github.com/gin-gonic/gin"
)

func PromptRoutes(router *gin.Engine) {
	aIPromptRepo := config.GetRepoCollection("AIPrompts")

	router.GET("/initializ/v1/ai/prompts", controllers.GetPrompts(aIPromptRepo))
	router.GET("/initializ/v1/ai/prompt/:promptId", controllers.GetPromptsByID(aIPromptRepo))
}