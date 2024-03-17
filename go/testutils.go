package main

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// checker routines to help with testing

func checkError(t *testing.T, err error, expect error) {
	if err != expect {
		t.Errorf("Error code %s not %s", err, expect)
	}
}

func checkOpsLen(t *testing.T, ops []*dynamodb.TransactWriteItem, explen int) {
	if len(ops) != explen {
		t.Errorf("Operation set length is %d not %d", len(ops), explen)
	}
}

func decodeResult(t *testing.T, res Response) opResult {
	var r opResult

	err := json.Unmarshal([]byte(res.Body), &r)

	if err != nil {
		t.Errorf("Cannot unmarshal result body, error %s", err)
	}

	return r
}

func decodeResultId(t *testing.T, res Response) UUID {
	us := decodeResult(t, res).Id

	u, err := ToUUID(us)

	if err != nil {
		t.Errorf("Invalid uuid %s: %s", us, err)
	}

	return u
}

func checkNewCounter(t *testing.T, input *dynamodb.TransactWriteItem, newid UUID,
	expCounterTable string,
	expCounterName string,
	expGroup UUID) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Update != nil {
		t.Error("Unexpected Update request")
	}

	if input.Put == nil {
		t.Fatal("Expected put request was not present")
	}

	put := input.Put

	if *put.TableName != expCounterTable {
		t.Errorf("Table name is %s not %s", *put.TableName, expCounterTable)
	}

	if kval, kerr := ToUUID(*put.Item[counterIdCol].S); kerr != nil {
		t.Error("Key is not a valid UUID")
	} else {
		if kval != newid {
			t.Errorf("Key is %s not %s", kval.String(), newid.String())
		}
	}

	if *put.Item[objectTypeCol].S != "Counter" {
		t.Error("Object type not correct in new counter")
	}

	nval := *put.Item[counterNameCol].S
	gval := *put.Item[counterGroupCol].S

	if cval, cverr := strconv.Atoi(*put.Item[counterCol].N); cverr != nil {
		t.Error("Count is not a valid integer")
	} else {
		if cval != 0 {
			t.Errorf("Initial counter is %d not 0", cval)
		}
	}

	if sval, sverr := strconv.Atoi(*put.Item[stepCol].N); sverr != nil {
		t.Error("Step is not a valid integer")
	} else {
		if sval != 1 {
			t.Errorf("Initial counter is %d not 1", sval)
		}
	}

	if nval != expCounterName {
		t.Errorf("Name is %s not %s", nval, expCounterName)
	}
	if gval != expGroup.String() {
		t.Errorf("Group is %s not %s", gval, expGroup.String())
	}
}

func checkCounterUpdate(t *testing.T, input *dynamodb.TransactWriteItem, expStep int, expQuery string,
	expCounterTable string,
	expGroup UUID,
	expCounterUUID UUID) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Put != nil {
		t.Error("Unexpected Put request")
	}

	if input.Update == nil {
		t.Fatal("Expected Update request was not present")
	}

	ud := input.Update

	if *ud.TableName != expCounterTable {
		t.Errorf("Table name is %s not %s", *ud.TableName, expCounterTable)
	}

	kval := *ud.Key[counterIdCol].S

	if *ud.Key[objectTypeCol].S != "Counter" {
		t.Error("Object type not correct in counter")
	}

	if step, sterr := strconv.Atoi(*ud.ExpressionAttributeValues[":"+stepInit].N); sterr != nil {
		t.Error("Step is not a valid integer")
	} else {
		if step != expStep {
			t.Errorf("Initial counter is %d not %d", step, expStep)
		}
	}
	if count, cerr := strconv.Atoi(*ud.ExpressionAttributeValues[":"+counterInit].N); cerr != nil {
		t.Error("Counter is not a valid integer")
	} else {
		if count != 0 {
			t.Errorf("Initial counter is %d not 0", count)
		}
	}

	if grp, gerr := ToUUID(*ud.ExpressionAttributeValues[":"+groupIdVal].S); gerr != nil {
		t.Errorf("Group is not a valid uuid: %s", gerr)
	} else {
		if grp != expGroup {
			t.Errorf("Initial counter is %s not %s", grp, expGroup.String())
		}
	}

	if kval != expCounterUUID.String() {
		t.Errorf("Counter UUID is %s not %s.", kval, expCounterUUID.String())
	}

	query := *ud.UpdateExpression

	if query != expQuery {
		t.Errorf("Query is %s not %s", query, expQuery)
	}
}

func checkCounterDelete(t *testing.T, input *dynamodb.TransactWriteItem, counterId UUID,
	expCounterTable string,
	expGroup UUID) {
	if input.Update != nil {
		t.Error("Unexpected Update request")
	}

	if input.Put != nil {
		t.Error("Unexpected Put request")
	}

	if input.Delete == nil {
		t.Fatal("Expected Delete request was not present")
	}

	dd := input.Delete

	if *dd.TableName != expCounterTable {
		t.Errorf("Table name is %s not %s", *dd.TableName, expCounterTable)
	}

	kval := *dd.Key[counterIdCol].S

	if *dd.Key[objectTypeCol].S != "Counter" {
		t.Error("Object type not correct in counter delete")
	}

	if kval != counterId.String() {
		t.Errorf("Key is %s not %s", kval, counterId.String())
	}

	if grp, gerr := ToUUID(*dd.ExpressionAttributeValues[":"+groupIdVal].S); gerr != nil {
		t.Errorf("Group is not a valid uuid: %s", gerr)
	} else {
		if grp != expGroup {
			t.Errorf("Initial counter is %s not %s", grp, expGroup.String())
		}
	}
}

func checkNewGroup(t *testing.T, input *dynamodb.TransactWriteItem,
	expGroupTable string,
	expGroup UUID,
	expGroupName string) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Update != nil {
		t.Error("Unexpected Update request")
	}

	if input.Put == nil {
		t.Fatal("Expected put request was not present")
	}

	put := input.Put

	if *put.TableName != expGroupTable {
		t.Errorf("Table name is %s not %s", *put.TableName, expGroupTable)
	}

	if kval, kerr := ToUUID(*put.Item[groupIdCol].S); kerr != nil {
		t.Error("Key is not a valid UUID")
	} else {
		if kval != expGroup {
			t.Errorf("Key is %s not %s", kval.String(), expGroup.String())
		}
	}

	if *put.Item[objectTypeCol].S != "Group" {
		t.Error("Object type not correct in new group")
	}

	nval := *put.Item[groupNameCol].S

	if _, present := put.Item[deleteMarkerCol]; present {
		t.Error("Delete marker in record")
	}

	if _, present := put.Item[counterListCol]; present {
		t.Error("Empty counter list in record")
	}

	if nval != expGroupName {
		t.Errorf("Name is %s not %s", nval, expGroupName)
	}
}

func checkGroupUpdate(t *testing.T, input *dynamodb.TransactWriteItem, expVal1 UUID, expQuery string,
	expGroupTable string,
	expGroup UUID,
	expUser UUID) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Put != nil {
		t.Error("Unexpected Put request")
	}

	if input.Update == nil {
		t.Fatal("Expected Update request was not present")
	}

	ud := input.Update

	if *ud.TableName != expGroupTable {
		t.Errorf("Table name is %s not %s", *ud.TableName, expGroupTable)
	}

	kval := *ud.Key[groupIdCol].S

	if *ud.Key[objectTypeCol].S != "Group" {
		t.Error("Object type not correct in group update")
	}

	if grp, gerr := ToUUID(*ud.ExpressionAttributeValues[":val1"].SS[0]); gerr != nil {
		t.Errorf("Val1 is not a valid uuid: %s", gerr)
	} else {
		if grp != expVal1 {
			t.Errorf("Initial counter is %s not %s", grp, expVal1.String())
		}
	}

	if kval != expGroup.String() {
		t.Errorf("Group UUID is %s not %s.", kval, expUser.String())
	}

	query := *ud.UpdateExpression

	if query != expQuery {
		t.Errorf("Query is %s not %s", query, expQuery)
	}
}

func checkNewUser(t *testing.T, input *dynamodb.TransactWriteItem,
	expUserTable string,
	expUser UUID,
	expUserName string) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Update != nil {
		t.Error("Unexpected Update request")
	}

	if input.Put == nil {
		t.Fatal("Expected put request was not present")
	}

	put := input.Put

	if *put.TableName != expUserTable {
		t.Errorf("Table name is %s not %s", *put.TableName, expUserTable)
	}

	kval := *put.Item[userIdCol].S
	if kval != expUser.String() {
		t.Errorf("Key is %s not %s", kval, expUser.String())
	}

	if *put.Item[objectTypeCol].S != "User" {
		t.Error("Object type not correct in new user")
	}

	nval := *put.Item[userNameCol].S

	if _, present := put.Item[deleteMarkerCol]; present {
		t.Error("Delete marker in record")
	}

	if _, present := put.Item[groupListCol]; present {
		t.Error("Empty counter list in record")
	}

	if nval != expUserName {
		t.Errorf("Name is %s not %s", nval, expUserName)
	}
}

func checkUserUpdate(t *testing.T, input *dynamodb.TransactWriteItem, expVal1 UUID, expQuery string,
	expUserTable string,
	expUser UUID) {
	if input.Delete != nil {
		t.Error("Unexpected delete request")
	}

	if input.Put != nil {
		t.Error("Unexpected Put request")
	}

	if input.Update == nil {
		t.Fatal("Expected Update request was not present")
	}

	ud := input.Update

	if *ud.TableName != expUserTable {
		t.Errorf("Table name is %s not %s", *ud.TableName, expUserTable)
	}

	if *ud.Key[objectTypeCol].S != "User" {
		t.Error("Object type not correct in user update")
	}

	kval := *ud.Key[userIdCol].S

	if grp, gerr := ToUUID(*ud.ExpressionAttributeValues[":val1"].SS[0]); gerr != nil {
		t.Errorf("Val1 is not a valid uuid: %s", gerr)
	} else {
		if grp != expVal1 {
			t.Errorf("Initial counter is %s not %s", grp, expVal1.String())
		}
	}

	if kval != expUser.String() {
		t.Errorf("Group UUID is %s not %s.", kval, expUser.String())
	}

	query := *ud.UpdateExpression

	if query != expQuery {
		t.Errorf("Query is %s not %s", query, expQuery)
	}
}
