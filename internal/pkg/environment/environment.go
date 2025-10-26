package environment

import (
	"fmt"
	"os"
	"strconv"

	"github.com/google/wire"
)

type (
	HTTPListenPort string
	RoutePrefix    string
	AppEnv         string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	MigrationsDir string

	SwaggerPath string

	OBSServiceDomain string

	JwtSecret            string
	TokenExpirationHours int
)

func getEnv(key string) (string, error) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("missing environment variable %s", key)
	}
	return val, nil
}

func GetHTTPListenPort() (HTTPListenPort, error) {
	val, err := getEnv("HTTP_LISTEN_PORT")
	return HTTPListenPort(val), err
}

func GetRoutePrefix() (RoutePrefix, error) {
	val, err := getEnv("ROUTE_PREFIX")
	return RoutePrefix(val), err
}

func GetAppEnv() (AppEnv, error) {
	val, err := getEnv("APP_ENV")
	return AppEnv(val), err
}

func GetDBHost() (DBHost, error) {
	val, err := getEnv("DB_HOST")
	return DBHost(val), err
}

func GetDBPort() (DBPort, error) {
	val, err := getEnv("DB_PORT")
	return DBPort(val), err
}

func GetDBUser() (DBUser, error) {
	val, err := getEnv("POSTGRES_USER")
	return DBUser(val), err
}

func GetDBPassword() (DBPassword, error) {
	val, err := getEnv("POSTGRES_PASSWORD")
	return DBPassword(val), err
}

func GetDBName() (DBName, error) {
	val, err := getEnv("POSTGRES_DB")
	return DBName(val), err
}

func GetDBSSLMode() (DBSSLMode, error) {
	val, err := getEnv("DB_SSLMODE")
	return DBSSLMode(val), err
}

func GetMigrationsDir() (MigrationsDir, error) {
	val, err := getEnv("POSTGRES_MIGRATIONS_DIR")
	return MigrationsDir(val), err
}

func GetSwaggerPath() (SwaggerPath, error) {
	val, err := getEnv("SWAGGER_PATH")
	return SwaggerPath(val), err
}

func GetOBSServiceDomain() (OBSServiceDomain, error) {
	val, err := getEnv("OBS_SERVICE_DOMAIN")
	return OBSServiceDomain(val), err
}

func GetJwtSecret() (JwtSecret, error) {
	val, err := getEnv("JWT_SECRET")
	return JwtSecret(val), err
}

func GetTokenExpirationHours() (TokenExpirationHours, error) {
	val, err := getEnv("JWT_TOKEN_EXPIRATION_HOURS")
	if err != nil {
		return 0, err
	}

	rv, err := strconv.Atoi(val)
	return TokenExpirationHours(rv), err
}

var WireSet = wire.NewSet(
	GetHTTPListenPort,
	GetRoutePrefix,
	GetAppEnv,
	GetDBHost,
	GetDBPort,
	GetDBUser,
	GetDBPassword,
	GetDBName,
	GetDBSSLMode,
	GetMigrationsDir,
	GetSwaggerPath,
	GetOBSServiceDomain,
	GetJwtSecret,
	GetTokenExpirationHours,
)
