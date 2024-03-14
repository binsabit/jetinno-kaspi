package pkg

import (
	"bytes"
	"context"
	"github.com/bytedance/sonic"
	"io"
	"net/http"
	"time"
)

type Request struct {
	URL    string
	Header map[string]string
	Method string
	Data   []byte
}

func NewRequest(url string, method string, header map[string]string, data interface{}) (*Request, error) {

	request := &Request{
		URL:    url,
		Method: method,
		Header: header,
	}
	if data != nil {
		reqData, err := sonic.ConfigFastest.Marshal(data)
		if err != nil {
			return nil, err
		}
		request.Data = reqData
	}

	return request, nil
}

func (r Request) Do() ([]byte, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	request, err := http.NewRequest(http.MethodPost, r.URL, bytes.NewBuffer(r.Data))
	if err != nil {
		return nil, err
	}

	for key, val := range r.Header {
		request.Header.Set(key, val)
	}

	request.WithContext(ctx)

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
