package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type GroupData struct {
	GroupId   string   `dynamodbav:"groupUUID"`
	Counters  []string `dynamodbav:"counters,stringset,omitempty"`
	GroupName string   `dynamodbav:"groupName"`
}

func group_create(ops []*dynamodb.TransactWriteItem, table *string, groupUUID UUID, groupName string) ([]*dynamodb.TransactWriteItem, error) {
	record, rerr := dynamodbattribute.MarshalMap(GroupData{
		GroupId:   groupUUID.String(),
		GroupName: groupName,
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

func group_update(ops []*dynamodb.TransactWriteItem, table *string, group UUID, query string, val1 UUID) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol: {S: aws.String(group.String())}},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":val1": {SS: []*string{aws.String(val1.String())}},
		},
		UpdateExpression:    aws.String(query),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_exists(%s) and attribute_not_exists(%s)", groupIdCol, deleteMarkerCol)),
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Update: &udr,
	})

	return ops, nil
}

func group_mark_delete(ops []*dynamodb.TransactWriteItem, table *string, group UUID) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol: {S: aws.String(group.String())}},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":true": {BOOL: aws.Bool(true)},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("SET %s = :true", deleteMarkerCol)),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_exists(%s)", groupIdCol)),
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Update: &udr,
	})

	return ops, nil
}

func group_fulldelete(ops []*dynamodb.TransactWriteItem, table *string, group UUID) ([]*dynamodb.TransactWriteItem, error) {
	dr := dynamodb.Delete{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol: {S: aws.String(group.String())}},
		TableName: table,
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Delete: &dr,
	})

	return ops, nil
}

const (
	gr_add    = iota
	gr_remove = iota
)

func gquery(mode int) string {
	switch mode {
	case gr_add:
		return "ADD counters :val1"
	case gr_remove:
		return "DELETE counters :val1"
	}
	return ""
}
