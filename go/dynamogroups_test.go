package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var expGroupTable = "groupTable"
var expGroupName = "groupName"

func TestGroupCreate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_group_create(ops, &expGroupTable, expGroup, expGroupName)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkNewGroup(t, ops[0], expGroupTable, expGroup, expGroupName)
}

func TestGroupUpdate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	nuuid := MakeUUID()

	ops, err = append_group_update(ops, &expGroupTable, &expGroup, "hello world", nuuid)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkGroupUpdate(t, ops[0], nuuid, "hello world", expGroupTable, expGroup, expUser)
}
