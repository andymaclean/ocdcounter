package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type opResult struct {
	Success bool
	Result  string
	Id      string
}

type committer interface {
	commit(ops []*dynamodb.TransactWriteItem, id UUID) (Response, error)
	GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
}

type dynamoCommitter struct {
	dbi dynamodbiface.DynamoDBAPI
}

func dynamodb_iface() committer {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	//Create DynamoDB client
	svc := dynamodb.New(sess)

	return dynamoCommitter{dbi: dynamodbiface.DynamoDBAPI(svc)}
}

func (ci dynamoCommitter) commit(ops []*dynamodb.TransactWriteItem, id UUID) (Response, error) {
	input := dynamodb.TransactWriteItemsInput{
		TransactItems: ops,
	}

	_, err := ci.dbi.TransactWriteItems(&input)

	if err != nil {
		return makeerror(err)
	}

	return makeresponse(opResult{Success: true, Result: "OK", Id: id.String()})
}

func (ci dynamoCommitter) GetItem(in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return ci.dbi.GetItem(in)
}

func dynamodb_read_counter(ci committer, counterTable *string, groupId UUID, counterId UUID) (Response, error) {
	out, err := ci.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			counterIdCol: {S: aws.String(counterId.String())},
		},
		TableName: counterTable,
	})
	if err != nil {
		return makeerror(err)
	}

	var cd CountData

	cderr := dynamodbattribute.UnmarshalMap(out.Item, &cd)

	if cderr != nil {
		return makeerror(cderr)
	}

	if cd.CounterGroup != groupId.String() {
		return makeerror(fmt.Errorf("counter group is %s not %s", cd.CounterGroup, groupId.String()))
	}

	return makeresponse(cd)
}

func dynamodb_counter_list(ci committer, groupTableName *string, groupId UUID) (Response, error) {
	out, err := ci.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol: {S: aws.String(groupId.String())},
		},
		TableName: groupTableName,
	})

	if err != nil {
		return makeerror(err)
	}

	var gd GroupData

	gderr := dynamodbattribute.UnmarshalMap(out.Item, &gd)

	if gderr != nil {
		return makeerror(gderr)
	}

	return makeresponse(map[string][]string{"Counters": gd.Counters})
}

func dynamodb_counter_operate(ci committer, counterTableName *string, group UUID, id UUID, query string, stepVal int) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = counter_update(ops, counterTableName, group, id, query, stepVal)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, id)
}

func dynamodb_counter_create(ci committer, groupTableName *string, counterTableName *string, name string, group UUID) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = counter_create(ops, counterTableName, newid, name, group)

	if err != nil {
		return makeerror(err)
	}

	ops, err = group_update(ops, groupTableName, group, gquery(gr_add), newid)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, newid)
}

func dynamodb_counter_delete(ci committer, groupTableName *string, counterTableName *string, group UUID, counterId UUID) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = counter_delete(ops, counterTableName, group, counterId)

	if err != nil {
		return makeerror(err)
	}

	ops, err = group_update(ops, groupTableName, group, gquery(gr_remove), counterId)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, counterId)
}

func dynamodb_group_create(ci committer, groupTableName *string, name string) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = group_create(ops, groupTableName, newid, name)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, newid)
}
