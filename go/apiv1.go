package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
)

var counterTableName string
var groupTableName string
var userTableName string
var dbi = dynamodb_iface()

func setTableNamesFromEnv() {
	counterTableName = os.Getenv("COUNTER_TABLE")
	groupTableName = os.Getenv("GROUP_TABLE")
	userTableName = os.Getenv("USER_TABLE")
}

func incCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, &counterTableName, groupId, counterId, dnquery(dq_current, dq_inc), 1)
		}
	}
}

func decCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, &counterTableName, groupId, counterId, dnquery(dq_current, dq_dec), 1)
		}
	}
}

func getCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_read_counter(dbi, &counterTableName, groupId, counterId)
		}
	}
}

func setCounterStep(ctx context.Context, req Request, userId *UUID) (Response, error) {
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
				return dynamodb_counter_operate(dbi, &counterTableName, groupId, counterId, dnquery(dq_init, dq_current), sv)
			}
		}
	}
}

func resetCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_operate(dbi, &counterTableName, groupId, counterId, dnquery(dq_current, dq_init), 1)
		}
	}
}

func deleteCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		if counterId, cerr := ToUUID(req.PathParameters["id"]); cerr != nil {
			return makeerror(cerr)
		} else {
			return dynamodb_counter_delete(dbi, &groupTableName, &counterTableName, groupId, counterId)
		}
	}
}

func createCounter(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		return dynamodb_counter_create(dbi, &groupTableName, &counterTableName, req.PathParameters["name"], groupId)
	}
}

func listCounters(ctx context.Context, req Request, userId *UUID) (Response, error) {
	if groupId, gerr := ToUUID(req.PathParameters["group"]); gerr != nil {
		return makeerror(gerr)
	} else {
		return dynamodb_counter_list(dbi, &groupTableName, groupId)
	}
}

func listGroups(ctx context.Context, req Request, userId *UUID) (Response, error) {
	return dynamodb_group_list(dbi, &userTableName, userId)
}

func createGroup(ctx context.Context, req Request, userId *UUID) (Response, error) {
	return dynamodb_group_create(dbi, &userTableName, &groupTableName, userId, req.PathParameters["name"])
}

func deleteGroup(ctx context.Context, req Request, userId *UUID) (Response, error) {
	return makeerror(errors.New("NYI"))
}
