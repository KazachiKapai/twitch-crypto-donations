package router

import (
	"twitch-crypto-donations/internal/app/noncegeneration"
	"twitch-crypto-donations/internal/app/paymentconfirmation"
	"twitch-crypto-donations/internal/app/senddonate"
	"twitch-crypto-donations/internal/app/setobswebhooks"
	"twitch-crypto-donations/internal/app/signatureverification"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/gin-gonic/gin"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Register              *setobswebhooks.Handler
	SendDonate            *senddonate.Handler
	NonceGenerator        *noncegeneration.Handler
	PaymentConfirmation   *paymentconfirmation.Handler
	SignatureVerification *signatureverification.Handler
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
		api.POST("/generate-nonce", middleware.New(handlers.NonceGenerator).Handle)
		api.POST("/verify-signature", middleware.New(handlers.SignatureVerification).Handle)
		api.POST("/set-obs-webhooks", middleware.New(handlers.Register).Handle)
		api.POST("/send-donate", middleware.New(handlers.SendDonate).Handle)
		api.POST("/confirm-payment", middleware.New(handlers.PaymentConfirmation).Handle)
	}

	return engine
}
