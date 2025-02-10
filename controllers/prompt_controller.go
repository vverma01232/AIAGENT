package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		findOptions := options.Find()
		findOptions.SetSort(bson.M{"created_at": -1})
		cursor, err := aIPromptsRepo.FindWithOption(bson.M{}, findOptions)
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

// SavePrompt				godoc
// @Tags					Prompt Apis
// @Summary					Save Prompt
// @Description				Save Prompt
// @Param					Prompt body models.Prompts true "Add the prompt in the Db"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/saveprompt [POST]
func SavePrompt(aIPromptsRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.Prompts
		c.BindJSON(&body)
		body.UpdatedAt = time.Now()
		body.CreatedAt = time.Now()
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

// Uploadprompt				godoc
// @Tags					Prompt Apis
// @Summary					Update Prompt
// @Description				Update Prompt In Db
// @Param					promptId path string true "promptId"
// @Param					Prompt body models.Prompts true "Update the prompt in the Db"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/updateprompt/{promptId} [PUT]
func UpdatePromptById(aIPromptsRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.Prompts
		promptId := c.Param("promptId")
		if promptId == "" {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Prompt Id is required",
			})
			return
		}

		err := c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    http.StatusBadRequest,
				"message": "Error occured while mapping the body : " + err.Error(),
			})
			return
		}
		promptObjectId, _ := primitive.ObjectIDFromHex(promptId)
		filter := primitive.M{
			"_id": promptObjectId,
		}
		update := primitive.M{
			"$set": primitive.M{
				"updated_at":  time.Now(),
				"updated_by":  body.UpdatedBy,
				"prompt":      body.Prompt,
				"prompt_rule": body.PromptRule,
			},
		}
		err = aIPromptsRepo.UpdateOne(filter, update, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error updating prompt",
			})
			return
		}

		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully saved the prompt.",
		})
	}
}
