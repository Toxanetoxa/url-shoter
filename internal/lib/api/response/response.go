package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOk,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errorsMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errorsMessages = append(errorsMessages, fmt.Sprintf("поле %s поле не обнаружено", err.Field()))
		case "url":
			errorsMessages = append(errorsMessages, fmt.Sprintf("урл %s поле URL не валидно", err.Field()))
		default:
			errorsMessages = append(errorsMessages, fmt.Sprintf("поле: %s не валидное поле", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errorsMessages, ", "),
	}
}
