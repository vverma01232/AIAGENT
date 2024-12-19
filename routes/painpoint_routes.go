package routes

import (
	"aiagent/config"
	"aiagent/controllers"

	"github.com/gin-gonic/gin"
)

func PainPointRoutes(router *gin.Engine) {
	painPointRepo := config.GetRepoCollection("PainPoints")

	router.GET("/initializ/v1/ai/painpoints", controllers.GetPainPoints(painPointRepo))
	router.POST("/initializ/v1/ai/painpoints", controllers.SaveAiResponseToDB(painPointRepo))
	router.DELETE("/initializ/v1/ai/painpoints/:id", controllers.DeletePainPoints(painPointRepo))
}
