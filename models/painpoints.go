package models

type PainPointModel struct {
	Role             string `json:"role" bson:"role"`
	PainPoint        string `json:"pain_points" bson:"pain_points"`
	ValueProposition string `json:"value_proposition" bson:"value_proposition"`
}

type PainPointRole struct {
	Role string `json:"role" bson:"role"`
}
