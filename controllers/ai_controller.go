package controllers

import (
	"aiagent/models"
	"aiagent/responses"
	"aiagent/services"
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// GenerateAI				godoc
// @Tags					AIAgent Apis
// @Summary					Generate with AI
// @Description				Generate with AI
// @Param					GenerateAI body models.GenerateAIBody true "Generate Body Response"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/generatewithAI [POST]
func GeneratewithAIHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body models.GenerateAIBody
		ctx.BindJSON(&body)

		if body.TODOResearch {
			scrapedData, err := services.ScrapeData(body.Linkedin_url)
			if err != nil {
				ReturnResponse(ctx, http.StatusBadRequest, "Error occurred while scraping the URL: "+err.Error(), nil)
				return
			}

			// Replace the placeholder with the scraped data
			body.Task = strings.ReplaceAll(body.Task, "**research**", scrapedData)
		}
		modelUri := os.Getenv("MODELURI")
		var modelConfig models.ModelConfig
		var message models.Message
		message.Role = "system"
		message.Content = body.SystemPrompt
		modelConfig.Messages = append(modelConfig.Messages, message)
		message.Role = "user"
		message.Content = body.Task
		modelConfig.Messages = append(modelConfig.Messages, message)
		modelConfig.Model = "meta-llama/Meta-Llama-3.1-8B-Instruct"
		modelConfig.Stream = body.Stream
		modelConfig.Temperature = 0.7
		modelConfig.MaxTokens = 5000

		modelBody, _ := json.Marshal(modelConfig)

		req, err := http.NewRequest("POST", modelUri, bytes.NewBuffer(modelBody))
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		req.Header.Set("Authorization", "Bearer "+os.Getenv("TOKEN"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		defer resp.Body.Close()

		// Check for streaming
		if modelConfig.Stream {
			ctx.Header("Content-Type", "text/event-stream")
			// Stream the response from Model API to the client
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				chunk := scanner.Text()
				ctx.Writer.WriteString(chunk + "\n")
				ctx.Writer.Flush()
			}
		} else {
			// Non-streaming: Accumulate the response and return once fully received
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				ReturnResponse(ctx, http.StatusBadRequest, "Error occured while generating the response.", nil)
				return
			}
			// Return the full response in JSON format
			var aiResponse struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}

			// Unmarshal the response body into the structure
			if err := json.Unmarshal(bodyBytes, &aiResponse); err != nil {
				ReturnResponse(ctx, http.StatusBadRequest, "Error occurred while parsing the AI response", nil)
				return
			}

			// Return the full response in JSON format
			ctx.Header("Content-Type", "application/json")
			ReturnResponse(ctx, http.StatusOK, "Successfully generate the response", string(bodyBytes))
		}
	}
}

func ReturnResponse(ctx *gin.Context, statusCode int, message string, data interface{}) {
	if data == nil {
		ctx.JSON(statusCode, responses.ApplicationResponse{
			Status:  statusCode,
			Message: message,
		})
	} else {
		ctx.JSON(statusCode, responses.ApplicationResponse{
			Status:  statusCode,
			Message: message,
			Data:    data,
		})
	}
}
