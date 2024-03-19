package services

import (
	"context"
	"fmt"
	"github.com/binsabit/jetinno-kapsi/internal/db"
	"github.com/binsabit/jetinno-kapsi/pkg"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	KaspiLogin          string
	KaspiPassword       string
	KaspiRefundURL      string
	KaspiRefundDuration time.Duration
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
	log.Printf("order: %+v\n", order)
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
		time.Sleep(time.Second * 3)
	}

}

func Refund(vcm int64, id int64) {

	order, err := db.Storage.GetOrderByID(context.Background(), id)
	if err != nil {
		log.Println("ENSURE PAYMENT PAY DONE ERROR: ", err)
		return
	}

	if order.Status == 2 || order.Status == 0 {
		return
	}

	token, err := GetTokenKaspi()
	if err != nil {
		log.Printf("refund failed vcm: %d, err: %v\n", vcm, err)
		return
	}

	err = MakeRefund(token, order)
	if err != nil {
		log.Println(err)
		return
	}

	err = db.Storage.UpdateOrder(context.Background(), vcm, order.OrderNo, 3)
	if err != nil {
		log.Println(err)
		return
	}
}

type KaspiLoginResponse struct {
	Token *string `json:"token,omitempty"`
}

func GetTokenKaspi() (string, error) {
	loginReq, err := pkg.NewRequest(
		KaspiRefundURL+"returnApi/Auth/GetToken",
		http.MethodPost,
		map[string]string{
			"Content-Type": "application/json",
		},
		map[string]string{
			"Login":    KaspiLogin,
			"Password": KaspiPassword,
		})

	if err != nil {
		return "", err
	}

	data, err := loginReq.Do()
	if err != nil {
		return "", fmt.Errorf("could not do request %v", err)
	}

	var loginResp KaspiLoginResponse

	err = sonic.ConfigFastest.Unmarshal(data, &loginResp)
	if err != nil {
		return "", fmt.Errorf("could not marshal data %v", err)
	}

	if loginResp.Token == nil {
		return "", fmt.Errorf("unauthorized for token")
	}
	return *loginResp.Token, nil

}

type RefundResponse struct {
	StatusCode            int    `json:"statusCode"`
	RequestIdentificatior string `json:"requestIdentificatior"`
	Error                 struct {
		Code         int    `json:"code"`
		ErrorMessage string `json:"errorMessage"`
	} `json:"error"`
}

func MakeRefund(token string, order db.Order) error {
	refundReq, err := pkg.NewRequest(
		KaspiRefundURL+"returnApi/Refund/RefundRequest",
		http.MethodPost,
		map[string]string{
			"Content-Type": "application/json",
			"token":        token,
		},
		map[string]any{
			"PaymentId":            order.TxnID,
			"ReturnAmount":         int64(order.Amount),
			"RefundIdentificatior": uuid.New().String(),
			"Reason":               "возврат средств клиенту",
		})

	if err != nil {
		return err
	}

	data, err := refundReq.Do()
	if err != nil {
		return err
	}
	var resp RefundResponse

	err = sonic.ConfigFastest.Unmarshal(data, &resp)
	if err != nil {
		return err
	}
	if resp.StatusCode != 0 {
		return fmt.Errorf("refund error with %d,msg:%s ", order.ID, resp.Error.ErrorMessage)
	}

	return nil
}
