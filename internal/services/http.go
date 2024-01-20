package services

import (
	"github.com/gofiber/fiber/v2"
)

func (s Server) RunHTTPServer(port string) error {
	s.SetUpRoutes()
	return s.Listen(":" + port)
}

func (s *Server) SetUpRoutes() {
	s.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
}
