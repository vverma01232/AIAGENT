package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"net/http"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// UploadExcel				godoc
// @Tags					UserData Apis
// @Summary					Upload Excel File
// @Description				Upload Excel File
// @Param					UploadExcel body models.UploadRequest true "File Data in base64 encoded"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/upload [POST]
func UploadExcel(userDataRepo repository.Repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.UploadRequest
		if err := ctx.BindJSON(&req); err != nil {
			log.Error("Error binding JSON:", err)
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		data, err := base64.StdEncoding.DecodeString(req.FileData)
		if err != nil {
			log.Error("Error in decoding:", err)
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error in decoding file :" + err.Error(),
			})
			return
		}
		excel, err := excelize.OpenReader(strings.NewReader(string(data)))
		if err != nil {
			log.Error("Failed to Read Excel Sheet", err)
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}

		var userList []models.UserDetails
		for i, row := range excel.GetRows("Sheet1") {
			if i == 0 {
				continue
			} else {
				user := models.UserDetails{
					Name:               row[1],
					Experience:         row[2],
					Location:           row[3],
					MobileNo:           row[4],
					Email:              row[5],
					Designation:        row[6],
					CompanyDetails:     row[7],
					LinkedInProfileUrl: row[8],
				}

				linkedinData, err := ScrapeLinkedInProfile(user.LinkedInProfileUrl)
				if err != nil {
					log.Warn("Error fetching LinkedIn data for", user.Name, ":", err)
				} else {
					user.LinkedInProfileData = linkedinData // Store scraped LinkedIn data
				}
				parts := strings.Split(user.Email, "@")
				if len(parts) > 1 {
					// Assuming the domain part is a valid company URL
					companyUrl := "https://www." + parts[1]
					companyDescription, err := scrapeCompanyData(companyUrl)
					if err != nil {
						log.Warn("Error fetching company data for", user.CompanyDetails, ":", err)
					} else {
						user.CompanyResearchedData = companyDescription
					}
				}
				userList = append(userList, user)
			}
		}

		data, _ = json.Marshal(userList)

		var interfaceData []interface{}
		json.Unmarshal(data, &interfaceData)
		_, err = userDataRepo.InsertMany(interfaceData, nil)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occured while uploading the data : " + err.Error(),
			})
			return
		}
		log.Info("Data added successfully")
		ctx.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Data uploaded successfully",
		})
	}
}

// GetAllUserData			godoc
// @Tags					UserData Apis
// @Summary					Get User Data
// @Description				Get all Data
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/allusers [GET]
func GetAllUserData(userDataRepo repository.Repository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cursor, err := userDataRepo.Find(bson.M{})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occured while fetching the data from db : " + err.Error(),
			})
			return
		}
		defer cursor.Close(context.TODO())

		var userData []models.UserDetails
		err = cursor.All(context.TODO(), &userData)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: "Error occured while fetching the data from db : " + err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Successfully fetched the user data",
			Data:    userData,
		})
	}
}
func scrapeCompanyData(companyUrl string) (string, error) {
	scrapperUrl := os.Getenv("SCRAPPERURI")
	scrapeBody := map[string]string{
		"url": companyUrl,
	}
	reqbodyBytes, err := json.Marshal(scrapeBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	req, err := http.NewRequest("POST", scrapperUrl+"/scrape", bytes.NewBuffer(reqbodyBytes))
	if err != nil {
		return "", fmt.Errorf("error occurred while making the scrape request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error occurred while generating the response: %w", err)
	}
	defer resp.Body.Close()
	var scrapeResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	scrapeResbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error occurred while reading the response body: %w", err)
	}
	err = json.Unmarshal(scrapeResbody, &scrapeResponse)
	if err != nil {
		return "", fmt.Errorf("error occurred while unmarshaling the response body: %w", err)
	}
	if len(scrapeResponse.Choices) > 0 {
		return scrapeResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content found in the scraper response")
}

func ScrapeLinkedInProfile(linkedinURL string) (string, error) {
	scrapperUrl := os.Getenv("SCRAPPERURI")
	requestBody := map[string]interface{}{
		"url": linkedinURL,
	}

	// Marshal request body to JSON
	reqBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Send POST request to scrapper URI
	req, err := http.NewRequest("POST", scrapperUrl+"/scrape", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("error occurred while making the scrape request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error occurred while generating the response: %w", err)
	}
	defer resp.Body.Close()

	// Read and unmarshal the response
	var scrapeResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"` // Scraped data
			} `json:"message"`
		} `json:"choices"`
	}

	scrapeResBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error occurred while reading the response body: %w", err)
	}

	err = json.Unmarshal(scrapeResBody, &scrapeResponse)
	if err != nil {
		return "", fmt.Errorf("error occurred while unmarshaling the response body: %w", err)
	}

	// Check if any scraped content is available
	if len(scrapeResponse.Choices) > 0 {
		return scrapeResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content found in the scraper response")
}
