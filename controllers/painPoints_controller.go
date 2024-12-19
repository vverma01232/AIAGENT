package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetPainPoints				godoc
// @Tags					Pain Points Apis
// @Summary					Get Pain Points and Value Proposition
// @Description				Get all Pain Points and Value Proposition
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/painpoints [GET]
func GetPainPoints(painPointRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.TODO()
		cursor, err := painPointRepo.Find(bson.M{})
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occurred while fetching pain points",
			})
			return
		}
		defer cursor.Close(context.TODO())

		var painPoints []models.PainPointModel
		err = cursor.All(ctx, &painPoints)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error occurred while fetching pain points: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully fetched the pain points",
			Data:    painPoints,
		})
	}
}

// SavePainPoints			godoc
// @Tags					Pain Points Apis
// @Summary					Save Pain Points and Value Proposition
// @Description				Save Pain Points and Value Proposition
// @Produce					application/json
// @Param                    request body models.PainPointRole true  "Pain Points and Value Proposition"
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/painpoints [POST]
func SaveAiResponseToDB(painPointRepo repository.Repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var apiResponseData struct {
			Role string `json:"role"`
		}

		// Bind the incoming request body (which contains the AI response)
		if err := ctx.ShouldBindJSON(&apiResponseData); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		modelUrl := os.Getenv("MODELURI")
		var modelConfig models.ModelConfig
		var message models.Message
		message.Role = "system"
		message.Content = `You are an expert assistant representing Initializ.ai, a platform specializing in developing, securing, and operating cloud-native and AI applications.\r\n \r\nInitializ.ai Overview:\r\n \r\n1. Unified Platform for GenAI & Cloud-Native Apps: Simplifies lifecycle management for modern apps.\r\n \r\n \r\n2. Security Features: Secure container building, continuous vulnerability scanning, AI-driven threat management.\r\n \r\n \r\n3. Deployment Capabilities: Instant deployments, automated pipelines, polyglot support (Python, NodeJS, Java, Go, .NET).\r\n \r\n \r\n4. AI Augmented Development: AI-enhanced environments, seamless tooling integration, deployment workflows.\r\n \r\n \r\n5. Private AI Services: Supports models like Llama, Whisper, and Stable Diffusion; GPU\/CPU cost efficiency.\r\n \r\n \r\n6. Observability & AI-Ops: Centralized logs, anomaly detection, predictive analytics, auto performance improvement.\r\n \r\n \r\n7. Kubernetes Simplification: Simplified Kubernetes management.\r\n \r\n \r\n8. Collaboration & Reporting: Advanced reporting, alerts, notifications, and forecasting.\r\n \r\n \r\n \r\nTask:\r\n \r\nWhen provided with a job title (e.g., Software Development Engineer, Project Manager) and Initializ.ai's values, your task is to:\r\n \r\n1. Identify key pain points for the given role.\r\n \r\n \r\n2. Explain how Initializ.ai addresses these pain points using its features.\r\n \r\n \r\n3. Integrate Initializ.ai's values\u2014simplification, security, innovation, and collaboration\u2014into the response.\r\n \r\n \r\n4. Provide a clear, professional response in 50 words or less.\r\n \r\n \r\n \r\nResponse Structure:\r\n \r\n1. Pain Points: List 1-2 challenges for the role.\r\n \r\n \r\n2. How Initializ Helps: Highlight 1-2 relevant Initializ.ai features.\r\n \r\n \r\n \r\nStay concise and ensure the response remains 50 words or less.\r\n\r\n#Guardrails\r\n1. Do  no generate any unwanted data\r\n2. Do no provide the the words count in response`
		modelConfig.Messages = append(modelConfig.Messages, message)
		message.Role = "user"
		message.Content = apiResponseData.Role
		modelConfig.Messages = append(modelConfig.Messages, message)
		modelConfig.Model = "meta-llama/Meta-Llama-3.1-8B-Instruct"
		modelConfig.Stream = false
		modelConfig.Temperature = 0.7
		modelConfig.MaxTokens = 5000

		modelBody, _ := json.Marshal(modelConfig)

		req, err := http.NewRequest("POST", modelUrl, bytes.NewBuffer(modelBody))
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
		resp, err := http.DefaultClient.Do(req)
		if http.StatusOK != resp.StatusCode {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		defer resp.Body.Close()
		var ModelResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		resbody, err := io.ReadAll(resp.Body)
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occurred while reading the response body.", nil)
			return
		}

		err = json.Unmarshal(resbody, &ModelResponse)
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occurred while reading the response body.", nil)
			return
		}
		reserachedData := ModelResponse.Choices[0].Message.Content
		defer resp.Body.Close()
		painPoints, valueProposition := splitContent(reserachedData)
		// Save the content to the database
		err = SavePainPoints(painPointRepo, painPoints, valueProposition, apiResponseData.Role)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error saving pain point to the database: %v", err)})
			return
		}

		ctx.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully saved the AI response as a pain point",
		})
	}
}
func splitContent(content string) (string, string) {
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "How Initializ Helps:", "How Initializ.ai Helps:")
	parts := strings.Split(content, "How Initializ.ai Helps:")
	var painPoints string
	var valueProposition string
	if len(parts) > 1 {
		painPoints = parts[0]
		valueProposition = parts[1]
	} else {
		painPoints = content
		valueProposition = "" // Empty value if not found
	}
	return painPoints, valueProposition
}

// SavePainPoints saves the generated content as a pain point in the database
func SavePainPoints(painPointRepo repository.Repository, painPoints, valueProposition, role string) error {
	painPoint := models.PainPointModel{
		Role:             role,
		PainPoint:        painPoints,
		ValueProposition: valueProposition,
	}

	// Insert the pain point into the database
	_, err := painPointRepo.InsertOne(painPoint)
	if err != nil {
		return fmt.Errorf("error saving pain point to the database: %v", err)
	}
	return nil
}

// DeleteCaseStudy				godoc
// @Tags					Pain Points Apis
// @Summary					Delete Pain Points and Value Proposition by ID
// @Description				Delete Pain Points and Value Proposition by ID
// @Param                    id   path string true "Pain Points ID"
// @Success					200 {object} responses.ApplicationResponse{}
// @Failure					404 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/painpoints/{id} [DELETE]
func DeletePainPoints(PainPointRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		painPointID := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(painPointID)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid Pain Points ID format.",
			})
			return
		}
		filter := bson.M{"_id": objectID}
		PainPointRepo.DeleteMany(filter)
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Pain Points deleted successfully",
		})
	}
}
