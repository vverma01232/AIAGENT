package models

type PainPointModel struct {
	Role             string `json:"role" bson:"role"`
	PainPoint        string `json:"pain_points" bson:"pain_points"`
	ValuePreposition string `json:"value_preposition" bson:"value_preposition"`
}
