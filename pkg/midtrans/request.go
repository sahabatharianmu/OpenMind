package midtrans

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type requestOptions struct {
	ctx          context.Context
	method       string
	url          string
	headers      map[string]string
	body         []byte
	responseDest interface{}
}

func (m *Service) doRequest(opts requestOptions) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(opts.url)
	req.Header.SetMethod(opts.method)

	for k, v := range opts.headers {
		req.Header.Set(k, v)
	}
	if len(opts.body) > 0 {
		req.Header.SetContentType("application/json")
		req.SetBody(opts.body)
	}

	m.logger.Debug("Executing HTTP request", zap.String("method", opts.method), zap.String("url", opts.url))

	done := make(chan error, 1)
	go func() {
		done <- m.client.Do(req, resp)
	}()

	select {
	case <-opts.ctx.Done():
		return nil, opts.ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	m.logger.Debug(
		"Received HTTP response",
		zap.Int("status_code", resp.StatusCode()),
		zap.String("body", string(resp.Body())),
	)

	if opts.responseDest != nil {
		if err := sonic.Unmarshal(resp.Body(), opts.responseDest); err != nil {
			return resp, fmt.Errorf("failed to parse response (body: %s): %w", string(resp.Body()), err)
		}
	}

	return resp, nil
}
