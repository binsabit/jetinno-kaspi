package pkg

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

type Request struct {
	URL    string
	Header map[string]string
	Method string
	Data   interface{}
}

func (r Request) Do() ([]byte, error) {
	js, err := sonic.ConfigFastest.Marshal(r.Data)
	if err != nil {
		return nil, err
	}

	agent := fiber.AcquireAgent()
	agent.Request().Header.SetMethod(r.Method)
	for key, val := range r.Header {
		agent.Set(key, val)
	}

	agent.Request().SetRequestURI(r.URL)
	agent.Body(js)

	err = agent.Parse()

	if err != nil {
		return nil, err
	}
	_, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return body, nil
}
