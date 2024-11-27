package main

import (
	"aiagent/config"
	_ "aiagent/docs"
	"aiagent/routes"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	cors "github.com/itsjamie/gin-cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title 		Init App aiagent
// @version		1.0
// @description Init App aiagent Open Api Spec
// @BaseUrl  	/
func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	config.LoadEnv()
	//initiate db connection
	config.ConnectDB()

	corsConfig := cors.Config{
		Origins:         "*",
		RequestHeaders:  "Origin, Authorization, Content-Type,App-User, Org_id, User-Agent",
		Methods:         "GET, POST, PUT,DELETE",
		Credentials:     false,
		ValidateHeaders: false,
		MaxAge:          10 * time.Minute,
	}
	router := gin.Default()
	router.Use(cors.Middleware(corsConfig))
	// Implmenting Swagger
	router.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	routes.AgentRoutes(router)
	routes.UserDataRouter(router)
	routes.PromptRoutes(router)

	router.Run(":8081")
	log.Infof("Server listening on http://localhost:8081/")
	if err := http.ListenAndServe("0.0.0.0:8081", router); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
