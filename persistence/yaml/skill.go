package yaml

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/flarexio/grimoire/skill"
)

func NewSkillRepository(dir string) skill.Repository {
	return &skillRepository{dir: dir}
}

type skillRepository struct {
	dir string
}

func (r *skillRepository) LoadSkills() ([]skill.Skill, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, err
	}

	var skills []skill.Skill

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		path := filepath.Join(r.dir, entry.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var s skill.Skill
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, err
		}

		skills = append(skills, s)
	}

	return skills, nil
}
