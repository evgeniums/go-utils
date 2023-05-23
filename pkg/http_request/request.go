package http_request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gorilla/schema"
)

type Request struct {
	NativeRequest   *http.Request
	NativeResponse  *http.Response
	ResponseStatus  int
	ResponseContent string
	GoodResponse    interface{}
	BadResponse     interface{}
	Serializer      message.Serializer
	Transport       http.RoundTripper
}

func NewPost(ctx op_context.Context, url string, msg interface{}, serializer ...message.Serializer) (*Request, error) {

	r := &Request{}
	r.Serializer = utils.OptionalArg[message.Serializer](message_json.Serializer, serializer...)

	c := ctx.TraceInMethod("http_request.NewPost", logger.Fields{"url": url})
	defer ctx.TraceOutMethod()

	var cmdByte []byte
	var err error

	cmdByte, err = r.Serializer.SerializeMessage(msg)
	if err != nil {
		c.SetMessage("failed to marshal message")
		return nil, c.SetError(err)
	}

	r.NativeRequest, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(cmdByte))
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	if r.Serializer.ContentMime() != "" {
		r.NativeRequest.Header.Set("Content-Type", utils.ConcatStrings(r.Serializer.ContentMime(), ";charset=UTF-8"))
	}

	return r, nil
}

func UrlEncode(msg interface{}) (string, error) {
	if msg != nil {
		encoder := schema.NewEncoder()
		encoder.SetAliasTag("json")
		encoder.RegisterEncoder(utils.DateNil, utils.DateReflectStr)
		v := url.Values{}
		err := encoder.Encode(msg, v)
		if err != nil {
			return "", err
		}
		return v.Encode(), nil
	}
	return "", nil
}

func NewGet(ctx op_context.Context, uRL string, msg interface{}) (*Request, error) {

	r := &Request{}

	c := ctx.TraceInMethod("http_request.NewGet", logger.Fields{"url": uRL})
	defer ctx.TraceOutMethod()

	var err error

	r.NativeRequest, err = http.NewRequest(http.MethodGet, uRL, nil)
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	query, err := UrlEncode(msg)
	if err != nil {
		c.SetMessage("failed to build query")
		return nil, c.SetError(err)
	}

	r.NativeRequest.URL.RawQuery = query

	return r, nil
}

func (r *Request) Send(ctx op_context.Context, relaxedParsing ...bool) error {

	c := ctx.TraceInMethod("Request.Send", logger.Fields{"url": r.NativeRequest.URL.String(), "method": r.NativeRequest.Method})

	var err error
	onExit := func() {
		if err != nil {
			c.SetError(err)
		}
		ctx.TraceOutMethod()
	}
	defer onExit()

	client := &http.Client{Transport: r.Transport}
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
				if r.Serializer == nil {
					r.Serializer = message_json.Serializer
				}
				err = r.Serializer.ParseMessage(body, obj)
			}

			if r.ResponseStatus < http.StatusBadRequest {
				if r.GoodResponse != nil {
					parseResponse(r.GoodResponse)
					if err != nil {
						if !utils.OptionalArg(false, relaxedParsing...) {
							c.SetMessage("failed to parse good response")
						} else {
							err = nil
						}
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
				c.LoggerFields()["response_content"] = r.ResponseContent
				c.LoggerFields()["response_status"] = r.ResponseStatus
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
	r.NativeRequest.Header.Set("Authorization", str)
}

func HttpHeadersSet(req *http.Request, headers ...map[string]string) {
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}
}
