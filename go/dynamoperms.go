package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type PermData struct {
	PrincipalId  string   `dynamodbav:"userUUID"`
	ObjectTypeId string   `dynamodbsv:"ObjectTypeUUID"`
	Rights       []string `dynamodbav:"rights,stringset,omitempty"`
}

var perm_read = "read"
var perm_inc = "inc"
var perm_dec = "dec"
var perm_config = "config"
var perm_admin = "admin"
var perm_create = "create"
var perm_delete = "delete"

func update_rights(ops []*dynamodb.TransactWriteItem, table *string, userId *UUID, objectType *string, objectId *UUID, query *string, rights []*string) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			principalIdCol: {S: aws.String(userId.String())},
		},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":vals": {SS: rights},
		},
		UpdateExpression: query,
	}

	ops = append(ops, &dynamodb.TransactWriteItem{
		Update: &udr,
	})

	return ops, nil
}

const (
	pm_add_rights    = iota
	pm_remove_rights = iota
)

func pquery(mode int) string {
	switch mode {
	case gr_add_ctr:
		return "ADD rights :vals"
	case gr_remove_ctr:
		return "DELETE rights :vals"
	}
	return ""
}
