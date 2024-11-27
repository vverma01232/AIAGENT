package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"context"
	"encoding/base64"
	"encoding/json"

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
		for _, row := range excel.GetRows("Sheet1") {
			if len(row) >= 6 {
				var user models.UserDetails
				user.FirstName = row[0]
				user.LastName = row[1]
				user.Email = row[2]
				user.CompanyDetails = row[3]
				user.LinkedInProfileUrl = row[4]
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
