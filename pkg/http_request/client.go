package http_request

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
)

type HttpClientConfig struct {
	TIMEOUT                  int `default:"15"`
	KEEP_ALIVE               int `default:"30"`
	MAX_IDLE_CONNECTIONS     int `default:"100"`
	IDLE_CONNECTIONS_TIMEOUT int `default:"90"`
	TLS_HANDSHAKE_TIMEOUT    int `default:"10"`
	EXPECT_CONTINUE_TIMEOUT  int `default:"1"`
}

type HttpClient struct {
	HttpClientConfig
	httpClient *http.Client
	context    context.Context
	cancel     context.CancelFunc
}

func (c *HttpClient) Config() interface{} {
	return &c.HttpClientConfig
}

func NewHttpClient(transport ...http.RoundTripper) *HttpClient {
	h := &HttpClient{}
	h.httpClient = &http.Client{Transport: utils.OptionalArg(nil, transport...)}
	h.context, h.cancel = context.WithCancel(context.Background())
	return h
}

func DefaultHttpClient() *HttpClient {
	return NewHttpClient(http.DefaultTransport)
}

func transportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func (h *HttpClient) Init(cfg config.Config, log logger.Logger, vld validator.Validator, configPath ...string) error {

	var err error

	if h.httpClient.Transport != nil {
		log.Info("Configuration: using transport from constructor")
		return nil
	}

	key := utils.OptionalString("http_client", configPath...)
	if cfg.IsSet(key) {
		err = object_config.LoadLogValidate(cfg, log, vld, h, key)
	} else {
		err = object_config.LoadLogValidate(cfg, log, vld, h, "http_client")
	}
	if err != nil {
		return log.PushFatalStack("failed to load configuration of http client", err)
	}

	h.httpClient.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: transportDialContext(&net.Dialer{
			Timeout:   time.Duration(h.TIMEOUT) * time.Second,
			KeepAlive: time.Duration(h.KEEP_ALIVE) * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          h.MAX_IDLE_CONNECTIONS,
		IdleConnTimeout:       time.Duration(h.IDLE_CONNECTIONS_TIMEOUT) * time.Second,
		TLSHandshakeTimeout:   time.Duration(h.TLS_HANDSHAKE_TIMEOUT) * time.Second,
		ExpectContinueTimeout: time.Duration(h.EXPECT_CONTINUE_TIMEOUT) * time.Second,
	}

	return nil
}

func (h *HttpClient) NewPost(ctx op_context.Context, url string, msg interface{}, serializer ...message.Serializer) (*Request, error) {
	req, err := NewPostWithContext(h.context, ctx, url, msg)
	if err != nil {
		return nil, err
	}
	req.client = h.httpClient
	return req, nil
}

func (h *HttpClient) NewGet(ctx op_context.Context, url string, msg interface{}) (*Request, error) {
	req, err := NewGetWithContext(h.context, ctx, url, msg)
	if err != nil {
		return nil, err
	}
	req.client = h.httpClient
	return req, nil
}

func (h *HttpClient) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(h.context, method, url, body)
}

func (h *HttpClient) Shutdown(ctx context.Context) error {
	if h.cancel != nil {
		h.cancel()
	}
	return nil
}
