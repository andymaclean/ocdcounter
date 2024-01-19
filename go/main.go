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
	var handler = os.Getenv("_HANDLER")
	switch handler {
	case "increment":
		lambda.Start(dynamocount_increment)
	case "decrement":
		lambda.Start(dynamocount_decrement)
	case "fetch":
		lambda.Start(dynamocount_fetch)
	case "setstep":
		lambda.Start(dynamocount_setstep)
	case "reset":
		lambda.Start(dynamocount_reset)
	default:
		lambda.Start(unknownHandler)
	}

}
