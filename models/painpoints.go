package models

import "time"

type PainPointModel struct {
	ID               string    `bson:"_id,omitempty" json:"id"`
	Role             string    `json:"role" bson:"role"`
	PainPoint        string    `json:"pain_points" bson:"pain_points"`
	ValueProposition string    `json:"value_proposition" bson:"value_proposition"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
}

type PainPointRole struct {
	Role string `json:"role" bson:"role"`
}
