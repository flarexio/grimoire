package grimoire

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/flarexio/grimoire/vector"
)

type Service interface {
	Close() error
	ListSkills(ctx context.Context, category string) ([]Skill, error)
	SearchSkills(ctx context.Context, query string, k ...int) ([]Skill, error)
	FindSkill(ctx context.Context, id string) (*Skill, error)
}

type ServiceMiddleware func(Service) Service

func NewService(ctx context.Context, store Store, vectorDB vector.VectorDB, cfg Config) (Service, error) {
	log := zap.L().With(
		zap.String("service", "grimoire"),
	)

	ctx, cancel := context.WithCancel(ctx)

	svc := &service{
		skills: make(map[string]Skill),
		log:    log,
		ctx:    ctx,
		cancel: cancel,
	}

	if vectorDB != nil {
		collection, err := vectorDB.Collection(cfg.Vector.Collection)
		if err != nil {
			cancel()
			return nil, err
		}

		svc.collection = collection
	}

	skills, err := store.LoadSkills()
	if err != nil {
		cancel()
		return nil, err
	}

	svc.indexSkills(ctx, skills)

	return svc, nil
}

type service struct {
	skills     map[string]Skill
	collection vector.Collection

	log    *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func (svc *service) Close() error {
	if svc.cancel != nil {
		svc.cancel()
		svc.cancel = nil
	}

	return nil
}

func (svc *service) indexSkills(ctx context.Context, skills []Skill) {
	log := svc.log.With(
		zap.String("action", "index_skills"),
	)

	for _, skill := range skills {
		log := log.With(
			zap.String("skill_id", skill.ID),
		)

		svc.skills[skill.ID] = skill

		if svc.collection != nil {
			doc := SkillToDocument(skill)
			existingDoc, err := svc.collection.FindDocument(ctx, doc.ID)
			if err != nil || existingDoc.ID != doc.ID {
				if err := svc.collection.AddDocument(ctx, doc); err != nil {
					log.Error(err.Error())
					continue
				}

				log.Info("added skill to vector collection")
			}
		}
	}

	log.Info("skills indexed", zap.Int("count", len(skills)))
}

func (svc *service) ListSkills(ctx context.Context, category string) ([]Skill, error) {
	skills := make([]Skill, 0, len(svc.skills))

	for _, skill := range svc.skills {
		if category != "" && skill.Category != category {
			continue
		}

		skills = append(skills, skill)
	}

	if len(skills) == 0 {
		return nil, ErrNoSkillsFound
	}

	return skills, nil
}

func (svc *service) SearchSkills(ctx context.Context, query string, k ...int) ([]Skill, error) {
	if svc.collection == nil {
		return nil, ErrVectorDBNotSet
	}

	n := 5
	if len(k) > 0 && k[0] > 0 {
		n = k[0]
	}

	docs, err := svc.collection.Query(ctx, query, n)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, ErrNoSkillsFound
	}

	skills := make([]Skill, len(docs))
	for i, doc := range docs {
		skillJSON, ok := doc.Metadata["skill_json"]
		if !ok {
			return nil, ErrInvalidSkillDocument
		}

		var skill Skill
		if err := json.Unmarshal([]byte(skillJSON), &skill); err != nil {
			return nil, err
		}

		skills[i] = skill
	}

	return skills, nil
}

func (svc *service) FindSkill(ctx context.Context, id string) (*Skill, error) {
	skill, ok := svc.skills[id]
	if !ok {
		return nil, ErrSkillNotFound
	}

	return &skill, nil
}
