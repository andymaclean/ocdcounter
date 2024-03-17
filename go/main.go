package main

import (
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func unknownHandler() error {
	return errors.New("UNKNOWN HANDLER")
}

func main() {
	api := APIHandler{
		dbo: DynamoOperator{
			counterTable:    os.Getenv("COUNTER_TABLE"),
			groupTable:      os.Getenv("GROUP_TABLE"),
			userTable:       os.Getenv("USER_TABLE"),
			permissionTable: os.Getenv("PERMISSION_TABLE"),
			userEmailIndex:  os.Getenv("USER_EMAIL_LOOKUP"),

			dbi: Create_DynamoDBInterface(),

			counterType: "Counter",
			userType:    "User",
			groupType:   "Group",
		},
	}

	switch os.Getenv("_HANDLER") {
	case "apipublic":
		lambda.Start(api.public_handler_gatewayv2)
	case "apiprivate":
		lambda.Start(api.private_handler_gatewayv2)
	default:
		lambda.Start(unknownHandler)
	}
}
