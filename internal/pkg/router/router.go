package router

import (
	"twitch-crypto-donations/internal/app/register"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/gin-gonic/gin"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Register *register.Handler
}

func New(
	engine *gin.Engine,
	handlers Handlers,
	routePrefix environment.RoutePrefix,
	swaggerPath environment.SwaggerPath,
	middlewares ...gin.HandlerFunc,
) *gin.Engine {
	engine.StaticFile("/swagger.yml", string(swaggerPath))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.yml"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	engine.Use(middlewares...)

	api := engine.Group(string(routePrefix))
	{
		api.POST("/register", middleware.New(handlers.Register).Handle)
	}

	return engine
}
