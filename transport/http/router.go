package http

import (
	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/flarexio/grimoire"

	mcpE "github.com/flarexio/grimoire/mcp"
)

func AddRouters(r *gin.Engine, endpoints grimoire.EndpointSet) {
	api := r.Group("/api")
	{
		api.GET("/skills", ListSkillsHandler(endpoints.ListSkills))
		api.GET("/skills/search", SearchSkillsHandler(endpoints.SearchSkills))
		api.GET("/skills/:name", FindSkillHandler(endpoints.FindSkill))
	}
}

func AddStreamableRouters(r *gin.Engine, endpoints map[mcp.MCPMethod]mcpE.MCPEndpoint) {
	mcpGroup := r.Group("/mcp")
	{
		mcpGroup.POST("/", MCPStreamableHandler(endpoints))
	}
}
