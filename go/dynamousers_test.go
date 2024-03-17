package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var expUserTable = "userTable"
var expUserName = "userName"
var expUser = MakeUUID()

func TestUserCreate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_user_create(ops, &expUserTable, expUser, &expUserName)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkNewUser(t, ops[0], expUserTable, expUser, expUserName)
}

func TestUserUpdate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	nuuid := MakeUUID()

	ops, err = append_user_update(ops, &expUserTable, &expUser, "hello world", nuuid)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkUserUpdate(t, ops[0], nuuid, "hello world", expUserTable, expUser)
}
