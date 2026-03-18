package skill

type Repository interface {
	LoadSkills() ([]Skill, error)
}
