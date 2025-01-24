package models

import (
	"time"
)

type Prompts struct {
	ID         string    `bson:"_id,omitempty" json:"id"`
	Name       string    `bson:"name" json:"name"`
	Prompt     string    `bson:"prompt,omitempty" json:"prompt,omitempty"`
	Purpose    string    `bson:"purpose,omitempty" json:"purpose,omitempty"`
	PromptRule string    `bson:"prompt_rule,omitempty" json:"prompt_rule,omitempty"`
	CreatedBy  string    `bson:"created_by,omitempty" json:"created_by,omitempty"`
	UpdatedAt  time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	CreatedAt  time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedBy  string    `bson:"updated_by,omitempty" json:"updated_by,omitempty"`
}

type UserDetails struct {
	Name                  string       `bson:"name" json:"name"`
	Experience            string       `bson:"experience" json:"experience"`
	Location              string       `bson:"location" json:"location"`
	MobileNo              string       `bson:"mob_no" json:"mob_no"`
	Email                 string       `bson:"email" json:"email"`
	CompanyDetails        string       `bson:"company" json:"company"`
	Designation           string       `bson:"designation" json:"designation"`
	LinkedInProfileUrl    string       `bson:"linkedin_url" json:"linkedin_url"`
	LinkedInProfileData   string       `bson:"linkedIn_data" json:"linkedIn_data"`
	CompanyResearchedData string       `bson:"company_data" json:"company_data"`
	CompanyWebsite        string       `json:"company_website" bson:"company_website"`
	AiOutput              UserAiOutput `bson:"ai_output" json:"ai_output"`
}

type GenerateAIBody struct {
	SystemPrompt string `bson:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Linkedin_url string `bson:"linkedin_url,omitempty" json:"linkedin_url,omitempty"`
	CompanyUrl   string `bson:"company_url,omitempty" json:"company_url,omitempty"`
	Stream       bool   `bson:"stream,omitempty" json:"stream,omitempty"`
	Task         string `bson:"task,omitempty" json:"task,omitempty"`
	TODOResearch bool   `bson:"to_do_research,omitempty" json:"to_do_research,omitempty"`
}
