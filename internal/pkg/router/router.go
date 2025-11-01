package router

import (
	"fmt"
	"twitch-crypto-donations/internal/app/donationsanalytics"
	"twitch-crypto-donations/internal/app/donationshistory"
	"twitch-crypto-donations/internal/app/getdefaultobssettings"
	"twitch-crypto-donations/internal/app/getstreamerinfo"
	"twitch-crypto-donations/internal/app/noncegeneration"
	"twitch-crypto-donations/internal/app/paymentconfirmation"
	"twitch-crypto-donations/internal/app/senddonate"
	"twitch-crypto-donations/internal/app/setobswebhooks"
	"twitch-crypto-donations/internal/app/setuserinfo"
	"twitch-crypto-donations/internal/app/signatureverification"
	"twitch-crypto-donations/internal/app/updatedefaultobssettings"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"

	"github.com/gin-gonic/gin"

	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

type Handlers struct {
	DonationsAnalytics       *donationsanalytics.Handler
	SetUserInfo              *setuserinfo.Handler
	GetStreamerInfo          *getstreamerinfo.Handler
	SetObsWebhooks           *setobswebhooks.Handler
	SendDonate               *senddonate.Handler
	NonceGenerator           *noncegeneration.Handler
	PaymentConfirmation      *paymentconfirmation.Handler
	SignatureVerification    *signatureverification.Handler
	DonationsHistory         *donationshistory.Handler
	GetDefaultObsSettings    *getdefaultobssettings.Handler
	UpdateDefaultObsSettings *updatedefaultobssettings.Handler
}

func New(
	engine *gin.Engine,
	handlers Handlers,
	routePrefix environment.RoutePrefix,
	swaggerPath environment.SwaggerPath,
	jwtMiddleware *middleware.JwtMiddleware,
	middlewares ...gin.HandlerFunc,
) *gin.Engine {
	engine.StaticFile("/swagger.yml", string(swaggerPath))
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger.yml"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	secure := engine.Group(fmt.Sprintf("%s/secure", routePrefix))
	secure.Use(middlewares...)
	secure.Use(jwtMiddleware.Request())
	{
		secure.GET("/donations-analytics", middleware.New(handlers.DonationsAnalytics).Handle)
		secure.PUT("/me", middleware.New(handlers.SetUserInfo).Handle)
		secure.GET("/me", middleware.New(handlers.GetStreamerInfo).Handle)
		secure.GET("/donations-history", middleware.New(handlers.DonationsHistory).Handle)
		secure.PUT("/update-default-obs-settings", middleware.New(handlers.UpdateDefaultObsSettings).Handle)
	}

	api := engine.Group(string(routePrefix))
	api.Use(middlewares...)
	{
		api.GET("/get-default-obs-settings/:address", middleware.New(handlers.GetDefaultObsSettings).Handle)
		api.GET("/streamer-info/:username", middleware.New(handlers.GetStreamerInfo).Handle)
		api.POST("/generate-nonce", middleware.New(handlers.NonceGenerator).Handle)
		api.POST("/verify-signature", middleware.New(handlers.SignatureVerification).Handle)
		api.POST("/set-obs-webhooks", middleware.New(handlers.SetObsWebhooks).Handle)
		api.POST("/send-donate", middleware.New(handlers.SendDonate).Handle)
		api.POST("/confirm-payment", middleware.New(handlers.PaymentConfirmation).Handle)
	}

	return engine
}
