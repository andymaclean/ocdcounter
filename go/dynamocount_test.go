package main

import (
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func TestDnquery_init(t *testing.T) {
	q := dnquery(dq_init, dq_init)
	expect := "SET stepVal=:stepinit,countVal=:countinit"
	if q != expect {
		t.Fatal("qnquery init failed", q, expect)
	}
}

func TestDnquery_current(t *testing.T) {
	q := dnquery(dq_current, dq_init)
	expect := "SET stepVal=if_not_exists(stepVal,:stepinit),countVal=:countinit"
	if q != expect {
		t.Fatal("qnquery init failed", q, expect)
	}
}

func TestDnquery_inc(t *testing.T) {
	q := dnquery(dq_init, dq_inc)
	expect := "SET stepVal=:stepinit,countVal=if_not_exists(countVal,:countinit) + :stepinit"
	if q != expect {
		t.Fatal("qnquery init failed", q, expect)
	}
}

func TestDnquery_dec(t *testing.T) {
	q := dnquery(dq_current, dq_dec)
	expect := "SET stepVal=if_not_exists(stepVal,:stepinit),countVal=if_not_exists(countVal,:countinit) - if_not_exists(stepVal,:stepinit)"
	if q != expect {
		t.Fatal("qnquery init failed", q, expect)
	}
}

//=====================  test with mock dynamodb ===============================

//func checkResponseCode(t *testing.T, res Response, expect int) {
//	if res.StatusCode != expect {
//		t.Errorf("Status code %d not %d", res.StatusCode, expect)
//	}
//}

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

const (
	expCounterTable = "CounterTable"
	expCounterName  = "TestCounter"
)

var expKey = MakeUUID().String()
var expGroup = MakeUUID()
var expCounterUUID = MakeUUID()

func checkNewCounter(t *testing.T, input *dynamodb.TransactWriteItem, newid UUID) {
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
		t.Errorf("Name is %s not %s", nval, expKey)
	}
	if gval != expGroup.String() {
		t.Errorf("Group is %s not %s", gval, expGroup.String())
	}
}

func TestCounterCreate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = counter_create(ops, expCounterTable, newid, expCounterName, expGroup)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkNewCounter(t, ops[0], newid)
}

func checkCounterUpdate(t *testing.T, input *dynamodb.TransactWriteItem, expStep int, expQuery string) {
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

	if grp, gerr := ToUUID(*ud.ExpressionAttributeValues[":"+groupId].S); gerr != nil {
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

func TestCounterUpdate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = counter_update(ops, expCounterTable, expGroup, expCounterUUID, "hello world", 12345)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkCounterUpdate(t, ops[0], 12345, "hello world")
}

func checkCounterDelete(t *testing.T, input *dynamodb.TransactWriteItem, counterId UUID) {
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

	if kval != counterId.String() {
		t.Errorf("Key is %s not %s", kval, counterId.String())
	}

	if grp, gerr := ToUUID(*dd.ExpressionAttributeValues[":"+groupId].S); gerr != nil {
		t.Errorf("Group is not a valid uuid: %s", gerr)
	} else {
		if grp != expGroup {
			t.Errorf("Initial counter is %s not %s", grp, expGroup.String())
		}
	}
}

func TestCounterDelete(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	dc := MakeUUID()

	ops, err = counter_delete(ops, expCounterTable, expGroup, dc)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkCounterDelete(t, ops[0], dc)
}

const (
	expGroupTable = "groupTable"
	expGroupName  = "groupName"
)

func checkNewGroup(t *testing.T, input *dynamodb.TransactWriteItem) {
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

func TestGroupCreate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = group_create(ops, expGroupTable, expGroup, expGroupName)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkNewGroup(t, ops[0])
}

func checkGroupUpdate(t *testing.T, input *dynamodb.TransactWriteItem, expVal1 UUID, expQuery string) {
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

	if grp, gerr := ToUUID(*ud.ExpressionAttributeValues[":val1"].SS[0]); gerr != nil {
		t.Errorf("Val1 is not a valid uuid: %s", gerr)
	} else {
		if grp != expVal1 {
			t.Errorf("Initial counter is %s not %s", grp, expVal1.String())
		}
	}

	if kval != expGroup.String() {
		t.Errorf("Group UUID is %s not %s.", kval, expGroup.String())
	}

	query := *ud.UpdateExpression

	if query != expQuery {
		t.Errorf("Query is %s not %s", query, expQuery)
	}
}

func TestGroupUpdate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	nuuid := MakeUUID()

	ops, err = group_update(ops, expGroupTable, expGroup, "hello world", nuuid)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkGroupUpdate(t, ops[0], nuuid, "hello world")
}
