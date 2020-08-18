package main

const imports = `
import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
)
`

const predefined = `
const (
	ErrIDBaseError    = "base-error"
	ErrIDUnknownError = "unknown-error"
)

type Error interface {
	error
	Attributes() *ErrorAttributes
	Hint() string
	ChainErrMessage(msg string)
}

type ErrorAttributes struct {
	ID            string
	Stack         []byte
	Msg           string
	HTTPCode      int
	OriginalError error
	DefaultHint   string
}

type BaseError struct {
	ErrorAttributes
}

func (e *BaseError) Error() string {
	return e.Msg
}

func (e *BaseError) Attributes() *ErrorAttributes {
	return &e.ErrorAttributes
}

func (e *BaseError) ChainErrMessage(msg string) {
	e.Msg = fmt.Sprintf("%s: %s", msg, e.Msg)
}

func (e *BaseError) Hint() string {
	return e.DefaultHint
}

func ChainErrMessage(err error, msg string) error {
	knownErr, ok := err.(Error)
	if !ok {
		msg := fmt.Sprintf("%s: %v", msg, err)
		attrs := ErrorAttributes{
			ID:            ErrIDBaseError,
			Stack:         getStack(),
			Msg:           msg,
			HTTPCode:      http.StatusInternalServerError,
			OriginalError: err,
			DefaultHint:   http.StatusText(http.StatusInternalServerError),
		}
		baseErr := &BaseError{
			ErrorAttributes: attrs,
		}
		return baseErr
	}
	knownErr.ChainErrMessage(msg)
	return knownErr
}

type ErrorData struct {
	Attrs *ErrorAttributes
	Hint  string
}

func PrepareErrorToSend(err error) *ErrorData {
	knownErr, ok := err.(Error)
	if !ok {
		attrs := &ErrorAttributes{
			ID:       ErrIDUnknownError,
			HTTPCode: http.StatusInternalServerError,
			Msg:      err.Error(),
		}
		errData := &ErrorData{
			Attrs: attrs,
			Hint:  http.StatusText(attrs.HTTPCode),
		}
		return errData
	}
	errData := &ErrorData{
		Attrs: knownErr.Attributes(),
		Hint:  knownErr.Hint(),
	}
	return errData
}

func newErrorMessage(msg string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: %v", msg, err)
	}
	return msg
}

func getStack() []byte {
	stack := debug.Stack()
	stackFrames := strings.Split(string(stack), "\n")
	stackFrames = append([]string{stackFrames[0]}, stackFrames[7:]...)
	stack = []byte(strings.Join(stackFrames, "\n"))
	return stack
}
`
