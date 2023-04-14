package identity

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kamva/mgm/v3"
	"github.com/superstackhq/common/logger"
	"github.com/superstackhq/identity/internal/app/identity/authentication"
	"github.com/superstackhq/identity/internal/app/identity/health"
	"github.com/superstackhq/identity/internal/app/identity/organization"
	"github.com/superstackhq/identity/internal/app/identity/user"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

const (
	serviceName = "identity"
)

type Config struct {
	Host          string
	Port          string
	MongoEndpoint string
	MongoDatabase string
	JwtSecretKey  string
}

type Server struct {
	config *Config
}

func NewServer(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() {
	logger.Init(serviceName)

	defer func() {
		_ = zap.L().Sync()
	}()

	err := mgm.SetDefaultConfig(nil, s.config.MongoDatabase, options.Client().ApplyURI(s.config.MongoEndpoint))

	if err != nil {
		zap.L().Panic("error while connecting to datastore", zap.Error(err))
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	authenticator := authentication.NewAuthenticator(s.config.JwtSecretKey)

	organizationManager := organization.NewManager()
	userManager := user.NewManager(organizationManager, authenticator)

	health.NewHandler(router).Register()
	organization.NewHandler(router, authenticator, organizationManager).Register()
	user.NewHandler(router, authenticator, userManager).Register()

	zap.L().Info("starting identity server", zap.String("host", s.config.Host), zap.String("port", s.config.Port))
	err = router.Run(fmt.Sprintf("%s:%s", s.config.Host, s.config.Port))

	if err != nil {
		zap.L().Panic("error while starting identity server", zap.Error(err))
	}
}
