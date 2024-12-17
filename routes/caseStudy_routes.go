package routes

import (
	"aiagent/config"
	"aiagent/controllers"

	"github.com/gin-gonic/gin"
)

func CaseStudyRoutes(router *gin.Engine) {
	caseStudyRepo := config.GetRepoCollection("CaseStudy")

	router.POST("/initializ/v1/ai/casestudy", controllers.SaveCaseStudy(caseStudyRepo))
	router.GET("/initializ/v1/ai/casestudy", controllers.GetCaseStudy(caseStudyRepo))
}
