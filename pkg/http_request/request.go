package http_request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/message"
	"github.com/evgeniums/go-backend-helpers/pkg/message/message_json"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/gorilla/schema"
)

const MaxDumpSize int = 2048

type Request struct {
	NativeRequest   *http.Request
	NativeResponse  *http.Response
	ResponseStatus  int
	Body            []byte
	ResponseContent string
	GoodResponse    interface{}
	BadResponse     interface{}
	Serializer      message.Serializer
	Transport       http.RoundTripper
	Timeout         int
	ParsingFailed   bool
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

	r.Body = cmdByte
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
		encoder.RegisterEncoder(utils.TimeNil, utils.TimeReflectStr)
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

	if strings.Contains(uRL, "?") {
		err = errors.New("URL must not contain query string, encode query in msg object")
		return nil, err
	}

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

	if ctx.Logger().DumpRequests() {
		dump, err1 := httputil.DumpRequestOut(r.NativeRequest, true)
		if err1 != nil {
			c.Logger().Error("Failed to dump HTTP request", err1)
		} else {
			c.Logger().Debug("Client dump HTTP request", logger.Fields{"http_request": string(dump)})
		}
	}

	client := &http.Client{Transport: r.Transport}
	if r.Timeout != 0 {
		client.Timeout = time.Second * time.Duration(r.Timeout)
	}
	r.NativeResponse, err = client.Do(r.NativeRequest)

	if ctx.Logger().DumpRequests() {
		if r.NativeResponse != nil {
			dump, err1 := httputil.DumpResponse(r.NativeResponse, true)
			if err1 != nil {
				c.Logger().Error("failed to dump HTTP response", err1)
			} else {
				// TODO make it singleton
				dumpStr := string(dump)
				text := utils.Substr(string(dump), 0, MaxDumpSize)
				if len(text) < len(dumpStr) {
					text = utils.ConcatStrings(text, "...")
				}
				c.Logger().Debug("Client dump HTTP response", logger.Fields{"http_response": text})
			}
		} else {
			c.Logger().Debug("Client dump HTTP response", logger.Fields{"http_response": ""})
		}
	}

	if err != nil {
		c.SetMessage("failed to client.Do")
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
				if err != nil {
					r.ParsingFailed = true
				}
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
				// set logger fields only if parsing fails
				if !ctx.Logger().DumpRequests() {
					text := utils.Substr(r.ResponseContent, 0, MaxDumpSize)
					if len(text) < len(r.ResponseContent) {
						text = utils.ConcatStrings(text, "...")
					}
					c.SetLoggerField("response_content", text)
				}
				c.SetLoggerField("response_status", r.ResponseStatus)
				return err
			}
		}
	}

	return nil
}

func (r *Request) SetHeader(key string, value string) {
	r.NativeRequest.Header.Set(key, value)
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

func NewMultipart(ctx op_context.Context, url string, files map[string]io.Reader, fields map[string]string, filesField ...string) (*Request, error) {

	// prepare
	filesFieldName := utils.OptionalArg("files", filesField...)
	r := &Request{}
	c := ctx.TraceInMethod("http_request.NewMultipart", logger.Fields{"url": url})
	defer ctx.TraceOutMethod()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// write other fields
	for key, r := range fields {
		err := w.WriteField(key, r)
		if err != nil {
			c.SetLoggerField("field", key)
			c.SetMessage("failed to create form field")
			return nil, c.SetError(err)
		}
	}

	// write files
	for key, r := range files {
		fw, err := w.CreateFormFile(filesFieldName, key)
		if err != nil {
			c.SetLoggerField("file", key)
			c.SetMessage("failed to create form file")
			return nil, c.SetError(err)
		}
		_, err = io.Copy(fw, r)
		if err != nil {
			c.SetLoggerField("file", key)
			c.SetMessage("failed to read file")
			return nil, c.SetError(err)
		}
	}
	w.Close()

	// create request
	var err error
	r.NativeRequest, err = http.NewRequest(http.MethodPost, url, &b)
	if err != nil {
		c.SetMessage("failed to create request")
		return nil, c.SetError(err)
	}

	// set content type
	r.NativeRequest.Header.Set("Content-Type", w.FormDataContentType())

	// done
	return r, nil
}
