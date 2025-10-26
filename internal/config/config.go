package config

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"twitch-crypto-donations/internal/app/register"
	"twitch-crypto-donations/internal/pkg/environment"
	"twitch-crypto-donations/internal/pkg/middleware"
	"twitch-crypto-donations/internal/pkg/router"
	"twitch-crypto-donations/internal/pkg/server"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

type (
	ConnectionString string
)

func NewHttpClient() *http.Client {
	return &http.Client{}
}

func NewDatabase(connString ConnectionString, dir environment.MigrationsDir) *sql.DB {
	db, err := sql.Open("postgres", string(connString))
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(fmt.Errorf("failed to connect to DB: %w", err))
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
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, dbSSLMode,
	)
	return ConnectionString(connStr)
}

func NewEngine(handlers router.Handlers, prefixRouter environment.RoutePrefix, swaggerPath environment.SwaggerPath, middlewares []gin.HandlerFunc) *gin.Engine {
	return router.New(gin.New(), handlers, prefixRouter, swaggerPath, middlewares...)
}

func NewMiddlewares(appEnv environment.AppEnv, path environment.SwaggerPath) []gin.HandlerFunc {
	validationMiddleware, err := middleware.NewValidationMiddleware(path)
	if err != nil {
		log.Fatalf("failed to initialize validation middleware: %v", err)
	}

	middlewares := []gin.HandlerFunc{validationMiddleware.Response(), validationMiddleware.Request()}
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
	register.New,

	wire.Bind(new(register.HttpClient), new(*http.Client)),
	wire.Bind(new(register.Database), new(*sql.DB)),
	wire.Struct(new(router.Handlers), "*"),

	NewConnectionString,
	NewDatabase,
	NewHttpClient,
	NewMiddlewares,
	NewEngine,
	NewServer,
)
