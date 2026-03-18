package grimoire

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"
)

type EndpointSet struct {
	ListSkills   endpoint.Endpoint
	SearchSkills endpoint.Endpoint
	FindSkill    endpoint.Endpoint
}

func ListSkillsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		return svc.ListSkills(ctx)
	}
}

type SearchSkillsRequest struct {
	Query string `json:"query"`
	K     int    `json:"k,omitempty"`
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
	Name string `json:"name"`
}

func FindSkillEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req, ok := request.(FindSkillRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}

		return svc.FindSkill(ctx, req.Name)
	}
}
