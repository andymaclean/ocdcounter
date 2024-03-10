package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// real dynamo DB interface.  All the code which actually talks to dynamo is in here.
// can be mocked for testing purposes and that's why it is separated like this.

type dynamoDBInterface struct {
	dbi dynamodbiface.DynamoDBAPI
}

func dynamodb_iface() dynamoDBInterface {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	return dynamoDBInterface{dbi: dynamodbiface.DynamoDBAPI(svc)}
}

func Create_DynamoDBInterface() dynamoDBInterface {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	return dynamoDBInterface{dbi: dynamodbiface.DynamoDBAPI(svc)}
}

func (ci dynamoDBInterface) commit(ops []*dynamodb.TransactWriteItem, id UUID) (Response, error) {
	input := dynamodb.TransactWriteItemsInput{
		TransactItems: ops,
	}

	_, err := ci.dbi.TransactWriteItems(&input)

	if err != nil {
		return makeerror(err)
	}

	return makeresponse(opResult{Success: true, Result: "OK", Id: id.String()})
}

func (ci dynamoDBInterface) inline_commit(ops []*dynamodb.TransactWriteItem) error {
	input := dynamodb.TransactWriteItemsInput{
		TransactItems: ops,
	}

	_, err := ci.dbi.TransactWriteItems(&input)

	return err
}

func (ci dynamoDBInterface) GetItem(in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return ci.dbi.GetItem(in)
}

func (ci dynamoDBInterface) Query(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return ci.dbi.Query(in)
}
