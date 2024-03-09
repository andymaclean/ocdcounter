package main

import (
	"bytes"
	"encoding/json"
)

func makeerror(err error) (Response, error) {
	return Response{
		StatusCode: 404,
		Body:       err.Error(),
	}, nil
}

func makeresponse(data any) (Response, error) {
	result, err := json.Marshal(data)

	if err != nil {
		return makeerror(err)
	}

	var buf bytes.Buffer

	json.HTMLEscape(&buf, result)

	var res = Response{
		StatusCode:      200,
		Body:            buf.String(),
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return res, nil
}
