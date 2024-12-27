package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"aiagent/services"
	"context"
	"net/http"

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
		if err := c.BindJSON(&body); err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Invalid input", nil)
			return
		}
		// Call the scrapeData function to get the scraped content
		researchedData, err := services.ScrapeData(body.URL)
		if err != nil {
			ReturnResponse(c, http.StatusBadRequest, "Error occurred while scraping the data", nil)
			return
		}
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
