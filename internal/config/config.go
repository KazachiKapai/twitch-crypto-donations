package config

import (
	"twitch-crypto-donations/internal/app/register"
	"twitch-crypto-donations/internal/pkg/database"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/router"
	"twitch-crypto-donations/internal/pkg/server"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func NewEngine(handlers router.Handlers, prefixRouter environment.RoutePrefix, middlewares []gin.HandlerFunc) *gin.Engine {
	return router.New(gin.New(), handlers, string(prefixRouter), middlewares...)
}

func NewMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.Logger(),
	}
}

func NewServer(engine *gin.Engine, listenPort environment.HTTPListenPort) *server.Server {
	return server.New(engine, string(listenPort))
}

var WireSet = wire.NewSet(
	environment.WireSet,

	database.New,
	wire.Bind(new(register.Database), new(*database.Database)),

	register.New,
	NewMiddlewares,
	wire.Struct(new(router.Handlers), "*"),
	NewEngine,
	NewServer,
)
