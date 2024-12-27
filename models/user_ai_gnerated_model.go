package models

import (
	"time"
)

type UserAiOutput struct {
	ColdCalls          AiGenerated
	AiResearch         AiGenerated
	QuestionBasedEmail AiGenerated
}
type AiGenerated struct {
	AiGeneratedOutpt string
	GeneratedAt      time.Time
}
