package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

type httpApiHandler struct {
	router *mux.Router

	httpAddr                       string
	httpsAddr, httpsCert, httpsKey string
}

type HttpCore interface {
	SetHttp(httpAddr string)
	SetHttps(httpsAddr string, httpsCert string, httpsKey string)
	UseMiddleware(mwf ...mux.MiddlewareFunc)
	StartListen()
}

func InitApiCore() HttpCore {
	handler := new(httpApiHandler)
	handler.router = NewRouter(handler)
	return handler
}

type responder struct {
	r *http.Request
	w http.ResponseWriter

	start        time.Time
	statusCode   int
	headers      map[string]string
	responseBody interface{}
}

func (r *responder) setHeader(name, value string) {
	if r.headers == nil {
		r.headers = make(map[string]string)
	}

	r.headers[name] = value
}

func (r *responder) handle(rp RequestProcessor) {
	r.start = time.Now()

	bodyBts, _ := io.ReadAll(r.r.Body)
	r.r.Body.Close()

	msg := fmt.Sprintf("Request: '%s %s', body: '%s'.", r.r.Method, r.r.RequestURI, bodyBts)
	log.Debug(msg)

	r.r.Body = io.NopCloser(bytes.NewBuffer(bodyBts))

	res, err := rp(r.r)
	if err != nil {
		r.setErrors(err)
	} else {
		r.setSuccess(res)
	}

	_ = r.sendResponse()
	r.outputToLog()
}

func (r *responder) sendResponse() error {
	for name, value := range r.headers {
		r.w.Header().Set(name, value)
	}

	if r.responseBody != nil {
		r.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	r.w.WriteHeader(r.statusCode)
	if r.responseBody != nil {
		return json.NewEncoder(r.w).Encode(r.responseBody)
	}

	return nil
}

func (r *responder) outputToLog() {
	processingTime := time.Now().Sub(r.start)
	var extra string
	if respErr, ok := r.responseBody.(responseError); ok {
		extra = fmt.Sprintf("'%s'", respErr.Message)
	}

	msg := fmt.Sprintf("Response to: '%s %s', response: %d %s (took: %v).", r.r.Method, r.r.RequestURI, r.statusCode, extra, processingTime)
	log.Info(msg)
}

func (r *responder) setSuccess(obj interface{}) {
	if obj == nil {
		r.statusCode = http.StatusNoContent
	} else {
		r.statusCode = http.StatusOK
	}

	r.responseBody = obj
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Fields  string `json:"fields"`
}

func (r *responder) setErrors(err error) {
	errObj := responseError{
		http.StatusBadRequest,
		err.Error(),
		"",
	}

	if httpError, ok := err.(HttpError); ok {
		errObj.Code = httpError.StatusCode()
	}

	if fieldsError, ok := err.(FieldError); ok {
		errObj.Fields = fieldsError.Fields()
	}

	r.statusCode = errObj.Code
	r.responseBody = errObj
}

type RequestProcessor func(r *http.Request) (interface{}, error)

func (rp RequestProcessor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := responder{r: r, w: w}
	resp.handle(rp)
}

type HttpError interface {
	StatusCode() int
}

type FieldError interface {
	Fields() string
}

// Unauthorized
type invalidTokenError struct {
	error
}

func (e invalidTokenError) StatusCode() int {
	return http.StatusUnauthorized
}

func (e invalidTokenError) Fields() string {
	return "token"
}

// Forbidden
type forbiddenError struct {
	error
}

func (e forbiddenError) StatusCode() int {
	return http.StatusForbidden
}

// Bad Request
type badRequestError struct {
	error
}

func (e badRequestError) StatusCode() int {
	return http.StatusBadRequest
}

// Internal Error
type internalError struct {
	error
}

func (e internalError) StatusCode() int {
	return http.StatusInternalServerError
}

type conflictError struct {
	message      string
	conflictCode string
}

func (e conflictError) StatusCode() int {
	return http.StatusConflict
}

func (e conflictError) Fields() string {
	return e.conflictCode
}

func (e conflictError) Error() string {
	return e.message
}

// Base error
type baseHttpError struct {
	error
	statusCode int
	fields     string
}

func (e baseHttpError) StatusCode() int {
	return e.statusCode
}

func (e baseHttpError) Fields() string {
	return e.fields
}

func NewRouter(h *httpApiHandler) *mux.Router {
	router := mux.NewRouter().StrictSlash(false)

	for _, r := range h.getRoutes() {
		router.Methods(r.Method).
			Path(r.Pattern).
			Name(r.Name).
			Handler(r.HandlerFunc)
	}

	for _, r := range h.getWSRoutes() {
		router.Handle(r.Path, websocket.Handler(r.HandlerFunc))
	}

	return router
}

func (h *httpApiHandler) SetHttp(httpAddr string) {
	h.httpAddr = httpAddr
}

func (h *httpApiHandler) SetHttps(httpsAddr string, httpsCert string, httpsKey string) {
	h.httpsAddr = httpsAddr
	h.httpsCert = httpsCert
	h.httpsKey = httpsKey
}

func (h *httpApiHandler) UseMiddleware(mwf ...mux.MiddlewareFunc) {
	h.router.Use(mwf...)
}

func (h *httpApiHandler) StartListen() {
	if h.httpAddr == "" && h.httpsAddr == "" {
		log.Fatal("Either HTTP or/and HTTPS must be enabled")
	}

	if h.httpAddr != "" {
		go func() {
			server := &http.Server{Addr: h.httpAddr, Handler: h.router}
			log.Infof("HTTP: Listening on addr %s", server.Addr)
			err := server.ListenAndServe()
			if err != nil {
				log.Fatalf("Could not start HTTP listener. %v\n", err)
			}
		}()
	}

	if h.httpsCert != "" && h.httpsKey != "" {
		go func() {
			server := &http.Server{Addr: h.httpsAddr, Handler: h.router}
			log.Infof("HTTPS: Listening on addr %s", server.Addr)
			err := server.ListenAndServeTLS(h.httpsCert, h.httpsKey)
			if err != nil {
				log.Fatalf("Could not start HTTPS listener. %v\n", err)
			}
		}()
	} else {
		log.Warning("To enable HTTPS server you must provide both cert and key file.")
	}
}
