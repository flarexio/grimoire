package grimoire

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

type EndpointSet struct {
	ListSkills   endpoint.Endpoint
	SearchSkills endpoint.Endpoint
	FindSkill     endpoint.Endpoint
}

type ListSkillsRequest struct {
	Category string `json:"category" form:"category"`
}

func ListSkillsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(ListSkillsRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return svc.ListSkills(ctx, req.Category)
	}
}

type SearchSkillsRequest struct {
	Query string `json:"query" form:"query"`
	K     int    `json:"k,omitempty" form:"k"`
}

func SearchSkillsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(SearchSkillsRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return svc.SearchSkills(ctx, req.Query, req.K)
	}
}

type FindSkillRequest struct {
	ID string `json:"id"`
}

func FindSkillEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(FindSkillRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return svc.FindSkill(ctx, req.ID)
	}
}
