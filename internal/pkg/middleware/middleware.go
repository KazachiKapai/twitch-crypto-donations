package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Request[T any] struct {
	Body       T
	PathParams map[string]string
	Headers    http.Header
}

type Response[T any] struct {
	Body       T
	StatusCode int
}

type Handler[RequestBody any, ResponseBody any] interface {
	Handle(context.Context, Request[RequestBody]) (*Response[ResponseBody], error)
}

type Middleware[Request any, Response any] struct {
	handler Handler[Request, Response]
}

func New[RequestBody any, ResponseBody any](handler Handler[RequestBody, ResponseBody]) *Middleware[RequestBody, ResponseBody] {
	return &Middleware[RequestBody, ResponseBody]{
		handler: handler,
	}
}

func (m *Middleware[RequestBody, ResponseBody]) Handle(ctx *gin.Context) {
	var requestBody RequestBody
	if ctx.Request.Body != nil && ctx.Request.Body != http.NoBody {
		if err := json.NewDecoder(ctx.Request.Body).Decode(&requestBody); err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
	}

	pathParams := make(map[string]string)
	for _, param := range ctx.Params {
		pathParams[param.Key] = param.Value
	}

	requestData := Request[RequestBody]{
		Body:       requestBody,
		PathParams: pathParams,
		Headers:    ctx.Request.Header,
	}

	responseCode := http.StatusInternalServerError
	response, err := m.handler.Handle(ctx.Request.Context(), requestData)
	if err != nil {
		if response != nil && response.StatusCode != 0 {
			responseCode = response.StatusCode
		}

		ctx.AbortWithStatusJSON(responseCode, gin.H{"error": err.Error()})
		return
	}

	if response == nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	ctx.JSON(response.StatusCode, response.Body)
}
