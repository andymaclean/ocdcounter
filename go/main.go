package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func unknownHandler() error {
	return errors.New("UNKNOWN HANDLER")
}

func unauthorizedHandler() error {
	return errors.New("UNAUTHORIZED HANDLER")
}

func loop(ctx context.Context, req Request, userId *UUID) (Response, error) {
	return makeresponse(req)
}

var public_handlers = map[string]func(ctx context.Context, req Request) (Response, error){
	"GET /login":  login,
	"GET /signup": signup,
}

func public_handler_gatewayv2(ctx context.Context, req Request) (Response, error) {
	f, found := public_handlers[req.RouteKey]

	if !found {
		return makeerror(fmt.Errorf("route %s not found", req.RouteKey))
	}

	return f(ctx, req)
}

var private_handlers = map[string]func(ctx context.Context, req Request, userId *UUID) (Response, error){
	"GET /loop":                 loop,
	"GET /loopua":               loop,
	"GET /api/v1/group":         listGroups,
	"POST /api/v1/group/{name}": createGroup,
	"DELETE /api/v1/group/{id}": deleteGroup,

	"GET /api/v1/group/{group}/counter":                 listCounters,
	"GET /api/v1/group/{group}/counter/{id}":            getCounter,
	"POST /api/v1/group/{group}/counter/{name}":         createCounter,
	"POST /api/v1/group/{group}/counter/{id}/increment": incCounter,
	"POST /api/v1/group/{group}/counter/{id}/decrement": decCounter,
	"POST /api/v1/group/{group}/counter/{id}/reset":     resetCounter,
	"POST /api/v1/group/{group}/counter/{id}/step":      setCounterStep,
	"DELETE /api/v1/group/{group}/counter/{id}":         deleteCounter,
}

func private_handler_gatewayv2(ctx context.Context, req Request) (Response, error) {
	if req.RequestContext.Authorizer == nil {
		return makeerror(unauthorizedHandler())
	}

	f, found := private_handlers[req.RouteKey]

	if !found {
		return makeerror(fmt.Errorf("route %s not found", req.RouteKey))
	}

	email, hasemail := req.RequestContext.Authorizer.JWT.Claims["cognito:username"]

	if !hasemail {
		return makeerror(fmt.Errorf("username is not in JWT claims"))
	}

	uuid, uerror := dynamodb_lookup_userUUID(dbi, &userTableName, &email)

	if uerror != nil {
		return makeerror(uerror)
	}

	return f(ctx, req, &uuid)
}

func main() {
	setTableNamesFromEnv()

	switch os.Getenv("_HANDLER") {
	case "apipublic":
		lambda.Start(public_handler_gatewayv2)
	case "apiprivate":
		lambda.Start(private_handler_gatewayv2)
	default:
		lambda.Start(unknownHandler)
	}
}
