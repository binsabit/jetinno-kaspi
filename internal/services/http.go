package services

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

func (s Server) RunHTTPServer(port int) error {
	s.SetUpRoutes()
	return s.Listen(fmt.Sprintf(":%d", port))
}

func (s *Server) SetUpRoutes() {
	s.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
}
