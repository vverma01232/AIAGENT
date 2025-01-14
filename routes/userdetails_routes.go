package routes

import (
	"aiagent/config"
	"aiagent/controllers"

	"github.com/gin-gonic/gin"
)

func UserDataRouter(router *gin.Engine) {
	userDataRepo := config.GetRepoCollection("UserData")
	promptRepo := config.GetRepoCollection("AIPrompts")
	router.POST("/initializ/v1/ai/upload", controllers.UploadExcel(userDataRepo, promptRepo))
	router.GET("/initializ/v1/ai/allusers", controllers.GetAllUserData(userDataRepo))
}
