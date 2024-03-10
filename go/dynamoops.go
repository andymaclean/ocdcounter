package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type opResult struct {
	Success bool
	Result  string
	Id      string
}

type DynamoOperator struct {
	groupTable      string `default:"group"`
	userTable       string `default:"user"`
	counterTable    string `default:"counter"`
	permissionTable string `default:"permission"`
	userEmailIndex  string `default:"userEmailIndex"`

	dbi DBInterface

	// put these here for ease of address-taking.
	counterType string `default:"counter"`
	userType    string `default:"user"`
	groupType   string `default:"group"`
}

func (dbo DynamoOperator) LookupUserUUID(email *string) (UUID, error) {
	resp, err := dbo.dbi.Query(&dynamodb.QueryInput{
		TableName: &dbo.userTable,
		IndexName: &dbo.userEmailIndex,
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

func (dbo DynamoOperator) CounterRead(s Session, counterId UUID) (Response, error) {
	out, err := dbo.dbi.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			counterIdCol:  {S: aws.String(counterId.String())},
			objectTypeCol: {S: &dbo.counterType},
		},
		TableName: &dbo.counterTable,
	})
	if err != nil {
		return makeerror(err)
	}

	var cd CountData

	cderr := dynamodbattribute.UnmarshalMap(out.Item, &cd)

	if cderr != nil {
		return makeerror(cderr)
	}

	if cd.CounterGroup != *s.GetGroupIdString() {
		return makeerror(fmt.Errorf("counter group is %s not %s", cd.CounterGroup, *s.GetGroupIdString()))
	}

	return makeresponse(cd)
}

func (dbo DynamoOperator) CounterList(s Session) (Response, error) {
	out, err := dbo.dbi.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			groupIdCol:    {S: s.GetGroupIdString()},
			objectTypeCol: {S: &dbo.groupType},
		},
		TableName: &dbo.groupTable,
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

func (dbo DynamoOperator) CounterUpdate(s Session, id UUID, query string, stepVal int) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_counter_update(ops, &dbo.counterTable, s.GetGroupId(), id, query, stepVal)

	if err != nil {
		return makeerror(err)
	}

	return dbo.dbi.commit(ops, id)
}

func (dbo DynamoOperator) CounterCreate(s Session, name string) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = append_counter_create(ops, &dbo.counterTable, newid, name, s.GetGroupId())

	if err != nil {
		return makeerror(err)
	}

	ops, err = append_group_update(ops, &dbo.groupTable, s.GetGroupId(), gquery(gr_add_ctr), newid)

	if err != nil {
		return makeerror(err)
	}

	return dbo.dbi.commit(ops, newid)
}

func (dbo DynamoOperator) CounterDelete(s Session, counterId UUID) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_counter_delete(ops, &dbo.counterTable, s.GetGroupId(), counterId)

	if err != nil {
		return makeerror(err)
	}

	ops, err = append_group_update(ops, &dbo.groupTable, s.GetGroupId(), gquery(gr_remove_ctr), counterId)

	if err != nil {
		return makeerror(err)
	}

	return dbo.dbi.commit(ops, counterId)
}

func (dbo DynamoOperator) GroupCreate(s Session, name string) (Response, error) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = append_group_create(ops, &dbo.groupTable, newid, name)

	if err != nil {
		return makeerror(err)
	}

	ops, err = append_user_update(ops, &dbo.userTable, s.GetUserId(), uquery(usr_add_grp), newid)

	if err != nil {
		return makeerror(err)
	}

	return dbo.dbi.commit(ops, newid)
}

func (dbo DynamoOperator) GroupList(s Session) (Response, error) {
	out, err := dbo.dbi.GetItem(&dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			userIdCol:     {S: s.GetUserIdString()},
			objectTypeCol: {S: &dbo.userType},
		},
		TableName: &dbo.userTable,
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

func (dbo DynamoOperator) UserCreate(newUserId UUID, name *string) error {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_user_create(ops, &dbo.userTable, newUserId, name)

	if err != nil {
		return err
	}

	return dbo.dbi.inline_commit(ops)
}
