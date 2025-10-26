package server

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine     *gin.Engine
	listenPort string
}

func New(engine *gin.Engine, listenPort string) *Server {
	return &Server{
		listenPort: listenPort,
		engine:     engine,
	}
}

func (s *Server) ServerHTTP() {
	err := s.engine.Run(":" + s.listenPort)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}

		log.Println("Server gracefully shut down.")
	}
}
