package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"

	"github.com/flarexio/grimoire"
)

func ListSkillsHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		resp, err := endpoint(ctx, nil)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func SearchSkillsHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("query")
		k, _ := strconv.Atoi(c.Query("k"))

		req := grimoire.SearchSkillsRequest{
			Query: query,
			K:     k,
		}

		ctx := c.Request.Context()
		resp, err := endpoint(ctx, req)
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}

func FindSkillHandler(endpoint endpoint.Endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		ctx := c.Request.Context()
		resp, err := endpoint(ctx, grimoire.FindSkillRequest{Name: name})
		if err != nil {
			c.String(http.StatusExpectationFailed, err.Error())
			c.Error(err)
			c.Abort()
			return
		}

		c.JSON(http.StatusOK, &resp)
	}
}
