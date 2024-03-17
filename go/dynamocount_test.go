package main

import (
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

var expCounterTable string = "CounterTable"
var expCounterName string = "TestCounter"

var expGroup = MakeUUID()
var expCounterUUID = MakeUUID()

func TestCounterCreate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	newid := MakeUUID()

	ops, err = append_counter_create(ops, &expCounterTable, newid, expCounterName, &expGroup)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkNewCounter(t, ops[0], newid, expCounterTable, expCounterName, expGroup)
}

func TestCounterUpdate(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	ops, err = append_counter_update(ops, &expCounterTable, &expGroup, expCounterUUID, "hello world", 12345)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkCounterUpdate(t, ops[0], 12345, "hello world", expCounterTable, expGroup, expCounterUUID)
}

func TestCounterDelete(t *testing.T) {
	var ops []*dynamodb.TransactWriteItem
	var err error

	dc := MakeUUID()

	ops, err = append_counter_delete(ops, &expCounterTable, &expGroup, dc)

	checkError(t, err, nil)

	checkOpsLen(t, ops, 1)

	checkCounterDelete(t, ops[0], dc, expCounterTable, expGroup)
}
