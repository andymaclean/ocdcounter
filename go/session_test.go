package main

import (
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func checkId(t *testing.T, expId UUID, resId UUID, resStr *string, resStr2 *string, tn string) {
	if resId != expId {
		t.Errorf("Expecting %s ID %s not %s", tn, expId.String(), resId.String())
	}

	if *resStr != expId.String() {
		t.Errorf("Expecting %s ID String %s not %s", tn, expId.String(), *resStr)
	}

	if resStr != resStr2 {
		t.Errorf("%s ID string pinters differ.  %p not %p ", tn, resStr, resStr2)
	}
}

func checkAPISession(t *testing.T, s *APISession, expUid UUID, expGid UUID) {
	checkId(t, expUid, *s.GetUserId(), s.GetUserIdString(), s.GetUserIdString(), "User")

	checkId(t, expGid, *s.GetGroupId(), s.GetGroupIdString(), s.GetGroupIdString(), "Group")
}

func TestUIDAPISession(t *testing.T) {
	expUid := MakeUUID()
	expGid := NullUUID()
	expEmail := "foo@bar.com"

	s := APISession{
		userId:    expUid,
		userEmail: &expEmail,
	}

	checkAPISession(t, &s, expUid, expGid)
}

func TestUIDGIDAPISession(t *testing.T) {
	expUid := MakeUUID()
	expGid := MakeUUID()
	expEmail := "foo@bar.com"

	s := APISession{
		userId:    expUid,
		groupId:   expGid,
		userEmail: &expEmail,
	}

	checkAPISession(t, &s, expUid, expGid)
}

func TestCreateAPISessionFail1(t *testing.T) {
	expUid := MakeUUID()

	dbo := MockDataOperator{
		userId: expUid,
	}

	req := Request{}

	_, err := Create_APISession(&dbo, req)

	if err == nil {
		t.Fatalf("Expecting a fail.")
	}

	if err.Error() != "username is not in JWT claims" {
		t.Errorf("Wrong error text %s", err.Error())
	}
}

func TestCreateAPISessionFail1b(t *testing.T) {
	expUid := MakeUUID()

	dbo := MockDataOperator{
		userId: expUid,
	}

	req := Request{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{},
				},
			},
		},
	}

	_, err := Create_APISession(&dbo, req)

	if err == nil {
		t.Fatalf("Expecting a fail.")
	}

	if err.Error() != "username is not in JWT claims" {
		t.Errorf("Wrong error text %s", err.Error())
	}
}

func TestCreateAPISessionFail2(t *testing.T) {
	expUid := MakeUUID()
	expEmail := "foo@bar.com"

	dbo := MockDataOperator{
		userId: expUid,
		retErr: fmt.Errorf("NOPE"),
	}

	req := Request{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{
						"cognito:username": expEmail,
					},
				},
			},
		},
	}

	_, err := Create_APISession(&dbo, req)

	if err == nil {
		t.Fatalf("Expecting a fail.")
	}

	if err.Error() != "NOPE" {
		t.Errorf("Wrong error text %s", err.Error())
	}
}

func TestCreateAPISessionUIDOnly(t *testing.T) {
	expUid := MakeUUID()
	expGid := NullUUID()
	expEmail := "foo@bar.com"

	dbo := MockDataOperator{
		userId: expUid,
	}

	req := Request{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{
						"cognito:username": expEmail,
					},
				},
			},
		},
	}

	s, err := Create_APISession(&dbo, req)

	if err != nil {
		t.Fatalf("Creation fail: %s", err.Error())
	}

	checkAPISession(t, &s, expUid, expGid)
}

func TestCreateAPISessionUIDGIDFail(t *testing.T) {
	expUid := MakeUUID()
	expEmail := "foo@bar.com"

	dbo := MockDataOperator{
		userId: expUid,
	}

	req := Request{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{
						"cognito:username": expEmail,
					},
				},
			},
		},
		PathParameters: map[string]string{
			"group": "FooBarBat",
		},
	}

	_, err := Create_APISession(&dbo, req)

	if err == nil {
		t.Fatalf("Expecting a fail.")
	}
}

func TestCreateAPISessionUIDGID(t *testing.T) {
	expUid := MakeUUID()
	expGid := MakeUUID()
	expEmail := "foo@bar.com"

	dbo := MockDataOperator{
		userId: expUid,
	}

	req := Request{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{
						"cognito:username": expEmail,
					},
				},
			},
		},
		PathParameters: map[string]string{
			"group": expGid.String(),
		},
	}

	s, err := Create_APISession(&dbo, req)

	if err != nil {
		t.Fatalf("Creation fail: %s", err.Error())
	}

	checkAPISession(t, &s, expUid, expGid)
}
