package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/aws/aws-lambda-go/events"
)

type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

type CountData struct {
	CounterVal int `json:"counterVal"`
	StepVal    int `json:"stepVal"`
}

type CountKey struct {
	name string
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

func dynamocount_handler(ctx context.Context, req Request, query string, stepval string) (Response, error) {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	udr := dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"name": {S: aws.String("defaultcount")}},
		ReturnValues: aws.String("ALL_NEW"),
		TableName:    aws.String(os.Getenv("COUNTER_TABLE")),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":stepinit":  {N: aws.String(stepval)},
			":countinit": {N: aws.String("0")}},
		UpdateExpression: aws.String(query),
	}

	udo, uderr := svc.UpdateItem(&udr)

	if uderr != nil {
		return makeerror(uderr)
	}

	counts := CountData{3, 4}

	umerr := dynamodbattribute.UnmarshalMap(udo.Attributes, &counts)

	if umerr != nil {
		return makeerror(umerr)
	}

	return makeresponse(&counts)
}

func dynamocount_increment(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(ctx, req, "SET stepVal=if_not_exists(stepVal,:stepinit), counterVal=if_not_exists(counterVal,:countinit) + if_not_exists(stepVal,:stepinit)", "1")
}

func dynamocount_decrement(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(ctx, req, "SET stepVal=if_not_exists(stepVal,:stepinit), counterVal=if_not_exists(counterVal,:countinit) - if_not_exists(stepVal,:stepinit)", "1")
}

func dynamocount_fetch(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(ctx, req, "SET stepVal=if_not_exists(stepVal,:stepinit), counterVal=if_not_exists(counterVal,:countinit)", "1")
}

func dynamocount_setstep(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(ctx, req, "SET stepVal=:stepinit, counterVal=if_not_exists(counterVal,:countinit)", req.PathParameters["stepVal"])
}

func dynamocount_reset(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(ctx, req, "SET stepVal=:stepinit, counterVal=:countinit", "1")
}
