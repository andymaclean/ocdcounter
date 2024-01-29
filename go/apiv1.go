package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aquasecurity/lmdrouter"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var tableName string
var dbi = dynamodb_iface()

func incCounter(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], false, dnquery(dq_current, dq_inc), "1")
}

func decCounter(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], false, dnquery(dq_current, dq_dec), "1")
}

func getCounter(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], false, dnquery(dq_current, dq_current), "1")
}

func setCounterStep(ctx context.Context, req Request) (Response, error) {
	sv := req.QueryStringParameters["stepVal"]
	log.Print("stepVal is ", sv)
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], false, dnquery(dq_init, dq_current), sv)
}

func resetCounter(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], false, dnquery(dq_current, dq_init), "1")
}

func createCounter(ctx context.Context, req Request) (Response, error) {
	return dynamocount_handler(dbi, tableName, req.PathParameters["id"], true, dnquery(dq_current, dq_current), "1")
}

type counterEntry struct {
	Name string `json:"counterName"`
}

func deleteCounter(ctx context.Context, req Request) (Response, error) {
	id := req.PathParameters["id"]

	input := dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			counterNameCol: {S: aws.String(id)}},
		TableName: aws.String(tableName),
	}

	dbo, err := dbi.DeleteItem(&input)

	if err != nil {
		return makeerror(err)
	}

	return makeresponse(dbo)
}

func listCounters(ctx context.Context, req Request) (Response, error) {
	rows, err := dbi.Scan(&dynamodb.ScanInput{
		ConsistentRead:       aws.Bool(true),
		ProjectionExpression: aws.String(counterNameCol),
		TableName:            aws.String(os.Getenv("COUNTER_TABLE")),
	})

	if err != nil {
		return makeerror(err)
	}

	res := []counterEntry{}

	umerr := dynamodbattribute.UnmarshalListOfMaps(rows.Items, &res)

	if umerr != nil {
		return makeerror(umerr)
	}

	return makeresponse(res)
}

func api_v1() {

	// Set up the lambda router to handle the API endpoints
	router := lmdrouter.NewRouter("/api/v1/counter")
	router.Route("GET", "/", listCounters)
	router.Route("GET", "/:id", getCounter)
	router.Route("POST", "/:id", createCounter)
	router.Route("POST", "/:id/increment", incCounter)
	router.Route("POST", "/:id/decrement", decCounter)
	router.Route("POST", "/:id/reset", resetCounter)
	router.Route("POST", "/:id/step", setCounterStep)
	router.Route("DELETE", "/:id", deleteCounter)

	tableName = os.Getenv("COUNTER_TABLE")

	lambda.Start(func(ctx context.Context, req Request) (Response, error) {
		//req.Path = "/" + req.PathParameters["Path"]
		log.Print("Router called with req ", req)
		return router.Handler(ctx, req)
	})
}
