package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
)

func incCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
		return makeerror(cerr)
	} else {
		return dbo.CounterUpdate(s, counterId, dnquery(dq_current, dq_inc), 1)
	}
}

func decCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
		return makeerror(cerr)
	} else {
		return dbo.CounterUpdate(s, counterId, dnquery(dq_current, dq_dec), 1)
	}
}

func getCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
		return makeerror(cerr)
	} else {
		return dbo.CounterRead(s, counterId)
	}
}

func setCounterStep(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if sv, sverr := strconv.Atoi(req.QueryStringParameters["stepVal"]); sverr != nil {
		return makeerror(sverr)
	} else {
		log.Print("stepVal is ", sv)
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dbo.CounterUpdate(s, counterId, dnquery(dq_init, dq_current), sv)
		}
	}
}

func resetCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
		return makeerror(cerr)
	} else {
		return dbo.CounterUpdate(s, counterId, dnquery(dq_current, dq_init), 1)
	}
}

func deleteCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
		return makeerror(cerr)
	} else {
		return dbo.CounterDelete(s, counterId)
	}
}

func createCounter(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return dbo.CounterCreate(s, req.PathParameters["name"])
}

func listCounters(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return dbo.CounterList(s)
}

func listGroups(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return dbo.GroupList(s)
}

func createGroup(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return dbo.GroupCreate(s, req.PathParameters["name"])
}

func deleteGroup(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return makeerror(errors.New("NYI"))
}

func unauthorizedHandler() error {
	return errors.New("UNAUTHORIZED HANDLER")
}

func loop(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error) {
	return makeresponse(req)
}

var public_handlers = map[string]func(ctx context.Context, req Request, dbo DataOperator) (Response, error){
	"GET /login":  login,
	"GET /signup": signup,
}

type APIHandler struct {
	dbo DataOperator
}

func (api APIHandler) public_handler_gatewayv2(ctx context.Context, req Request) (Response, error) {
	f, found := public_handlers[req.RouteKey]

	if !found {
		return makeerror(fmt.Errorf("route %s not found", req.RouteKey))
	}

	return f(ctx, req, api.dbo)
}

var private_handlers = map[string]func(ctx context.Context, req Request, dbo DataOperator, s Session) (Response, error){
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

func (api APIHandler) private_handler_gatewayv2(ctx context.Context, req Request) (Response, error) {
	if req.RequestContext.Authorizer == nil {
		return makeerror(unauthorizedHandler())
	}

	f, found := private_handlers[req.RouteKey]

	if !found {
		return makeerror(fmt.Errorf("route %s not found", req.RouteKey))
	}
	session, serr := Create_APISession(api.dbo, req)

	if serr != nil {
		return makeerror(serr)
	}

	return f(ctx, req, api.dbo, session)
}
