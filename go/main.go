package main

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func unknownHandler() error {
	return errors.New("UNKNOWN HANDLER")
}

func loop(ctx context.Context, req Request) (Response, error) {
	return makeresponse(req)
}

func main() {
	var handler = os.Getenv("_HANDLER")
	switch handler {
	case "apiv1":
		api_v1() // This init routine has a lambda router and will call lambda.Start itself
	case "signup":
		lambda.Start(signup)
	case "login":
		lambda.Start(login)
	case "loop":
		lambda.Start(loop)
	default:
		lambda.Start(unknownHandler)
	}
}
