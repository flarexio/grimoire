package grimoire

import (
	"context"

	"go.uber.org/zap"

	"github.com/flarexio/grimoire/skill"
)

func LoggingMiddleware(log *zap.Logger) ServiceMiddleware {
	log = log.With(
		zap.String("service", "grimoire"),
	)

	return func(next Service) Service {
		log.Info("service initialized")

		return &loggingMiddleware{
			log:  log,
			next: next,
		}
	}
}

type loggingMiddleware struct {
	log  *zap.Logger
	next Service
}

func (mw *loggingMiddleware) Close() error {
	log := mw.log.With(
		zap.String("action", "close"),
	)

	err := mw.next.Close()
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info("service closed")
	return nil
}

func (mw *loggingMiddleware) ListSkills(ctx context.Context) ([]skill.Skill, error) {
	log := mw.log.With(
		zap.String("action", "list_skills"),
	)

	skills, err := mw.next.ListSkills(ctx)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("skills listed", zap.Int("count", len(skills)))
	return skills, nil
}

func (mw *loggingMiddleware) SearchSkills(ctx context.Context, query string, k ...int) ([]skill.Skill, error) {
	log := mw.log.With(
		zap.String("action", "search_skills"),
		zap.String("query", query),
	)

	if len(k) > 0 && k[0] > 0 {
		log = log.With(zap.Int("k", k[0]))
	}

	skills, err := mw.next.SearchSkills(ctx, query, k...)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("skills searched", zap.Int("count", len(skills)))
	return skills, nil
}

func (mw *loggingMiddleware) FindSkill(ctx context.Context, name string) (*skill.Skill, error) {
	log := mw.log.With(
		zap.String("action", "find_skill"),
		zap.String("skill_name", name),
	)

	s, err := mw.next.FindSkill(ctx, name)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	log.Info("skill retrieved")
	return s, nil
}
