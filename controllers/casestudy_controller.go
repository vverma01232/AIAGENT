package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaveCaseStudy            	godoc
// @Tags					Case Study Apis
// @Summary					Save Case Study
// @Description				Save Case Study
// @Produce					application/json
// @Param                    request body models.Casestudy true  "Case Study"
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/casestudy [POST]
func SaveCaseStudy(caseStudyRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.CaseStudy
		c.BindJSON(&body)
		scrapperUrl := os.Getenv("SCRAPPERURI")
		scrapperBody := map[string]string{
			"url": body.URL,
		}
		reqbodyBytes, _ := json.Marshal(scrapperBody)
		req, err := http.NewRequest("POST", scrapperUrl+"/scrape", bytes.NewBuffer(reqbodyBytes))
		if err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Error", nil)
			return
		}
		req.Header.Set("Content-type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Error occured while generating the response.", nil)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			ReturnResponse(c, http.StatusBadRequest, "Error occured while reseraching.", err)
			return
		}
		var ScrapeResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		scrapeResbody, err := io.ReadAll(resp.Body)
		if err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Error occurred while reading the response body.", nil)
			return
		}

		err = json.Unmarshal(scrapeResbody, &ScrapeResponse)
		if err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Error occurred while reading the response body.", nil)
			return
		}
		researchedData := ScrapeResponse.Choices[0].Message.Content
		caseStudy := models.CaseStudy{
			URL:            body.URL,
			ResearchedData: researchedData,
		}
		if _, err := caseStudyRepo.InsertOne(caseStudy); err != nil {
			ReturnResponse(c, http.StatusInternalServerError, "Error saving case study to the database", nil)
			return
		}
		ReturnResponse(c, http.StatusOK, "Scraped data saved successfully", researchedData)
	}
}

// GetCaseStudy				godoc
// @Tags					Case Study Apis
// @Summary					Get Case Study
// @Description				Get Case Study Api
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/casestudy [GET]
func GetCaseStudy(caseStudyRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := caseStudyRepo.Find(bson.M{})
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occured while fetching the data from db : " + err.Error(),
			})
			return
		}
		defer cursor.Close(context.TODO())

		var caseStudy []models.CaseStudy
		err = cursor.All(context.TODO(), &caseStudy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error occured while fetching the data from db : " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully fetched the prompts",
			Data:    caseStudy,
		})
	}
}

// DeleteCaseStudy				godoc
// @Tags					Case Study Apis
// @Summary					Delete Case Study by ID
// @Description				Delete Case Study by ID
// @Param                    id   path string true "Case Study ID"
// @Success					200 {object} responses.ApplicationResponse{}
// @Failure					404 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/casestudy/{id} [DELETE]
func DeleteCaseStudy(caseStudyRepo repository.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		caseStudyID := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(caseStudyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Invalid case study ID format.",
			})
			return
		}
		filter := bson.M{"_id": objectID}
		caseStudyRepo.DeleteMany(filter)
		c.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Case study deleted successfully",
		})
	}
}
