package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

type updatefunc func(*dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)

type mockDBI struct {
	t   *testing.T
	udf updatefunc
}

func (m mockDBI) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return m.udf(input)
}

func checkResponseCode(t *testing.T, res Response, expect int) {
	if res.StatusCode != expect {
		t.Errorf("Status code %d not %d", res.StatusCode, expect)
	}
}

func checkResponseValues(t *testing.T, res Response, expectCount int, expectStep int) {
	var resc CountData

	err := json.Unmarshal([]byte(res.Body), &resc)

	if err != nil {
		t.Fatal("Error unmarshaling body: ", err)
	}

	if resc.CounterVal != expectCount {
		t.Errorf("Expected count %d, got %d", expectCount, resc.CounterVal)
	}

	if resc.StepVal != expectStep {
		t.Errorf("Expected step %d, got %d", expectStep, resc.StepVal)
	}
}

const (
	expTable = "CounterTable"
	expKey   = "testCounter"
)

func check_ddbi(t *testing.T, input *dynamodb.UpdateItemInput) {
	if *input.TableName != expTable {
		t.Errorf("Table name is %s not %s", *input.TableName, expTable)
	}

	kval := *input.Key[counterNameCol].S

	if kval != expKey {
		t.Errorf("Key name is %s not %s", kval, expKey)
	}
}

func TestHandlerNoCreate(t *testing.T) {
	dbi := mockDBI{
		t: t,
		udf: func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			check_ddbi(t, input)

			if input.ConditionExpression == nil {
				t.Fatal("Nil conditional expression.")
			}

			return &dynamodb.UpdateItemOutput{}, errors.New("this failed")
		}}
	res, err := dynamocount_handler(dbi, expTable, expKey, false, dnquery(dq_init, dq_init), "50")

	checkResponseCode(t, res, 404)

	if err == nil {
		t.Error("err is expected to be non nil.")
	}
}

func TestHandlerCreate(t *testing.T) {
	dbi := mockDBI{
		t: t,
		udf: func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			check_ddbi(t, input)

			if input.ConditionExpression != nil {
				t.Fatal("Non nil conditional expression: ", *input.ConditionExpression)
			}

			cd := CountData{3, 4}
			av, _ := dynamodbattribute.MarshalMap(cd)

			return &dynamodb.UpdateItemOutput{
				Attributes: av,
			}, nil
		}}

	res, err := dynamocount_handler(dbi, expTable, expKey, true, dnquery(dq_init, dq_init), "50")

	checkResponseCode(t, res, 200)

	if err != nil {
		t.Error("err is expected nil, is:", err)
	}

	checkResponseValues(t, res, 3, 4)
}
