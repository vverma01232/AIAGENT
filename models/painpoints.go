package models

type PainPointModel struct {
	Role      string `json:"role" bson:"role"`
	PainPoint string `json:"pain_points" bson:"pain_points"`
}
