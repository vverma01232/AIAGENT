package models

type CaseStudy struct {
	ID             string `bson:"_id,omitempty" json:"id"`
	URL            string `json:"url" bson:"url"`
	ResearchedData string `json:"researched_data" bson:"researched_data"`
}

type Casestudy struct {
	URL string `json:"url" bson:"url"`
}
