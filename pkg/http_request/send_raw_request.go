package http_request

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

type RedirectHandler func(req *http.Request, via []*http.Request) error

func SendRawRequest(ctx op_context.Context, request *http.Request, redirectHandler ...RedirectHandler) (*http.Response, error) {

	c := ctx.TraceInMethod("http_request.Send", logger.Fields{"url": request.URL.Path, "method": request.Method})
	defer ctx.TraceOutMethod()

	client := &http.Client{}
	if len(redirectHandler) != 0 {
		client.CheckRedirect = redirectHandler[0]
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, c.SetError(err)
	}

	return response, nil
}
