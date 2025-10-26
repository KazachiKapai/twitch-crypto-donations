package router

import (
	"twitch-crypto-donations/internal/app/register"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Register *register.Handler
}

func New(engine *gin.Engine, handlers Handlers, routePrefix string, middlewares ...gin.HandlerFunc) *gin.Engine {
	engine.Use(middlewares...)

	api := engine.Group(routePrefix)
	api.POST("/register", middleware.New(handlers.Register).Handle)

	return engine
}
