package grimoire

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flarexio/grimoire/skill"
	"github.com/flarexio/grimoire/vector"
)

var (
	ErrVectorDBNotSet = errors.New("vector database not set")
)

type Config struct {
	SkillsDir string        `yaml:"skillsDir"`
	Vector    vector.Config `yaml:"vector"`
}

func SkillToDocument(s skill.Skill) vector.Document {
	return vector.Document{
		ID:       generateDocumentID(s),
		Content:  buildSearchContent(s),
		Metadata: buildMetadata(s),
	}
}

func generateDocumentID(s skill.Skill) string {
	data := fmt.Sprintf("%s|%s|%s", s.ID, s.Name, s.Description)
	hash := sha256.Sum256([]byte(data))
	return "skill_" + hex.EncodeToString(hash[:12])
}

func buildSearchContent(s skill.Skill) string {
	parts := []string{s.Name}

	if s.Description != "" {
		parts = append(parts, s.Description)
	}

	if len(s.Tags) > 0 {
		parts = append(parts, strings.Join(s.Tags, " "))
	}

	return strings.Join(parts, " ")
}

func buildMetadata(s skill.Skill) map[string]string {
	metadata := map[string]string{
		"skill_id":    s.ID,
		"skill_name":  s.Name,
		"description": s.Description,
	}

	if bs, err := json.Marshal(s); err == nil {
		metadata["skill_json"] = string(bs)
	}

	return metadata
}
