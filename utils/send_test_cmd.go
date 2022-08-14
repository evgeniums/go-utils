package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-querystring/query"
)

func SendTestPost(t *testing.T, g *gin.Engine, path string, cmd interface{}, auths ...string) (int, string) {
	_, responseCode, responseMessage := SendTestPostResponse(t, g, path, cmd, auths...)
	return responseCode, responseMessage
}

func SendTestPostResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, auths ...string) (*httptest.ResponseRecorder, int, string) {
	return SendTestRequestResponse(t, g, http.MethodPost, path, cmd, auths...)
}

func SendTestRequest(t *testing.T, g *gin.Engine, method string, path string, cmd interface{}, auths ...string) (int, string) {
	_, responseCode, responseMessage := SendTestRequestResponse(t, g, method, path, cmd, auths...)
	return responseCode, responseMessage
}

func PrepareTestRequest(t *testing.T, g *gin.Engine, method string, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder, noCookie ...bool) (*http.Request, string) {

	var cmdStr string
	var content io.Reader
	if cmd != nil {
		cmdBytes, _ := json.Marshal(cmd)
		cmdStr = string(cmdBytes)
		content = bytes.NewBuffer(cmdBytes)
	}
	req, err := http.NewRequest(method, path, content)
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}

	if len(noCookie) == 0 {
		cookies := prevResponse.Result().Cookies()
		for i := range cookies {
			t.Logf("Cookie: %v", cookies[i])
			req.AddCookie(cookies[i])
		}
	}

	return req, cmdStr
}

func SendPreparedTestRequestResponse(t *testing.T, g *gin.Engine, req *http.Request) (*httptest.ResponseRecorder, int, string) {
	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestRequestResponse(t *testing.T, g *gin.Engine, method string, path string, cmd interface{}, auths ...string) (*httptest.ResponseRecorder, int, string) {

	var content io.Reader
	if cmd != nil {
		cmdStr, _ := json.Marshal(cmd)
		content = bytes.NewBuffer(cmdStr)
	}
	req, err := http.NewRequest(method, path, content)
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	if len(auths) > 0 {
		req.Header.Set("Authorization", auths[0])
	}
	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestPostOnResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder) (int, string) {
	_, responseCode, responseMessage := SendTestPostOnResponseResponse(t, g, path, cmd, prevResponse)
	return responseCode, responseMessage
}

func SendTestPostOnResponseResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder) (*httptest.ResponseRecorder, int, string) {
	cmdStr, _ := json.Marshal(cmd)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(cmdStr))
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	cookies := prevResponse.Result().Cookies()
	for i := range cookies {
		t.Logf("Cookie: %v", cookies[i])
		req.AddCookie(cookies[i])
	}

	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func CheckTestPost(t *testing.T, g *gin.Engine, path string, cmd interface{}, sampleCode int, sampleMessage string) {
	responseCode, responseMessage := SendTestPost(t, g, path, cmd)
	CheckTestResponse(t, responseCode, responseMessage, sampleCode, sampleMessage)
}

func PrepareTestGet(t *testing.T, g *gin.Engine, path string, cmd interface{}) *http.Request {

	v, _ := query.Values(cmd)

	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.URL.RawQuery = v.Encode()
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	return req
}

func SendPreparedTestGet(t *testing.T, g *gin.Engine, req *http.Request) (*httptest.ResponseRecorder, int, string) {
	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestGetResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}) (*httptest.ResponseRecorder, int, string) {

	v, _ := query.Values(cmd)

	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.URL.RawQuery = v.Encode()
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestGet(t *testing.T, g *gin.Engine, path string, cmd interface{}) (int, string) {
	_, responseCode, responseMessage := SendTestGetResponse(t, g, path, cmd)
	return responseCode, responseMessage
}

func SendTestGetOnResponseResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder) (*httptest.ResponseRecorder, int, string) {

	v, _ := query.Values(cmd)

	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.URL.RawQuery = v.Encode()
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	cookies := prevResponse.Result().Cookies()
	for i := range cookies {
		t.Logf("Cookie: %v", cookies[i])
		req.AddCookie(cookies[i])
	}

	if cmd != nil {
		t.Logf("Cmd: %+v, Request query: %v", cmd, req.URL.RawQuery)
	}

	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestGetOnResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder) (int, string) {

	req, err := http.NewRequest(http.MethodGet, path, nil)

	if cmd != nil {
		v, _ := query.Values(cmd)
		req.URL.RawQuery = v.Encode()
	}

	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	cookies := prevResponse.Result().Cookies()
	for i := range cookies {
		t.Logf("Cookie: %v", cookies[i])
		req.AddCookie(cookies[i])
	}

	if cmd != nil {
		t.Logf("Cmd: %+v, Request query: %v", cmd, req.URL.RawQuery)
	}

	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return responseCode, responseMessage
}

func CheckTestResponse(t *testing.T, responseCode int, responseMessage string, sampleCode int, sampleMessage string) {
	if responseMessage != sampleMessage {
		t.Fatalf("Invalid response message: expected %v, got %v", sampleMessage, responseMessage)
	}
	if responseCode != sampleCode {
		t.Fatalf("Invalid response code: expected %d, got %d", sampleCode, responseCode)
	}
}

func CheckTestStatus(t *testing.T, responseCode int, sampleCode int) {
	if responseCode != sampleCode {
		t.Fatalf("Invalid response code: expected %d, got %d", sampleCode, responseCode)
	}
}

func SendTestFileOnResponseResponse(t *testing.T, g *gin.Engine, path string, fileName string, prevResponse *httptest.ResponseRecorder, auths ...string) (*httptest.ResponseRecorder, int, string) {

	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("Failed to open file: %s", err)
	}
	defer file.Close()

	req, err := http.NewRequest(http.MethodPost, path, file)
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Disposition", "attachment; filename="+filepath.Base(fileName))

	if len(auths) > 0 {
		req.Header.Set("Authorization", auths[0])
	}

	cookies := prevResponse.Result().Cookies()
	for i := range cookies {
		req.AddCookie(cookies[i])
	}

	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return response, responseCode, responseMessage
}

func SendTestDeleteOnResponse(t *testing.T, g *gin.Engine, path string, cmd interface{}, prevResponse *httptest.ResponseRecorder) (int, string) {

	v, _ := query.Values(cmd)

	req, err := http.NewRequest(http.MethodDelete, path, nil)
	req.URL.RawQuery = v.Encode()
	if err != nil {
		t.Fatalf("Couldn't create request: %s", err)
	}
	req.Header.Set("User-Agent", "ExampleTest")
	cookies := prevResponse.Result().Cookies()
	for i := range cookies {
		req.AddCookie(cookies[i])
	}

	if cmd != nil {
		t.Logf("Cmd: %+v, Request query: %v", cmd, req.URL.RawQuery)
	}

	response := httptest.NewRecorder()
	g.ServeHTTP(response, req)
	responseCode := response.Code
	responseMessage := response.Body.String()
	return responseCode, responseMessage
}
