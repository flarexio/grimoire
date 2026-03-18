package grimoire

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/flarexio/grimoire/skill"
	"github.com/flarexio/grimoire/vector"
)

type Service interface {
	Close() error
	ListSkills(ctx context.Context) ([]skill.Skill, error)
	SearchSkills(ctx context.Context, query string, k ...int) ([]skill.Skill, error)
	FindSkill(ctx context.Context, name string) (*skill.Skill, error)
}

type ServiceMiddleware func(Service) Service

func NewService(ctx context.Context, repo skill.Repository, vectorDB vector.VectorDB, cfg Config) (Service, error) {
	log := zap.L().With(
		zap.String("service", "grimoire"),
	)

	ctx, cancel := context.WithCancel(ctx)

	svc := &service{
		skillsByName: make(map[string]skill.Skill),
		log:          log,
		ctx:          ctx,
		cancel:       cancel,
	}

	if vectorDB != nil {
		collection, err := vectorDB.Collection(cfg.Vector.Collection)
		if err != nil {
			cancel()
			return nil, err
		}

		svc.collection = collection
	}

	skills, err := repo.LoadSkills()
	if err != nil {
		cancel()
		return nil, err
	}

	svc.skills = skills
	svc.indexSkills(ctx)

	return svc, nil
}

type service struct {
	skills       []skill.Skill
	skillsByName map[string]skill.Skill
	collection   vector.Collection

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

func (svc *service) indexSkills(ctx context.Context) {
	log := svc.log.With(
		zap.String("action", "index_skills"),
	)

	for _, s := range svc.skills {
		log := log.With(
			zap.String("skill_id", s.ID),
		)

		svc.skillsByName[s.Name] = s

		if svc.collection != nil {
			doc := SkillToDocument(s)
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

	log.Info("skills indexed", zap.Int("count", len(svc.skills)))
}

func (svc *service) ListSkills(ctx context.Context) ([]skill.Skill, error) {
	if len(svc.skills) == 0 {
		return nil, skill.ErrNoSkillsFound
	}

	return svc.skills, nil
}

func (svc *service) SearchSkills(ctx context.Context, query string, k ...int) ([]skill.Skill, error) {
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
		return nil, skill.ErrNoSkillsFound
	}

	skills := make([]skill.Skill, len(docs))
	for i, doc := range docs {
		skillJSON, ok := doc.Metadata["skill_json"]
		if !ok {
			return nil, skill.ErrInvalidSkillDocument
		}

		var s skill.Skill
		if err := json.Unmarshal([]byte(skillJSON), &s); err != nil {
			return nil, err
		}

		skills[i] = s
	}

	return skills, nil
}

func (svc *service) FindSkill(ctx context.Context, name string) (*skill.Skill, error) {
	s, ok := svc.skillsByName[name]
	if !ok {
		return nil, skill.ErrSkillNotFound
	}

	return &s, nil
}
