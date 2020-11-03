package http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/muhammadisa/go-kit-boilerplate/middleware"
	"github.com/muhammadisa/go-kit-boilerplate/services/user/delivery"
	"github.com/muhammadisa/go-kit-boilerplate/utils/decodeencode"
)

// NewHTTPServe create http server with go standard lib
func NewHTTPServe(
	_ context.Context,
	svcEndpoints delivery.Endpoints,
	logger log.Logger,
) http.Handler {
	// Initialize mux router error logger and error
	r := mux.NewRouter()
	var options []httptransport.ServerOption
	errorLogger := httptransport.ServerErrorLogger(logger)
	errorEncoder := httptransport.ServerErrorEncoder(decodeencode.EncodeErrorResponse)
	options = append(options, errorLogger, errorEncoder)

	// Attaching middlewares
	r.Use(middleware.ContentTypeMiddleware)
	r.Use(middleware.AllowOrigin)
	r.Use(mux.CORSMethodMiddleware(r))

	// Creating routes
	r.Methods("POST").Path("/user/register").Handler(httptransport.NewServer(
		svcEndpoints.Register,
		decodeRegisterRequest,
		decodeencode.EncodeResponse,
		options...,
	))
	r.Methods("POST").Path("/user/login").Handler(httptransport.NewServer(
		svcEndpoints.Login,
		decodeLoginRequest,
		decodeencode.EncodeResponse,
		options...,
	))

	return r
}

func decodeRegisterRequest(
	_ context.Context,
	r *http.Request,
) (interface{}, error) {
	var req delivery.CreateRegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func decodeLoginRequest(
	_ context.Context,
	r *http.Request,
) (interface{}, error) {
	var req delivery.CreateLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
