package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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
	httppkg "twitch-crypto-donations/internal/pkg/http"
	"twitch-crypto-donations/internal/pkg/jwt"
	"twitch-crypto-donations/internal/pkg/logger"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/obsservice"
	"twitch-crypto-donations/internal/pkg/router"
	"twitch-crypto-donations/internal/pkg/server"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

type (
	ConnectionString string
)

func NewLogger() *logger.LogrusAdapter {
	return logger.New(logrus.StandardLogger())
}

func NewHttpClient() *http.Client {
	return &http.Client{}
}

func NewRpcClient(rpcEndpoint environment.RpcEndpoint) *rpc.Client {
	return rpc.New(string(rpcEndpoint))
}

func NewDatabase(connString ConnectionString, dir environment.MigrationsDir) *sql.DB {
	log.Printf("DEBUG: Connection string: %s", string(connString))

	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", string(connString))
		if err != nil {
			log.Printf("Attempt %d: Failed to open DB: %v", i+1, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		err = db.Ping()
		if err == nil {
			log.Printf("Successfully connected to database on attempt %d", i+1)
			break
		}

		log.Printf("Attempt %d: Failed to ping DB: %v", i+1, err)
		db.Close()
		time.Sleep(time.Second * time.Duration(i+1))
	}

	if err != nil {
		panic(fmt.Errorf("failed to connect to DB after 10 attempts: %w", err))
	}

	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set postgres dialect: %v", err)
	}

	if err = goose.Up(db, string(dir)); err != nil {
		log.Fatalf("failed to up migrations: %v", err)
	}

	return db
}

func NewConnectionString(
	host environment.DBHost,
	port environment.DBPort,
	user environment.DBUser,
	password environment.DBPassword,
	dbName environment.DBName,
	dbSSLMode environment.DBSSLMode,
) ConnectionString {
	hostStr := string(host)

	var connStr string

	if strings.HasPrefix(hostStr, "/") {
		connStr = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s",
			hostStr, user, password, dbName, dbSSLMode,
		)
	} else {
		connStr = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			hostStr, port, user, password, dbName, dbSSLMode,
		)
	}

	return ConnectionString(connStr)
}

func NewEngine(
	handlers router.Handlers,
	prefixRouter environment.RoutePrefix,
	swaggerPath environment.SwaggerPath,
	secret environment.JwtSecret,
	logger *logger.LogrusAdapter,
	middlewares []gin.HandlerFunc,
) *gin.Engine {
	return router.New(gin.New(), handlers, prefixRouter, swaggerPath, middleware.NewJwtMiddleware(secret, logger), middlewares...)
}

func NewMiddlewares(appEnv environment.AppEnv, path environment.SwaggerPath) []gin.HandlerFunc {
	corsMiddleware := middleware.NewCorsMiddleware()

	validationMiddleware, err := middleware.NewValidationMiddleware(path)
	if err != nil {
		log.Fatalf("failed to initialize validation middleware: %v", err)
	}

	middlewares := []gin.HandlerFunc{validationMiddleware.Response(), validationMiddleware.Request(), corsMiddleware.Request()}
	if appEnv == "development" {
		middlewares = append(middlewares, gin.Logger())
	}

	return middlewares
}

func NewServer(engine *gin.Engine, listenPort environment.HTTPListenPort) *server.Server {
	return server.New(engine, string(listenPort))
}

var WireSet = wire.NewSet(
	environment.WireSet,
	jwt.New,
	httppkg.New,
	obsservice.New,
	senddonate.New,
	setuserinfo.New,
	setobswebhooks.New,
	noncegeneration.New,
	getstreamerinfo.New,
	donationshistory.New,
	paymentconfirmation.New,
	getdefaultobssettings.New,
	signatureverification.New,
	updatedefaultobssettings.New,

	wire.Bind(new(getdefaultobssettings.Database), new(*sql.DB)),
	wire.Bind(new(getdefaultobssettings.ObsService), new(*obsservice.ObsService)),
	wire.Bind(new(setuserinfo.Database), new(*sql.DB)),
	wire.Bind(new(getstreamerinfo.Database), new(*sql.DB)),
	wire.Bind(new(updatedefaultobssettings.ObsService), new(*obsservice.ObsService)),
	wire.Bind(new(donationshistory.Database), new(*sql.DB)),
	wire.Bind(new(paymentconfirmation.RpcClient), new(*rpc.Client)),
	wire.Bind(new(noncegeneration.Database), new(*sql.DB)),
	wire.Bind(new(signatureverification.JwtManager), new(*jwt.Manager)),
	wire.Bind(new(signatureverification.Database), new(*sql.DB)),
	wire.Bind(new(setobswebhooks.ObsService), new(*obsservice.ObsService)),
	wire.Bind(new(setobswebhooks.Database), new(*sql.DB)),
	wire.Bind(new(senddonate.ObsService), new(*obsservice.ObsService)),
	wire.Bind(new(senddonate.Database), new(*sql.DB)),
	wire.Bind(new(obsservice.Logger), new(*logger.LogrusAdapter)),
	wire.Bind(new(obsservice.Database), new(*sql.DB)),
	wire.Bind(new(obsservice.HttpClient), new(*httppkg.Client)),
	wire.Bind(new(httppkg.HttpClient), new(*http.Client)),

	wire.Struct(new(router.Handlers), "*"),

	NewLogger,
	NewRpcClient,
	NewConnectionString,
	NewDatabase,
	NewHttpClient,
	NewMiddlewares,
	NewEngine,
	NewServer,
)
