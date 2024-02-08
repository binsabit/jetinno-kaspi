package services

import (
	"context"
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

	s.Post("/pay-done", func(ctx *fiber.Ctx) error {
		var input struct {
			OrderID int64 `json:"order_id"`
		}

		if err := ctx.BodyParser(&input); err != nil {
			log.Println("http-server:", err)
			return ctx.SendStatus(fiber.StatusBadRequest)
		}

		log.Println("ORDER ID", input.OrderID)
		order, err := db.Storage.GetOrderByID(ctx.Context(), input.OrderID)
		if err != nil {
			log.Println("http-server:", err)
			return ctx.SendStatus(fiber.StatusInternalServerError)
		}

		s.EnsureOrderPayment(order)

		return ctx.SendStatus(fiber.StatusOK)

	})
}

func (s *Server) EnsureOrderPayment(order db.Order) {

	if order.Status != 0 {
		return
	}

	vmcno, err := strconv.ParseInt(order.VendingMachineNo, 10, 64)
	if err != nil {
		log.Println("could not parse vcmno")
		return
	}
	for order.Status == 0 {
		log.Println("trying to ", vmcno)
		val, ok := s.TCPServer.Clients.Load(vmcno)
		if !ok {
			log.Println("did not found vcmno")
			return
		}

		res := val.(*Client).PayDone(context.Background(), order)
		if res == nil {
			log.Println("ENSURE PAYMENT PAY DONE ERROR")
			return
		}

		err := val.(*Client).Write(*res)

		if err != nil {
			log.Println("ENSURE PAYMENT PAY DONE ERROR: ", err)
			return
		}

		order, err = db.Storage.GetOrderByID(context.Background(), order.ID)
		if err != nil {
			log.Println("ENSURE PAYMENT PAY DONE ERROR: ", err)
			return
		}
	}
}
