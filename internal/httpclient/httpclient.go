package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultTimeout = 10 * time.Second
const MaxBodyBytes = 1 << 20

type Response struct {
	StatusCode int
	Body       []byte
}

type Client struct {
	HTTP *http.Client
}

func New(c *http.Client) *Client {
	if c == nil {
		c = &http.Client{Timeout: DefaultTimeout}
	}
	return &Client{HTTP: c}
}

func (c *Client) Do(request *http.Request) (Response, error) {
	response, err := c.HTTP.Do(request)
	if err != nil {
		return Response{}, err
	}
	defer response.Body.Close()
	body, err := readBody(response.Body)
	if err != nil {
		return Response{}, fmt.Errorf("httpclient: reading response from %s: %w", request.URL, err)
	}
	return Response{
		StatusCode: response.StatusCode,
		Body:       body,
	}, nil
}

func readBody(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, MaxBodyBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxBodyBytes {
		return nil, fmt.Errorf("body exceeds %d bytes", MaxBodyBytes)
	}
	return data, nil
}
