package http_request

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

func SendRawRequest(ctx op_context.Context, request *http.Request) (*http.Response, error) {

	c := ctx.TraceInMethod("http_request.Send", logger.Fields{"path": request.URL.Path, "method": request.Method})
	defer ctx.TraceOutMethod()

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, c.SetError(err)
	}

	return response, nil
}
