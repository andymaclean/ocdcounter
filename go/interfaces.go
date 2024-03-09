package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// generic session specific routines for session parameters to operations
// these are for task level operations and represent the session-specific parameters for the operation.
// i.e. what group are they operating in, what user are they, etc.
type Session interface {
	GetUserId() UUID
	GetUserIdString() *string

	GetGroupId() UUID
	GetGroupIdString() *string
}

// Data operator is the high level interface which lambda calls
// this provides a task-based API and is mainly used for collecting together
// global parameters like dynamo table names.
// It could be used for mocking but mostly I used it as a handy collection of function definitions
type DataOperator interface {
	// User lookup by e-mail
	LookupUserUUID(email *string) (UUID, error)

	// add a new record for a user
	UserCreate(userId UUID, name *string) error

	// CRUD functions for counters
	CounterCreate(s Session, counterName string) (Response, error)
	CounterRead(s Session, counterId UUID) (Response, error)
	CounterUpdate(s Session, id UUID, query string, stepVal int) (Response, error)
	CounterList(s Session) (Response, error)
	CounterDelete(s Session, counterId UUID) (Response, error)

	// CRUD functions for groups
	GroupCreate(s Session, name string) (Response, error)
	GroupList(s Session) (Response, error)
}

// DBInterface is the low level interface which actually talks to DynamoDB.
// Separated so I can mock out things for testing.
type DBInterface interface {
	commit(ops []*dynamodb.TransactWriteItem, id UUID) (Response, error)
	inline_commit(ops []*dynamodb.TransactWriteItem) error
	GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}
