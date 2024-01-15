package services

import (
	"fmt"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func (s Server) RunHTTPServer(port string) error {
	s.SetUpRoutes()
	return s.HTTPServer.Listen(":" + port)
}
func (s *Server) SetUpRoutes() {

	s.HTTPServer.Get("/payment", s.WebHookHandler)
	s.HTTPServer.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	s.HTTPServer.Get("/log/:id", func(ctx *fiber.Ctx) error {
		return ctx.Download(fmt.Sprintf("./logs/%s.txt", ctx.Params("id")))
	})

	s.HTTPServer.Get("/product/:id", func(ctx *fiber.Ctx) error {
		vmcNo, err := ctx.ParamsInt("id", 0)
		if err != nil {
			return err
		}

		tcpClient, ok := s.TCPServer.Clients.Load(vmcNo)

		if !ok {
			return fmt.Errorf("vmc not exists")
		}

		return tcpClient.(*Client).Write(Request{
			VmcNo:   int64(vmcNo),
			Command: "rinsing",
		})

	})
}

func (s *Server) WebHookHandler(ctx *fiber.Ctx) error {
	query := ctx.Queries()
	request := NewKaspiWebHookRequest(query)

	var response KaspiWebHookResponse
	switch request.Command {
	case "check":
	case "pay":
		prvTxnID, result, err := s.ProcessPayment(ctx.Context(), request)
		if err != nil {
			log.Printf("Error while handling webhook payment: %v err:%v\n", request, err)
		}

		response = KaspiWebHookResponse{
			Result:        result,
			TxnID:         request.TxnID,
			ProviderTxnID: prvTxnID,
			Sum:           strconv.FormatFloat(request.Sum, 'f', 2, 64),
		}
	}

	if response.Result == pkg.KASPI_PAYMENT_SUCCESS {
		//vccNo, _ := strconv.ParseInt(request.Account, 10, 64)
		//s.TCPClients[vccNo].WriteToConn([]byte("success"))
	}
	return ctx.Status(fiber.StatusOK).XML(response)
}
