package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type UserData struct {
	UserId     string   `dynamodbav:"objectUUID"`
	Groups     []string `dynamodbav:"groups,stringset,omitempty"`
	UserName   string   `dynamodbav:"userEmail"`
	ObjectType string   `dynamodbav:"objectType"`
}

func user_create(ops []*dynamodb.TransactWriteItem, table *string, userId UUID, userName *string) ([]*dynamodb.TransactWriteItem, error) {
	record, rerr := dynamodbattribute.MarshalMap(UserData{
		UserId:     userId.String(),
		UserName:   *userName,
		ObjectType: "User",
	})

	if rerr != nil {
		return ops, rerr
	}

	input := dynamodb.Put{
		TableName: table,
		Item:      record,
	}

	ops = append(ops, &dynamodb.TransactWriteItem{
		Put: &input,
	})

	return ops, nil
}

func user_update(ops []*dynamodb.TransactWriteItem, table *string, user *UUID, query string, val1 UUID) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			userIdCol:     {S: aws.String(user.String())},
			objectTypeCol: {S: aws.String("User")},
		},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val1": {SS: []*string{aws.String(val1.String())}},
		},
		UpdateExpression:    aws.String(query),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_exists(%s) and attribute_not_exists(%s)", userIdCol, deleteMarkerCol)),
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Update: &udr,
	})

	return ops, nil
}

const (
	usr_add_grp    = iota
	usr_remove_grp = iota
)

func uquery(mode int) string {
	switch mode {
	case usr_add_grp:
		return "ADD groups :val1"
	case usr_remove_grp:
		return "DELETE groups :val1"
	}
	return ""
}
