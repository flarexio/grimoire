package yaml

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/flarexio/grimoire"
)

func NewStore(dir string) grimoire.Store {
	return &store{dir: dir}
}

type store struct {
	dir string
}

func (s *store) LoadSkills() ([]grimoire.Skill, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var skills []grimoire.Skill

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(s.dir, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var skill grimoire.Skill
		if err := yaml.Unmarshal(data, &skill); err != nil {
			return nil, err
		}

		skills = append(skills, skill)
	}

	return skills, nil
}
