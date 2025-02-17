package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type GroupData struct {
	GroupId    string   `dynamodbav:"objectUUID"`
	Counters   []string `dynamodbav:"counters,stringset,omitempty"`
	GroupName  string   `dynamodbav:"groupName"`
	ObjectType string   `dynamodbav:"objectType"`
}

func append_group_create(ops []*dynamodb.TransactWriteItem, table *string, groupUUID UUID, groupName string) ([]*dynamodb.TransactWriteItem, error) {
	record, rerr := dynamodbattribute.MarshalMap(GroupData{
		GroupId:    groupUUID.String(),
		ObjectType: "Group",
		GroupName:  groupName,
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

func append_group_update(ops []*dynamodb.TransactWriteItem, table *string, groupId *UUID, query string, val1 UUID) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol:    {S: aws.String(groupId.String())},
			objectTypeCol: {S: aws.String("Group")},
		},
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

func group_mark_delete(ops []*dynamodb.TransactWriteItem, table *string, groupId *UUID) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol:    {S: aws.String(groupId.String())},
			objectTypeCol: {S: aws.String("Group")},
		},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":true": {BOOL: aws.Bool(true)},
			":type": {S: aws.String("Group")},
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

func group_fulldelete(ops []*dynamodb.TransactWriteItem, table *string, groupId *UUID) ([]*dynamodb.TransactWriteItem, error) {
	dr := dynamodb.Delete{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol: {S: aws.String(groupId.String())}},
		TableName: table,
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Delete: &dr,
	})

	return ops, nil
}

const (
	gr_add_ctr    = iota
	gr_remove_ctr = iota
)

func gquery(mode int) string {
	switch mode {
	case gr_add_ctr:
		return "ADD counters :val1"
	case gr_remove_ctr:
		return "DELETE counters :val1"
	}
	return ""
}
