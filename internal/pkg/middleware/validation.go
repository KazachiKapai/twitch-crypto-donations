package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"twitch-crypto-donations/internal/pkg/environment"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/gin-gonic/gin"
)

type ValidationMiddleware struct {
	router routers.Router
	spec   *openapi3.T
}

func NewValidationMiddleware(swaggerPath environment.SwaggerPath) (*ValidationMiddleware, error) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(string(swaggerPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger file: %w", err)
	}

	ctx := context.Background()
	if err = spec.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid swagger specification: %w", err)
	}

	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	return &ValidationMiddleware{
		router: router,
		spec:   spec,
	}, nil
}

func (m *ValidationMiddleware) Request() gin.HandlerFunc {
	return func(c *gin.Context) {
		route, pathParams, err := m.router.FindRoute(c.Request)
		if err != nil {
			c.Next()
			return
		}

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Failed to read request body",
				})
				c.Abort()
				return
			}

			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		ctx := context.Background()
		if err = openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Request validation failed",
				"details": formatValidationError(err),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *ValidationMiddleware) Response() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		route, pathParams, err := m.router.FindRoute(c.Request)
		if err != nil {
			return
		}

		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: &openapi3filter.RequestValidationInput{
				Request:    c.Request,
				PathParams: pathParams,
				Route:      route,
			},
			Status: c.Writer.Status(),
			Header: c.Writer.Header(),
		}

		if blw.body.Len() > 0 {
			responseValidationInput.SetBodyBytes(blw.body.Bytes())
		}

		ctx := context.Background()
		if err = openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
			log.Printf("Response validation error: %v", err)
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func formatValidationError(err error) interface{} {
	switch e := err.(type) {
	case *openapi3filter.RequestError:
		if e.Err != nil {
			return formatValidationError(e.Err)
		}
		return e.Error()

	case *openapi3filter.SecurityRequirementsError:
		return map[string]string{
			"type":    "security",
			"message": "Security requirements not met",
		}

	case openapi3.MultiError:
		var errors []string
		for _, subErr := range e {
			errors = append(errors, subErr.Error())
		}
		return errors

	case *openapi3.SchemaError:
		return map[string]string{
			"type":    "schema",
			"message": e.Error(),
			"field":   e.SchemaField,
		}

	default:
		return err.Error()
	}
}
