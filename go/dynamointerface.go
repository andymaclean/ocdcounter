package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

// real dynamo DB interface.  All the code which actually talks to dynamo is in here.
// can be mocked for testing purposes and that's why it is separated like this.

func Create_DynamoDBInterface() dynamodbiface.DynamoDBAPI {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	return dynamodbiface.DynamoDBAPI(svc)
}
