package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	t           *testing.T
	udf         updatefunc
	countername string
}

func (m mockDBI) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	cn := input.Key["name"].S
	if m.countername != *cn {
		return nil, errors.New(fmt.Sprintf("Incorrect counter name %s not %s", *cn, m.countername))
	}
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

func TestHandlerNoCreate(t *testing.T) {
	dbi := mockDBI{
		t:           t,
		countername: "test",
		udf: func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			if input.ConditionExpression == nil {
				t.Fatal("Nil conditional expression.")
			}

			return &dynamodb.UpdateItemOutput{}, errors.New("this failed")
		}}
	res, err := dynamocount_handler(dbi, "test", false, dnquery(dq_init, dq_init), "50")

	checkResponseCode(t, res, 404)

	if err == nil {
		t.Error("err is expected to be non nil.")
	}
}

func TestHandlerCreate(t *testing.T) {
	dbi := mockDBI{
		t:           t,
		countername: "test",
		udf: func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
			if input.ConditionExpression != nil {
				t.Fatal("Non nil conditional expression: ", *input.ConditionExpression)
			}

			cd := CountData{3, 4}
			av, _ := dynamodbattribute.MarshalMap(cd)

			return &dynamodb.UpdateItemOutput{
				Attributes: av,
			}, nil
		}}

	res, err := dynamocount_handler(dbi, "test", true, dnquery(dq_init, dq_init), "50")

	checkResponseCode(t, res, 200)

	if err != nil {
		t.Error("err is expected nil, is:", err)
	}

	checkResponseValues(t, res, 3, 4)
}
