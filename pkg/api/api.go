package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"go.uber.org/zap"
)

type APIController interface {
	BasePath() string
	Register(rg *gin.RouterGroup) error
}

type Server struct {
	gin    *gin.Engine
	config config.Config
	auth   *authHandler
}

func NewServer(log *zap.Logger, cfg config.Config, debug bool) *Server {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(
		ginzap.Ginzap(log, time.RFC3339, true),
		ginzap.RecoveryWithZap(log, true),
	)

	engine.NoRoute(ServeSPA("/", "./frontend/dist/"))

	if debug {
		engine.Use(
			cors.New(cors.Config{
				AllowOrigins: []string{"http://localhost:5173"},
				AllowMethods: []string{"GET", "PUT", "PATCH", "POST", "OPTIONS"},
				AllowHeaders: []string{"Origin", "Authorization", "Content-Type"},
				MaxAge:       12 * time.Hour,
			}),
		)
	}

	auth := newAuth(log.Sugar(), cfg)

	s := &Server{
		gin:    engine,
		config: cfg,
		auth:   auth,
	}

	engine.GET("api/config", s.getConfig)

	return s
}

func (s *Server) RegisterAll(controllers []APIController) error {
	r := s.gin.Group("api")
	for _, c := range controllers {
		if err := c.Register(r.Group(c.BasePath(), s.auth.middleware())); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) Listen() {
	if s.config.Server.TLSCertFile != "" && s.config.Server.TLSKeyFile != "" {
		s.gin.RunTLS(s.config.Server.ListenAddress, s.config.Server.TLSCertFile, s.config.Server.TLSKeyFile)
	}
	s.gin.Run(s.config.Server.ListenAddress)
}

type FrontendConfig struct {
	OIDCAuthority string `json:"oidcAuthority"`
	OIDCClientID  string `json:"oidcClientID"`
}

func (s *Server) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, FrontendConfig{
		OIDCAuthority: s.config.Frontend.OIDCAuthority,
		OIDCClientID:  s.config.Frontend.OIDCClientID,
	})
}
