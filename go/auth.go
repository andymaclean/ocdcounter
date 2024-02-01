package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func login(ctx context.Context, req Request) (Response, error) {
	mySession := session.Must(session.NewSession())
	svc := cognitoidentityprovider.New(mySession)
	email := aws.String(req.QueryStringParameters["email"])
	pwd := aws.String(req.QueryStringParameters["password"])
	pool := aws.String(os.Getenv("USER_POOL"))
	client := aws.String(os.Getenv("USER_POOL_CLIENT"))

	input := cognitoidentityprovider.AdminInitiateAuthInput{
		AuthFlow:   aws.String("ADMIN_NO_SRP_AUTH"),
		UserPoolId: pool,
		ClientId:   client,
		AuthParameters: map[string]*string{
			"USERNAME": email,
			"PASSWORD": pwd,
		},
	}

	resp, err := svc.AdminInitiateAuth(&input)

	if err != nil {
		return makeerror(err)
	}

	return makeresponse(map[string]string{"Result": "OK", "Token": *resp.AuthenticationResult.IdToken})
}

func signup(ctx context.Context, req Request) (Response, error) {
	mySession := session.Must(session.NewSession())
	svc := cognitoidentityprovider.New(mySession)

	email := aws.String(req.QueryStringParameters["email"])
	pwd := aws.String(req.QueryStringParameters["password"])
	pool := aws.String(os.Getenv("USER_POOL"))

	input := cognitoidentityprovider.AdminCreateUserInput{
		MessageAction: aws.String("SUPPRESS"),
		UserPoolId:    pool,
		Username:      email,
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: email,
			}, {
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
	}

	_, err := svc.AdminCreateUser(&input)

	if err != nil {
		return makeerror(err)
	}

	pwinput := cognitoidentityprovider.AdminSetUserPasswordInput{
		Password:   pwd,
		UserPoolId: pool,
		Username:   email,
		Permanent:  aws.Bool(true),
	}

	_, pwerr := svc.AdminSetUserPassword(&pwinput)

	if pwerr != nil {
		return makeerror(pwerr)
	}

	return makeresponse(map[string]string{"Result": "OK"})
}
