package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/aquasecurity/lmdrouter"
	"github.com/aws/aws-lambda-go/lambda"
)

var counterTableName string
var groupTableName string
var dbi = dynamodb_iface()

func incCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, counterTableName, groupId, counterId, dnquery(dq_current, dq_inc), 1)
		}
	}
}

func decCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, counterTableName, groupId, counterId, dnquery(dq_current, dq_dec), 1)
		}
	}
}

func getCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, counterTableName, groupId, counterId, dnquery(dq_current, dq_current), 1)
		}
	}
}

func setCounterStep(ctx context.Context, req Request) (Response, error) {
	if sv, sverr := strconv.Atoi(req.QueryStringParameters["stepVal"]); sverr != nil {
		return makeerror(sverr)
	} else {
		if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
			return makeerror(gerr)
		} else {
			log.Print("stepVal is ", sv)
			if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
				return makeerror(cerr)
			} else {
				return dynamodb_counter_operate(dbi, counterTableName, groupId, counterId, dnquery(dq_init, dq_current), sv)
			}
		}
	}
}

func resetCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, counterTableName, groupId, counterId, dnquery(dq_current, dq_init), 1)
		}
	}
}

func deleteCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_delete(dbi, groupTableName, counterTableName, groupId, counterId)
		}
	}
}

func createCounter(ctx context.Context, req Request) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		return dynamodb_counter_create(dbi, groupTableName, counterTableName, req.PathParameters["name"], groupId)
	}
}

func listCounters(ctx context.Context, req Request) (Response, error) {
	return makeerror(errors.New("NYI"))
}

func listGroups(ctx context.Context, req Request) (Response, error) {
	return makeerror(errors.New("NYI"))
}

func createGroup(ctx context.Context, req Request) (Response, error) {
	return dynamodb_group_create(dbi, groupTableName, req.PathParameters["name"])
}

func deleteGroup(ctx context.Context, req Request) (Response, error) {
	return makeerror(errors.New("NYI"))
}

func api_v1() {

	// Set up the lambda router to handle the API endpoints
	router := lmdrouter.NewRouter("/api/v1/group")
	router.Route("GET", "/", listGroups)
	router.Route("POST", "/:name", createGroup)
	router.Route("DELETE", "/:id", deleteGroup)

	router.Route("GET", "/:group/counter", listCounters)
	router.Route("GET", "/:group/counter/:id", getCounter)
	router.Route("POST", "/:group/counter/:name", createCounter)
	router.Route("POST", "/:group/counter/:id/increment", incCounter)
	router.Route("POST", "/:group/counter/:id/decrement", decCounter)
	router.Route("POST", "/:group/counter/:id/reset", resetCounter)
	router.Route("POST", "/:group/counter/:id/step", setCounterStep)
	router.Route("DELETE", "/:group/counter/:id", deleteCounter)

	counterTableName = os.Getenv("COUNTER_TABLE")
	groupTableName = os.Getenv("GROUP_TABLE")

	lambda.Start(func(ctx context.Context, req Request) (Response, error) {
		//req.Path = "/" + req.PathParameters["Path"]
		log.Print("Router called with req ", req)
		return router.Handler(ctx, req)
	})
}
