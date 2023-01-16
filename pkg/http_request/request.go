package http_request

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/google/go-querystring/query"
)

const (
	FormatJson string = "json"
	FormatXml  string = "xml"
)

type Request struct {
	NativeRequest   *http.Request
	NativeResponse  *http.Response
	ResponseStatus  int
	ResponseContent string
	GoodResponse    interface{}
	BadResponse     interface{}
	Format          string
}

func NewPost(ctx op_context.Context, url string, msg interface{}, format ...string) (*Request, error) {

	r := &Request{}
	r.Format = utils.OptionalArg(FormatJson, format...)

	c := ctx.TraceInMethod("http_request.NewPost", logger.Fields{"url": url, "format": r.Format})
	defer ctx.TraceOutMethod()

	var cmdByte []byte
	var err error

	if r.Format == FormatXml {
		cmdByte, err = xml.Marshal(msg)
		r.NativeRequest.Header.Set("Content-Type", "application/xml;charset=UTF-8")
	} else {
		cmdByte, err = json.Marshal(msg)
		r.NativeRequest.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	if err != nil {
		c.SetMessage("failed to marshal message")
		return nil, c.SetError(err)
	}

	r.NativeRequest, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(cmdByte))
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	return r, nil
}

func NewGet(ctx op_context.Context, url string, msg interface{}, format ...string) (*Request, error) {

	r := &Request{}
	r.Format = utils.OptionalArg(FormatJson, format...)

	c := ctx.TraceInMethod("http_request.NewGet", logger.Fields{"url": url, "format": r.Format})
	defer ctx.TraceOutMethod()

	var err error

	r.NativeRequest, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	v, err := query.Values(msg)
	if err != nil {
		c.SetMessage("failed to build query")
		return nil, c.SetError(err)
	}

	r.NativeRequest.URL.RawQuery = v.Encode()

	return r, nil
}

func (r *Request) Send(ctx op_context.Context) error {

	c := ctx.TraceInMethod("Request.Send", logger.Fields{"url": r.NativeRequest.URL.String(), "method": r.NativeRequest.Method})

	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	client := &http.Client{}
	r.NativeResponse, err = client.Do(r.NativeRequest)
	if err != nil {
		c.SetMessage("failed to send request")
		return err
	}
	if r.NativeResponse != nil {
		r.ResponseStatus = r.NativeResponse.StatusCode
		if r.NativeResponse.Body != nil {
			var body []byte
			body, _ = io.ReadAll(r.NativeResponse.Body)
			r.NativeResponse.Body.Close()

			r.ResponseContent = string(body)

			parseResponse := func(obj interface{}) {
				if r.Format == FormatXml {
					err = xml.Unmarshal(body, obj)
				} else {
					err = json.Unmarshal(body, obj)
				}
			}

			if r.ResponseStatus < http.StatusBadRequest {
				if r.GoodResponse != nil {
					parseResponse(r.GoodResponse)
					if err != nil {
						c.SetMessage("failed to parse good response")
					}
				}
			} else {
				if r.BadResponse != nil {
					parseResponse(r.BadResponse)
					if err != nil {
						c.SetMessage("failed to parse bad response")
					}
				}
			}

			if err != nil {
				c.Fields()["response_content"] = r.ResponseContent
				c.Fields()["response_status"] = r.ResponseStatus
				return err
			}
		}
	}

	return nil
}

func (r *Request) AddHeader(key string, value string) {
	r.NativeRequest.Header.Add(key, value)
}

func (r *Request) SetAuthHeader(key string, value string) {
	str := fmt.Sprintf("%s %s", key, value)
	r.AddHeader("Authorization", str)
}
