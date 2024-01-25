package services

import (
	"fmt"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func (s Server) RunHTTPServer(port int) error {
	s.SetUpRoutes()
	return s.Listen(fmt.Sprintf(":%d", port))

}

func (s *Server) SetUpRoutes() {
	s.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	s.Post("pay-done", func(ctx *fiber.Ctx) error {
		var input struct {
			OrderID int64 `json:"order_id"`
		}

		if err := ctx.BodyParser(&input); err != nil {
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		order, err := db.Storage.GetOrderByID(ctx.Context(), input.OrderID)
		if err != nil {
			log.Println("http-server:", err)
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		vmcno, _ := strconv.Atoi(order.VendingMachineNo)
		val, ok := s.TCPServer.Clients.Load(vmcno)
		if !ok {
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		res := val.(*Client).PayDone(ctx.Context(), order)
		if res == nil {
			return ctx.SendStatus(fiber.StatusNotFound)
		}

		err = val.(*Client).Write(*res)

		if err != nil {
			log.Println("http-server", err)
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		return ctx.SendStatus(fiber.StatusOK)

	})
}
