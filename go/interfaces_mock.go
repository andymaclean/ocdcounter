package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type MockDataOperator struct {
	userId UUID

	newId UUID

	retErr error

	expEmail string

	funcName []string
}

func (mo *MockDataOperator) LookupUserUUID(email *string) (UUID, error) {
	if mo.expEmail == "" || mo.expEmail == *email {
		mo.funcName = append(mo.funcName, "LookupUserUUID")
		return mo.userId, mo.retErr
	} else {
		return NullUUID(), fmt.Errorf("wrong email:  %s not %s", *email, mo.expEmail)
	}
}

// add a new record for a user
func (mo *MockDataOperator) UserCreate(userId UUID, name *string) error {
	mo.funcName = append(mo.funcName, "UserCreate")
	return mo.retErr
}

// CRUD functions for counters
func (mo *MockDataOperator) CounterCreate(s Session, counterName string) (Response, error) {
	mo.funcName = append(mo.funcName, "CounterCreate")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{Success: true, Result: "OK", Id: mo.newId.String()})
}
func (mo *MockDataOperator) CounterRead(s Session, counterId UUID) (Response, error) {
	mo.funcName = append(mo.funcName, "CounterRead")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{Success: true, Result: "OK", Id: counterId.String()})
}
func (mo *MockDataOperator) CounterUpdate(s Session, id UUID, query string, stepVal int) (Response, error) {
	mo.funcName = append(mo.funcName, "CounterUpdate")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{Success: true, Result: "OK", Id: id.String()})
}
func (mo *MockDataOperator) CounterList(s Session) (Response, error) {
	mo.funcName = append(mo.funcName, "CounterList")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{
		Result:  "OK",
		Success: true,
		Id:      "?",
		Items:   []string{mo.newId.String()},
	})
}
func (mo *MockDataOperator) CounterDelete(s Session, counterId UUID) (Response, error) {
	mo.funcName = append(mo.funcName, "CounterDelete")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{Success: true, Result: "OK", Id: counterId.String()})
}

// CRUD functions for groups
func (mo *MockDataOperator) GroupCreate(s Session, name string) (Response, error) {
	mo.funcName = append(mo.funcName, "GroupCreate")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{Success: true, Result: "OK", Id: mo.newId.String()})
}
func (mo *MockDataOperator) GroupList(s Session) (Response, error) {
	mo.funcName = append(mo.funcName, "GroupList")
	if mo.retErr != nil {
		return makeerror(mo.retErr)
	}
	return makeresponse(opResult{
		Result:  "OK",
		Success: true,
		Id:      "?",
		Items:   []string{mo.newId.String()},
	})
}

type MockDBInterface struct {
	twi dynamodb.TransactWriteItemsInput
	gii dynamodb.GetItemInput
	qi  dynamodb.QueryInput

	gio dynamodb.GetItemOutput

	retErr error
}

func (mo *MockDBInterface) TransactWriteItems(input *dynamodb.TransactWriteItemsInput) (*dynamodb.TransactWriteItemsOutput, error) {
	mo.twi = *input
	return nil, mo.retErr
}

func (mo *MockDBInterface) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	mo.gii = *input
	return &mo.gio, mo.retErr
}

func (mo *MockDBInterface) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	mo.qi = *input
	return nil, mo.retErr
}
