package models

type CaseStudy struct {
	URL         string `json:"url" bson:"url"`
	ScrapedData string `json:"scraped_data" bson:"scraped_data"`
}

type Casestudy struct {
	URL string `json:"url" bson:"url"`
}
