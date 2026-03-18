package skill

import "errors"

var (
	ErrSkillNotFound        = errors.New("skill not found")
	ErrNoSkillsFound        = errors.New("no skills found")
	ErrInvalidSkillDocument = errors.New("invalid skill document")
)

type Skill struct {
	ID             string   `json:"id" yaml:"id"`
	Name           string   `json:"name" yaml:"name"`
	Description    string   `json:"description" yaml:"description"`
	Tags           []string `json:"tags" yaml:"tags"`
	Prompt         string   `json:"prompt" yaml:"prompt"`
	SuggestedTools []string `json:"suggested_tools" yaml:"suggestedTools"`
}
