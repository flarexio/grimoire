package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/flarexio/grimoire"
	"github.com/flarexio/grimoire/persistence/chromem"

	mcpE "github.com/flarexio/grimoire/mcp"
	yamlStore "github.com/flarexio/grimoire/store/yaml"
	httpT "github.com/flarexio/grimoire/transport/http"
)

func main() {
	cmd := &cli.Command{
		Name:  "grimoire",
		Usage: "Grimoire skill server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "Path to the Grimoire data directory",
			},
			&cli.StringFlag{
				Name:  "http-addr",
				Usage: "HTTP server address",
				Value: ":8080",
			},
		},
		Action: run,
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	path := cmd.String("path")
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path = filepath.Join(homeDir, ".flarex", "grimoire")
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	f, err := os.Open(filepath.Join(path, "config.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg grimoire.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	// Resolve relative skills directory
	if !filepath.IsAbs(cfg.SkillsDir) {
		cfg.SkillsDir = filepath.Join(path, cfg.SkillsDir)
	}

	cfg.Vector.Path = filepath.Join(path, "vectors")

	vectorDB, err := chromem.NewChromemVectorDB(cfg.Vector)
	if err != nil {
		return err
	}

	store := yamlStore.NewStore(cfg.SkillsDir)

	svc, err := grimoire.NewService(ctx, store, vectorDB, cfg)
	if err != nil {
		return err
	}
	defer svc.Close()

	svc = grimoire.LoggingMiddleware(logger)(svc)

	endpoints := grimoire.EndpointSet{
		ListSkills:   grimoire.ListSkillsEndpoint(svc),
		SearchSkills: grimoire.SearchSkillsEndpoint(svc),
		FindSkill:     grimoire.FindSkillEndpoint(svc),
	}

	r := gin.Default()
	httpT.AddRouters(r, endpoints)

	mcpEndpoints := make(map[mcp.MCPMethod]mcpE.MCPEndpoint)
	mcpEndpoints[mcp.MethodInitialize] = mcpE.InitializeEndpoint(svc)
	mcpEndpoints[mcp.MethodPing] = mcpE.PingEndpoint(svc)
	mcpEndpoints[mcp.MethodToolsList] = mcpE.ListToolsEndpoint(svc)
	mcpEndpoints[mcp.MethodToolsCall] = mcpE.CallToolEndpoint(svc)
	httpT.AddStreamableRouters(r, mcpEndpoints)

	httpAddr := cmd.String("http-addr")
	go r.Run(httpAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sign := <-quit

	logger.Info("graceful shutdown", zap.String("signal", sign.String()))
	return nil
}
