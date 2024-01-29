package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

	"github.com/aws/aws-lambda-go/events"
)

const (
	stepCol        = "stepVal"
	counterCol     = "countVal"
	stepInit       = "stepinit"
	counterInit    = "countinit"
	counterNameCol = "counterName"
)

type Response = events.APIGatewayProxyResponse
type Request = events.APIGatewayProxyRequest

type CountData struct {
	CounterVal int `json:"countVal"`
	StepVal    int `json:"stepVal"`
}

type CountKey struct {
	name string `json:"counterName"`
}

type DBI interface {
	// all of the calls we actually make to dynamo need to be here.  Yuck!
	UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
}

func makeerror(err error) (Response, error) {
	return Response{StatusCode: 404}, err
}

func makeresponse(data any) (Response, error) {
	result, err := json.Marshal(data)

	if err != nil {
		return makeerror(err)
	}

	var buf bytes.Buffer

	json.HTMLEscape(&buf, result)

	var res = Response{
		StatusCode:      200,
		Body:            buf.String(),
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return res, nil
}

func dynamodb_iface() dynamodbiface.DynamoDBAPI {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	return dynamodbiface.DynamoDBAPI(svc)
}

func dynamocount_handler(dbi DBI, table string, counter string, create bool, query string, stepval string) (Response, error) {
	udr := dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			counterNameCol: {S: aws.String(counter)}},
		ReturnValues: aws.String("ALL_NEW"),
		TableName:    aws.String(table),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":" + stepInit:    {N: aws.String(stepval)},
			":" + counterInit: {N: aws.String("0")}},
		UpdateExpression: aws.String(query),
	}

	//log.Print("Update Query: ", query)

	if !create { // if we can't create a record, force 'name' to already be there
		udr.SetConditionExpression("attribute_exists(" + counterNameCol + ")")
	}

	udo, uderr := dbi.UpdateItem(&udr)

	if uderr != nil {
		return makeerror(uderr)
	}

	counts := CountData{}

	umerr := dynamodbattribute.UnmarshalMap(udo.Attributes, &counts)

	if umerr != nil {
		return makeerror(umerr)
	}

	return makeresponse(&counts)
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
