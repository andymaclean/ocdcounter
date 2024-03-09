package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/events"
)

type Response = events.APIGatewayV2HTTPResponse
type Request = events.APIGatewayV2HTTPRequest

type CountData struct {
	CounterId    string `json:"objectUUID"`
	CounterName  string `json:"counterName"`
	CounterGroup string `json:"counterGroupUUID"`
	CounterVal   int    `json:"countVal"`
	StepVal      int    `json:"stepVal"`
	ObjectType   string `json:"objectType"`
}

func counter_create(ops []*dynamodb.TransactWriteItem, table *string, counterUUID UUID, counterName string, groupId *UUID) ([]*dynamodb.TransactWriteItem, error) {
	record, rerr := dynamodbattribute.MarshalMap(CountData{
		CounterId:    counterUUID.String(),
		ObjectType:   "Counter",
		CounterName:  counterName,
		CounterGroup: groupId.String(),
		CounterVal:   0,
		StepVal:      1,
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

func counter_update(ops []*dynamodb.TransactWriteItem, table *string, groupId *UUID, counterId UUID, query string, stepval int) ([]*dynamodb.TransactWriteItem, error) {
	udr := dynamodb.Update{
		Key: map[string]*dynamodb.AttributeValue{
			counterIdCol:  {S: aws.String(counterId.String())},
			objectTypeCol: {S: aws.String("Counter")},
		},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":" + stepInit:    {N: aws.String(fmt.Sprintf("%d", stepval))},
			":" + counterInit: {N: aws.String("0")},
			":" + groupIdVal:  {S: aws.String(groupId.String())},
		},
		UpdateExpression:    aws.String(query),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_exists(%s) and %s = :%s", counterIdCol, counterGroupCol, groupIdVal)),
	}

	//log.Print("Update Query: ", query)
	ops = append(ops, &dynamodb.TransactWriteItem{
		Update: &udr,
	})

	return ops, nil
}

func counter_delete(ops []*dynamodb.TransactWriteItem, table *string, groupId *UUID, counterId UUID) ([]*dynamodb.TransactWriteItem, error) {
	dr := dynamodb.Delete{
		Key: map[string]*dynamodb.AttributeValue{
			counterIdCol:  {S: aws.String(counterId.String())},
			objectTypeCol: {S: aws.String("Counter")},
		},
		TableName: table,
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":" + groupIdVal: {S: aws.String(groupId.String())},
		},
		ConditionExpression: aws.String(fmt.Sprintf("%s = :%s", counterGroupCol, groupIdVal)),
	}

	ops = append(ops, &dynamodb.TransactWriteItem{
		Delete: &dr,
	})

	return ops, nil
}

const (
	dq_init    = iota
	dq_current = iota
	dq_inc     = iota
	dq_dec     = iota
)

func dnquery(stepmode int, countermode int) string {

	colexpr := func(mode int,
		colName string, defaultName string) string {
		switch mode {
		case dq_init:
			return fmt.Sprintf(":%s", defaultName)
		case dq_current:
			return fmt.Sprintf("if_not_exists(%s,:%s)", colName, defaultName)
		}
		return ""
	}

	xcolexpr := func(mode int,
		colName string, defaultName string,
		stepmode int,
		stepName string, stepDefault string) string {
		switch mode {
		case dq_inc:
			return fmt.Sprintf("%s + %s",
				colexpr(dq_current, colName, defaultName),
				colexpr(stepmode, stepName, stepDefault),
			)
		case dq_dec:
			return fmt.Sprintf("%s - %s",
				colexpr(dq_current, colName, defaultName),
				colexpr(stepmode, stepName, stepDefault),
			)
		}
		return colexpr(mode, colName, defaultName)
	}

	return fmt.Sprintf("SET %s=%s,%s=%s",
		stepCol,
		colexpr(stepmode, stepCol, stepInit),
		counterCol,
		xcolexpr(countermode, counterCol, counterInit, stepmode, stepCol, stepInit),
	)
}
