package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetPrompts				godoc
// @Tags					Prompt Apis
// @Summary					Get Prompts
// @Description				Get all AI Prompts
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/prompts [GET]
func GetPrompts(aIPromptsRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {

		cursor, err := aIPromptsRepo.Find(bson.M{})
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occured while fetching the data from db : " + err.Error(),
			})
			return
		}
		defer cursor.Close(context.TODO())

		var prompts []models.Prompts
		err = cursor.All(context.TODO(), &prompts)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error occured while fetching the data from db : " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully fetched the prompts",
			Data:    prompts,
		})
	}
}

// GetPromptbyID			godoc
// @Tags					Prompt Apis
// @Summary					Get Prompt by ID
// @Description				Get AI Prompts by ID
// @Param					promptId path string true "promptId"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/prompt/{promptId} [GET]
func GetPromptsByID(aIPromptsRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		promptId := c.Param("promptId")
		if promptId == "" {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Prompt Id is required.",
			})
			return
		}
		objectId, _ := primitive.ObjectIDFromHex(promptId)
		var prompts models.Prompts
		err := aIPromptsRepo.FindOne(bson.M{"_id": objectId}).Decode(&prompts)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Error occured while fetching the data from db : " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully fetched the prompt by id",
			Data:    prompts,
		})
	}
}

// UploadExcel				godoc
// @Tags					UserData Apis
// @Summary					Upload Excel File
// @Description				Upload Excel File
// @Param					UploadExcel body models.Prompts true "Upload the prompt in the Db"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/saveprompt [POST]
func SavePrompt(aIPromptsRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.Prompts
		c.BindJSON(&body)
		_, err := aIPromptsRepo.InsertOne(body)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Error occured while saving the prompt in db : " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully saved the prompt.",
		})
	}
}
