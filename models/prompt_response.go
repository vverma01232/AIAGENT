package models

type Prompts struct {
	ID     string `bson:"_id,omitempty" json:"id"`
	Name   string `bson:"name" json:"name"`
	Prompt string `bson:"prompt,omitempty" json:"prompt,omitempty"`
}

type UserDetails struct {
	ID                 string `bson:"_id,omitempty" json:"id,omitempty"`
	FirstName          string `bson:"first_name" json:"first_name"`
	LastName           string `bson:"last_name" json:"last_name"`
	Email              string `bson:"email" json:"email"`
	CompanyDetails     string `bson:"company" json:"company"`
	LinkedInProfileUrl string `bson:"linkedin_url" json:"linkedin_url"`
}

type GenerateAIBody struct {
	SystemPrompt string `bson:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Linkedin_url string `bson:"linkedin_url,omitempty" json:"linkedin_url,omitempty"`
	CompanyUrl   string `bson:"company_url,omitempty" json:"company_url,omitempty"`
	Stream       bool   `bson:"stream,omitempty" json:"stream,omitempty"`
	Task         string `bson:"task,omitempty" json:"task,omitempty"`
	TODOResearch bool   `bson:"to_do_research,omitempty" json:"to_do_research,omitempty"`
}
