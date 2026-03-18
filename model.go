package grimoire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flarexio/grimoire/vector"
)

var (
	ErrSkillNotFound        = errors.New("skill not found")
	ErrNoSkillsFound        = errors.New("no skills found")
	ErrVectorDBNotSet       = errors.New("vector database not set")
	ErrInvalidSkillDocument = errors.New("invalid skill document")
)

type Config struct {
	SkillsDir string        `yaml:"skillsDir"`
	Vector    vector.Config `yaml:"vector"`
}

type Skill struct {
	ID             string   `json:"id" yaml:"id"`
	Name           string   `json:"name" yaml:"name"`
	Description    string   `json:"description" yaml:"description"`
	Category       string   `json:"category" yaml:"category"`
	Tags           []string `json:"tags" yaml:"tags"`
	Prompt         string   `json:"prompt" yaml:"prompt"`
	SuggestedTools []string `json:"suggested_tools" yaml:"suggestedTools"`
}

// Store defines the interface for loading skills from a backend.
type Store interface {
	LoadSkills() ([]Skill, error)
}

func SkillToDocument(skill Skill) vector.Document {
	return vector.Document{
		ID:       generateDocumentID(skill),
		Content:  buildSearchContent(skill),
		Metadata: buildMetadata(skill),
	}
}

func generateDocumentID(skill Skill) string {
	data := fmt.Sprintf("%s|%s|%s", skill.ID, skill.Name, skill.Description)
	hash := sha256.Sum256([]byte(data))
	return "skill_" + hex.EncodeToString(hash[:12])
}

func buildSearchContent(skill Skill) string {
	parts := []string{skill.Name}

	if skill.Description != "" {
		parts = append(parts, skill.Description)
	}

	if skill.Category != "" {
		parts = append(parts, skill.Category)
	}

	if len(skill.Tags) > 0 {
		parts = append(parts, strings.Join(skill.Tags, " "))
	}

	return strings.Join(parts, " ")
}

func buildMetadata(skill Skill) map[string]string {
	metadata := map[string]string{
		"skill_id":    skill.ID,
		"skill_name":  skill.Name,
		"description": skill.Description,
		"category":    skill.Category,
	}

	if bs, err := json.Marshal(skill); err == nil {
		metadata["skill_json"] = string(bs)
	}

	return metadata
}
