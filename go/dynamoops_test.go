package main

import (
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
	s, dbo, dbi := mockEnv(MakeUUID(), MakeUUID(), expEmail)

	group1 := MakeUUID()
	group2 := MakeUUID()

	udm, err := dynamodbattribute.Marshal(UserData{
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

	if err != nil {
		t.Errorf("Group list error %s", err)
	}

}
