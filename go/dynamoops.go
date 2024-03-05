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
)

type opResult struct {
	Success bool
	Result  string
	Id      string
}

type committer interface {
	commit(ops []*dynamodb.TransactWriteItem, id UUID) (Response, error)
	inline_commit(ops []*dynamodb.TransactWriteItem) error
	GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
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

func makeerror(err error) (Response, error) {
	return Response{
		StatusCode: 404,
		Body:       err.Error(),
	}, nil
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

func (ci dynamoCommitter) inline_commit(ops []*dynamodb.TransactWriteItem) error {
	input := dynamodb.TransactWriteItemsInput{
		TransactItems: ops,
	}

	_, err := ci.dbi.TransactWriteItems(&input)

	return err
}

func (ci dynamoCommitter) GetItem(in *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return ci.dbi.GetItem(in)
}

func (ci dynamoCommitter) Query(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return ci.dbi.Query(in)
}

func dynamodb_lookup_userUUID(ci committer, userTable *string, email *string) (UUID, error) {
	resp, err := ci.Query(&dynamodb.QueryInput{
		TableName: userTable,
		IndexName: aws.String(userEmailIndex),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":email": {S: email},
		},
		KeyConditionExpression: aws.String(emailCol + " = :email"),
		ProjectionExpression:   aws.String(userIdCol),
	})

	if err != nil {
		return NullUUID(), err
	}

	if resi := len(resp.Items); resi != 1 {
		return NullUUID(), fmt.Errorf("incorrect Item count (%d) from user lookup", resi)
	}

	return ToUUID(*resp.Items[0][userIdCol].S)
}

func dynamodb_read_counter(ci committer, counterTable *string, groupId UUID, counterId UUID) (Response, error) {
	out, err := ci.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			counterIdCol:  {S: aws.String(counterId.String())},
			objectTypeCol: {S: aws.String("Counter")},
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
			groupIdCol:    {S: aws.String(groupId.String())},
			objectTypeCol: {S: aws.String("Group")},
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

	ops, err = group_update(ops, groupTableName, group, gquery(gr_add_ctr), newid)

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

	ops, err = group_update(ops, groupTableName, group, gquery(gr_remove_ctr), counterId)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, counterId)
}

func dynamodb_group_create(ci committer, userTableName *string, groupTableName *string, creator *UUID, name string) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = group_create(ops, groupTableName, newid, name)

	if err != nil {
		return makeerror(err)
	}

	ops, err = user_update(ops, userTableName, creator, uquery(usr_add_grp), newid)

	if err != nil {
		return makeerror(err)
	}

	return ci.commit(ops, newid)
}

func dynamodb_user_create(ci committer, userTableName *string, userId UUID, name *string) error {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = user_create(ops, userTableName, userId, name)

	if err != nil {
		return err
	}

	return ci.inline_commit(ops)
}

func dynamodb_group_list(ci committer, userTableName *string, userId *UUID) (Response, error) {
	out, err := ci.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			userIdCol:     {S: aws.String(userId.String())},
			objectTypeCol: {S: aws.String("User")},
		},
		TableName: userTableName,
	})

	if err != nil {
		return makeerror(err)
	}

	var ud UserData

	gderr := dynamodbattribute.UnmarshalMap(out.Item, &ud)

	if gderr != nil {
		return makeerror(gderr)
	}

	return makeresponse(map[string][]string{"Groups": ud.Groups})
}
