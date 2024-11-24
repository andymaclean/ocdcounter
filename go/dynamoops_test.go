package main

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func mockEnv(expUid UUID, expGid UUID, expEmail string) (Session, DynamoOperator, *MockDBInterface) {
	s := APISession{
		userId:    expUid,
		groupId:   expGid,
		userEmail: &expEmail,
	}

	dbi := MockDBInterface{}

	dbo := DynamoOperator{
		counterTable:    "CounterTable",
		groupTable:      "GroupTable",
		userTable:       "UserTable",
		permissionTable: "PermissionTable",
		userEmailIndex:  "UserEmailIndex",

		dbi: &dbi,

		counterType: "Counter",
		userType:    "User",
		groupType:   "Group",
	}
	return &s, dbo, &dbi
}

func TestDBOUserCreate(t *testing.T) {
	var expEmail = "foo@bar.com"

	_, dbo, dbi := mockEnv(MakeUUID(), MakeUUID(), expEmail)

	newid := MakeUUID()

	err := dbo.UserCreate(newid, &expEmail)

	checkError(t, err, nil)

	checkOpsLen(t, dbi.twi.TransactItems, 1)

	checkNewUser(t, dbi.twi.TransactItems[0], dbo.userTable, newid, expEmail)
}

func TestDBOGroupCreate(t *testing.T) {
	var groupName = "AGroup"
	var expEmail = "foo@bar.com"

	s, dbo, dbi := mockEnv(MakeUUID(), MakeUUID(), expEmail)

	resp, err := dbo.GroupCreate(s, groupName)

	checkError(t, err, nil)

	checkOpsLen(t, dbi.twi.TransactItems, 2)

	newid := decodeResultId(t, resp)

	checkNewGroup(t, dbi.twi.TransactItems[0], dbo.groupTable, newid, groupName)
	checkUserUpdate(t, dbi.twi.TransactItems[1], newid, uquery(usr_add_grp), dbo.userTable, *s.GetUserId())
}

func TestDBOGroupList(t *testing.T) {
	var expEmail = "foo@bar.com"
	s, dbo, dbi := mockEnv(MakeUUID(), MakeUUID(), expEmail)

	group1 := MakeUUID()
	group2 := MakeUUID()

	udm, err := dynamodbattribute.MarshalMap(UserData{
		UserId:   s.GetUserId().String(),
		UserName: "MrSmith",
		Groups:   []string{group1.String(), group2.String()},
	})

	if err != nil {
		panic("oops")
	}

	dbi.gio = dynamodb.GetItemOutput{
		Item: udm,
	}

	resp, err := dbo.GroupList(s)

	checkError(t, err, nil)

	var r opResult

	umerr := json.Unmarshal([]byte(resp.Body), &r)

	checkError(t, umerr, nil)

	if r.Id != s.GetUserId().String() {
		t.Errorf("Expected id %s, got %s", s.GetUserId().String(), r.Id)
	}

	if r.Result != "OK" || r.Success != true {
		t.Errorf("Unexpected list result %s, success %t", r.Result, r.Success)
	}

	if len(r.Items) != 2 || r.Items[0] != group1.String() || r.Items[1] != group2.String() {
		t.Errorf("Group list is incorrect:  %s", r.Items)
	}
}

func TestDBOCounterList(t *testing.T) {
	var expEmail = "foo@bar.com"
	s, dbo, dbi := mockEnv(MakeUUID(), MakeUUID(), expEmail)

	group := MakeUUID()

	counter1 := MakeUUID()
	counter2 := MakeUUID()

	udm, err := dynamodbattribute.MarshalMap(GroupData{
		GroupId:    group.String(),
		GroupName:  "MrSmithGroup",
		ObjectType: "Group",
		Counters:   []string{counter1.String(), counter2.String()},
	})

	if err != nil {
		panic("oops")
	}

	dbi.gio = dynamodb.GetItemOutput{
		Item: udm,
	}

	resp, err := dbo.CounterList(s)

	checkError(t, err, nil)

	var r opResult

	umerr := json.Unmarshal([]byte(resp.Body), &r)

	checkError(t, umerr, nil)

	if r.Id != group.String() {
		t.Errorf("Expected id %s, got %s", group.String(), r.Id)
	}

	if r.Result != "OK" || r.Success != true {
		t.Errorf("Unexpected list result %s, success %t", r.Result, r.Success)
	}

	if len(r.Items) != 2 || r.Items[0] != counter1.String() || r.Items[1] != counter2.String() {
		t.Errorf("Counter list is incorrect:  %s", r.Items)
	}
}
