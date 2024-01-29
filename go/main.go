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
	case "apiv1":
		api_v1() // This init routine has a lambda router and will call lambda.Start itself
	default:
		lambda.Start(unknownHandler)
	}
}
