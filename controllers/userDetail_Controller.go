package controllers

import (
	"aiagent/models"
	"aiagent/repository"
	"aiagent/responses"
	"aiagent/services"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"net/http"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UploadExcel				godoc
// @Tags					UserData Apis
// @Summary					Upload Excel File
// @Description				Upload Excel File in Base 64 format
// @Param metadata body models.UploadRequest true "File metadata"
// @Produce					application/json
// @Success					200 {object} responses.ApplicationResponse{}
// @Router					/initializ/v1/ai/upload [POST]
func UploadExcel(userDataRepo repository.Repository, promptRepo repository.Repository, painPointRepo repository.Repository) gin.HandlerFunc {
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
		excel, err := excelize.OpenReader(bytes.NewReader(data))
		if err != nil {
			log.Error("Failed to Read Excel Sheet", err)
			ctx.JSON(http.StatusBadRequest, responses.ApplicationResponse{
				Status:  http.StatusBadRequest,
				Message: err.Error(),
			})
			return
		}
		// Read the header row to get the column names
		rows := excel.GetRows("Sheet1")
		headerRow := rows[0]
		headerMap := make(map[string]int)
		for idx, header := range headerRow {
			headerMap[strings.ToLower(strings.TrimSpace(header))] = idx
		}
		for i, row := range excel.GetRows("Sheet1") {
			if i == 0 {
				continue
			} else {
				// Prepare the user details
				user := models.UserDetails{}

				// Dynamically map columns to UserDetails struct fields
				if index, exists := headerMap["name"]; exists {
					user.Name = row[index]
				}
				if index, exists := headerMap["experience"]; exists {
					user.Experience = row[index]
				}
				if index, exists := headerMap["location"]; exists {
					user.Location = row[index]
				}
				if index, exists := headerMap["mobile no"]; exists {
					user.MobileNo = row[index]
				}
				if index, exists := headerMap["email"]; exists {
					user.Email = row[index]
				}
				if index, exists := headerMap["designation"]; exists {
					user.Designation = row[index]
				}
				if index, exists := headerMap["company"]; exists {
					user.CompanyDetails = row[index]
				}
				if index, exists := headerMap["linkedin url"]; exists {
					user.LinkedInProfileUrl = row[index]
				}
				if index, exists := headerMap["company url"]; exists {
					user.CompanyWebsite = row[index]
				}
				var wg sync.WaitGroup
				wg.Add(1)
				go func(user *models.UserDetails) {
					defer wg.Done()
					linkedinData, err := services.ScrapeData(user.LinkedInProfileUrl)
					if err != nil {
						log.Warn("Error fetching LinkedIn data for", user.Name, ":", err)
					} else {
						user.LinkedInProfileData = linkedinData
					}
				}(&user)
				wg.Add(1)
				go func(user *models.UserDetails) {
					defer wg.Done()
					var companyUrl string

					// If the company website exists, use it
					if len(user.CompanyWebsite) > 0 {
						companyUrl = user.CompanyWebsite
					} else {
						// Otherwise, construct the URL using the email domain
						parts := strings.Split(user.Email, "@")
						if len(parts) > 1 {
							companyUrl = "https://www." + parts[1]
						}
					}

					// Scrape company data using the URL
					if len(companyUrl) > 0 {
						companyDescription, err := services.ScrapeData(companyUrl)
						if err != nil {
							log.Warn("Error fetching company data for", user.CompanyDetails, ":", err)
						} else {
							user.CompanyResearchedData = companyDescription
						}
					}
				}(&user)

				// Fetch prompts from the database
				prompts, err := fetchPrompts()
				if err != nil {
					log.Error("Error fetching prompts:", err)
					ctx.JSON(http.StatusInternalServerError, responses.ApplicationResponse{
						Status:  http.StatusInternalServerError,
						Message: "Error fetching prompts from the database",
					})
					return
				}
				// Generate AI Output for the user
				userAiOutput := generateAiOutput(user, prompts, painPointRepo)
				user.AiOutput = userAiOutput

				_, err = userDataRepo.InsertOne(user)
				if err != nil {
					log.Error("Error occurred while inserting user data:", err)
				}
			}
		}

		log.Info("Data uploaded successfully")
		ctx.JSON(http.StatusOK, responses.ApplicationResponse{
			Status:  http.StatusOK,
			Message: "Data uploaded and AI output generated successfully",
		})
	}
}

// fetchPrompts retrieves the saved prompts from the database for AI generation
func fetchPrompts() (map[string]models.Prompts, error) {

	promptMap := map[string]models.Prompts{
		"Cold Calls": {
			Prompt:     "Create a brief, natural-sounding icebreaker for a cold call to **first_name**, **title** at **company** . Use the provided research to inform your approach, focusing on a relevant pain point that our service can address. The goal is to sound human and conversational while still being direct about the purpose of the call. ---Start of Research Information--- **AI Research** (Managed By Initializ) ---End of Research Information--- Guidelines: Start with a brief, friendly greeting. Mention your **sender_name** and **sender_company**. Ask if they have a moment to talk about a specific pain point or challenge related to their role or industry. The pain point should be directly related to a service or solution your company offers. Keep it brief - aim for 2-3 sentences maximum. Use natural language and avoid jargon or overly formal phrasing. Be prepared to elaborate on the pain point if given permission to continue. Example format (but feel free to vary): 'Hi **first_name** , this is **sender_first_name** from **sender_company** . Do you have a quick moment to discuss [specific pain point related to prospect's role or recent company development]?' If given permission to continue: Briefly elaborate on the pain point, relating it to the prospect's specific situation or a recent industry trend. Then, ask an open-ended question to encourage dialogue. Remember, the goal is to quickly establish relevance and open a conversation about how your service can address their specific challenges.",
			PromptRule: "Use ONLY information explicitly stated in the provided research. Do not add any details or make any inferences not directly supported by the research.Only output the script not any descriptive headings. Do not output quotation/speech marks. If the research doesn't provide enough information for a specific point, use a phrase like 'Based on the information available to me...' and stick to what you know for certain.Keep the entire icebreaker under 20 seconds when spoken aloud. Do not use industry jargon unless it's specifically mentioned in the research as relevant to this prospect. Be prepared to say 'I don't have enough information about that' if asked about something not covered in the research. Do not attempt to fill in gaps in the research with assumptions or generalizations.If referencing any statistics or specific claims, only use those explicitly stated in the research.The open-ended question must be directly related to information provided in the research.If the research doesn't provide a clear pain point or value proposition, default to a more general, research-based question about their role or industry.",
		},
		"AI Research": {
			Prompt:     "[Here is your task]:You are an experienced Sales Development Representative (SDR) at **sender_company**. Your goal is to research and create a personalized outreach strategy for **first_name** , a **title** at **company**. Use the information provided to craft a detailed, relevant summary that will help engage this prospect effectively. Your analysis should be insightful, demonstrating a deep understanding of both **sender_company**'s offerings and the prospect's potential needs.Analyze Context: Briefly summarize **first_name** s role as **title** at **company** , including industry and potential. When describing **first_name**  current role and activities, ensure you are referencing their most recent active experience as listed on their LinkedIn profile, which should be indicated by a date range ending with 'present' . Do not use information from older positions unless explicitly relevant to the current analysisIdentify Key Challenges: List 3 challenges **company** likely faces, based on our value propositions for **title** below focusing on areas **sender_company**  can address, based on our value propositions below. Focus on challenges specific to **first_name**  role as **title** , using the provided source data to identify role-specific priorities & symptoms of challenges. Ensure these solely align with the value propositions below.Present **sender_company**  Solutions: For each challenge, explain how **sender_company**  a. Addresses the specific need challenges b. Highlights a benefit to company   c. Explains the benefit tocompany and **first_name**'s role. For each solution, provide hyper-specific language that demonstrate how **sender_company**  can improve an outcome for **company** (ensure this is completely factual). Use words not numbers to communicate this.Provide Concrete Example: Give one specific example of how **sender_company** could solve a unique challenge for company , based on their industry or structure. Ensure this example uses language and metrics highly specific to **first_name**'s role and industry, avoiding generic AI buzzwords.Recent Company News:Identify a recent newsworthy event or development specific to company  or **first_name**'s role. Ensure the news is from the last 6 months only. Briefly explain how this event might relate to the challenges or priorities identified earlier.[Use the following information as sources]:Linkedin profile: '**linkedin_profile** .'company website data: '**company_website_data** '**sender_company** value propositions here: '**sender_value_propositions** .",
			PromptRule: "You are a top marketing/sales agent with outstanding account research and email writing skills. Your attention to detail and communication expertise drive excellent results and strong client relationships. You are adaptable, empathetic, and relentlessly goal-oriented.Ensure the google news used it from the last 3 months only. Use language that resonates with first_name  based on their priorities.Ensure every point references how it benefits company , linked to the client types.Avoid generic language and provide specific, personalized details. Do not format with any * or # ",
		},
		"Question Based Email": {
			Prompt:     "As a representative from **sender_company** , craft a highly personalized email to **first_name** , **title**  at company. Utilize the provided research information, including biometrics, to identify top priorities, challenges, and relevant KPIs specific to **first_name**'s role and industry.Critical Rules:Strictly output in **language** language.Use a **tone** toneThe length of the email should be maximum **length** words[RESEARCH INFORMATION]: 'AI Research (Managed By Initializ)' [/ RESEARCH INFORMATION]Format:Greet with their first name - **first_name** Open with an observation or news hook directly relevant to company or **first_name**'s current situation. (Naturalize the language)Transition into a thought-provoking question that connects your opening to a specific challenge or priority you've identified for **first_name**'s role.Present a hyper-specific value proposition addressing this challenge. Use role-specific language and metrics to clearly demonstrate how **sender_company** can measurably improve a key metric or outcome for company.Craft a call-to-action focused on how **sender_company**  can help improve **first_name**'s current process related to the challenge discussed.Sign off professionally with - **sender_first_name** P.S. Include a brief, personalized comment referencing **first_name**  and an insight from your research, with a subtle touch of humor. DO NOT talk about location. Limit to one sentence.",
			PromptRule: "Tone: Informal, conversational, and non-salesy.Length: Maximum 100 words, preferably under 90.Personalization: Ensure all content is highly relevant and tailored to first_name's specific role, industry, and current situation. Demonstrate a deep understanding of their challenges and priorities.Language:Prioritize 'you' language to focus on the prospect.Use language and metrics hyper-specific to first_name 's job function and challenges.Avoid generic AI buzzwords, overly technical jargon, and generic industry trends.Content:Focus more on the prospect's company than on sender_company.Avoid phrases like 'At company' or 'I hope this message finds you well.'Don't use flattery or over-complimentary language (e.g., 'truly impressive,' 'truly remarkable').Omit any references to working with similar brands or social proof.Structure:Use line breaks between sentences for readability.Don't use company name suffixes (LTD, PLC, INC).Do not:Describe your own feelings. Instead, provide a descriptive perspective.Offer invitations (e.g., for drinks) in the P.S. line.Mention the weather.List sources or reference the research process.Demonstrate a nuanced understanding of first_name's current processes and how sender_company  can improve them.Do not write a greetingUse more 'you' language.Never say At company Reference more about the company than us.Put a lot of whitespace between each sentence, which is a line gap, so it looks spaced outDo not put any company name suffixes like LTD, PLC, INC, you are writing an email to first_name who works at company . The email should be maximum 120 words. Under 100 words is preferable.Never say - I hope this message finds you well.Do not list sources. Use social intelligence to write a professional and succinct email. Do not describe how you feel. Instead provide a descriptive perspective. For example, avoid flattery and being over-complimentary. For example, 'truly impressive', 'truly remarkable', 'truly game-changer', 'truly inspiring', and similar should be completely avoided. ",
		},
	}
	return promptMap, nil
}

// generates AI outputs for Cold Calls, AI Research, and Question-Based Email using fetched prompts
func generateAiOutput(user models.UserDetails, prompts map[string]models.Prompts, painPointRepo repository.Repository) models.UserAiOutput {
	// Replace placeholders in the prompts
	coldCallOutput, _ := performResearchUsingPrompt(replacePlaceholders(prompts["Cold Calls"].Prompt, user, painPointRepo), prompts["Cold Calls"].PromptRule)
	aiResearchOutput, _ := performResearchUsingPrompt(replacePlaceholders(prompts["AI Research"].Prompt, user, painPointRepo), prompts["AI Research"].PromptRule)
	questionBasedEmailOutput, _ := performResearchUsingPrompt(replacePlaceholders(prompts["Question Based Email"].Prompt, user, painPointRepo), prompts["Question Based Email"].PromptRule)

	return models.UserAiOutput{
		ColdCalls: models.AiGenerated{
			AiGeneratedOutpt: coldCallOutput,
			GeneratedAt:      time.Now(),
		},
		AiResearch: models.AiGenerated{
			AiGeneratedOutpt: aiResearchOutput,
			GeneratedAt:      time.Now(),
		},
		QuestionBasedEmail: models.AiGenerated{
			AiGeneratedOutpt: questionBasedEmailOutput,
			GeneratedAt:      time.Now(),
		},
	}
}

// replacing placeholders in the prompt with actual user data
func replacePlaceholders(prompt string, user models.UserDetails, painPointRepo repository.Repository) string {
	firstName := ""
	if len(user.Name) > 0 {
		parts := strings.Fields(user.Name)
		if len(parts) > 0 {
			firstName = parts[0]
		}
	}

	valueProposition, err := GetPainPointsForRole(painPointRepo, user.Designation)
	if err != nil {
		return "At Initializ.ai, we provide a unified platform designed to streamline and simplify the entire lifecycle of cloud-native and AI applications. Our solutions address the complexity of managing modern application infrastructure while enhancing security, deployment efficiency, and developer productivity. Whether you're looking to build, secure, deploy, or optimize your applications, Initializ.ai offers an all-in-one platform that reduces operational overhead and accelerates innovation."
	}

	// Replace placeholders in the prompt
	prompt = strings.Replace(prompt, "**first_name**", firstName, -1)
	prompt = strings.Replace(prompt, "**title**", user.Designation, -1)
	prompt = strings.Replace(prompt, "**Name**", user.Name, -1)
	prompt = strings.Replace(prompt, "**Experience**", user.Experience, -1)
	prompt = strings.Replace(prompt, "**Location**", user.Location, -1)
	prompt = strings.Replace(prompt, "**company**", user.CompanyDetails, -1)
	prompt = strings.Replace(prompt, "**linkedin_profile**", user.LinkedInProfileUrl, -1)
	prompt = strings.Replace(prompt, "**company_website_data**", user.CompanyResearchedData, -1)
	prompt = strings.Replace(prompt, "**sender_value_propositions**", valueProposition, -1)
	prompt = strings.Replace(prompt, "**AI Research**", user.AiOutput.AiResearch.AiGeneratedOutpt, -1)
	prompt = strings.Replace(prompt, "**language**", "English", -1)
	prompt = strings.Replace(prompt, "**tone**", "Conversational", -1)
	prompt = strings.Replace(prompt, "**sender_company**", "initializ.ai", -1)
	prompt = strings.Replace(prompt, "**sender_first_name**", "Yash", -1)

	return prompt
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
		findOptions := options.Find().SetSort(bson.D{{Key: "ai_output.coldcalls.generatedat", Value: -1}})
		cursor, err := userDataRepo.FindWithOption(bson.M{}, findOptions)
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
func performResearchUsingPrompt(prompt string, promptRule string) (string, error) {
	modelUri := os.Getenv("MODELURI")
	apiToken := os.Getenv("TOKEN")
	if apiToken == "" {
		return "", fmt.Errorf("bearer token not found. Please set the API token")
	}

	modelConfig := models.ModelConfig{
		Model:     "meta-llama/Meta-Llama-3.1-8B-Instruct",
		MaxTokens: 5000,
		Stream:    false,
		Messages: []models.Message{
			{Role: "system", Content: promptRule},
			{Role: "user", Content: prompt},
		},
	}

	jsonConfig, err := json.Marshal(modelConfig)
	if err != nil {
		return "", fmt.Errorf("error marshaling config: %v", err)
	}

	// Prepare the HTTP request with Bearer token
	req, err := http.NewRequest("POST", modelUri, bytes.NewBuffer(jsonConfig))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Handle unsuccessful responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse and extract the response content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Define the structured response format to capture the entire AI response
	var aiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	// Unmarshal the body into the structured response
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Check if we have any valid content in the response
	if len(aiResponse.Choices) > 0 && aiResponse.Choices[0].Message.Content != "" {
		return aiResponse.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response content found")
}

func GetPainPointsForRole(painPointRepo repository.Repository, role string) (string, error) {
	var painPoint models.PainPointModel
	err := painPointRepo.FindOne(bson.M{"role": role}).Decode(&painPoint)
	if err != nil {
		return "", fmt.Errorf("error fetching pain points for role %s: %v", role, err)
	}
	return painPoint.ValueProposition, nil
}
