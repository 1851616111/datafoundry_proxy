package messages

import (
	"fmt"
)

type Error struct {
	code    uint
	message string
}

var (
	Errors            [NumErrors]*Error
	ErrorNone         *Error
	ErrorUnkown       *Error
	ErrorJsonBuilding *Error
)

const (
	ErrorCodeAny = -1 // this is for testing only

	ErrorCodeNone = 0

	ErrorCodeUnkown                = 5900
	ErrorCodeJsonBuilding          = 5901
	ErrorCodeUrlNotSupported       = 5902
	ErrorCodeDbNotInitlized        = 5903
	ErrorCodeAuthFailed            = 5904
	ErrorCodePermissionDenied      = 5905
	ErrorCodeInvalidParameters     = 5906
	ErrorCodeGetDataItem           = 5907
	ErrorCodeCreateMessage         = 5908
	ErrorCodeModifyMessage         = 5909
	ErrorCodeGetMessage            = 5910
	ErrorCodeQueryMessage          = 5911
	ErrorCodeGetMessageStats       = 5912
	ErrorCodeResetMessageStats     = 5913
	ErrorCodeParseJsonFailed       = 5914
	ErrorCodeFailedToConnectRemote = 5915
	ErrorCodeNotOkRemoteResponse   = 5916
	ErrorCodeInvalidRemoteResponse = 5917

	NumErrors = 6999 // about 50k memroy wasted
)

func init() {
	initError(ErrorCodeNone, "OK")
	initError(ErrorCodeUnkown, "unknown error")
	initError(ErrorCodeJsonBuilding, "json building error")

	initError(ErrorCodeUrlNotSupported, "unsupported url")
	initError(ErrorCodeDbNotInitlized, "db is not inited")
	initError(ErrorCodeAuthFailed, "auth failed")
	initError(ErrorCodePermissionDenied, "permission denied")
	initError(ErrorCodeInvalidParameters, "invalid parameters")
	initError(ErrorCodeGetDataItem, "failed to get data item")
	initError(ErrorCodeCreateMessage, "failed to create message")
	initError(ErrorCodeModifyMessage, "failed to modify message")
	initError(ErrorCodeGetMessage, "failed to get message")
	initError(ErrorCodeQueryMessage, "failed to  query messages")
	initError(ErrorCodeGetMessageStats, "failed to get message statistics")
	initError(ErrorCodeResetMessageStats, "failed to reset message statistics")
	initError(ErrorCodeParseJsonFailed, "parse json failed")
	initError(ErrorCodeFailedToConnectRemote, "failed to connect remote")
	initError(ErrorCodeNotOkRemoteResponse, "remote response is not ok")
	initError(ErrorCodeInvalidRemoteResponse, "remote response error")

	ErrorNone = GetError(ErrorCodeNone)
	ErrorUnkown = GetError(ErrorCodeUnkown)
	ErrorJsonBuilding = GetError(ErrorCodeJsonBuilding)
}

func initError(code uint, message string) {
	if code < NumErrors {
		Errors[code] = newError(code, message)
	}
}

func GetError(code uint) *Error {
	if code > NumErrors {
		return Errors[ErrorCodeUnkown]
	}

	return Errors[code]
}

func GetError2(code uint, message string) *Error {
	e := GetError(code)
	if e == nil {
		return newError(code, message)
	} else {
		return newError(code, fmt.Sprintf("%s (%s)", e.message, message))
	}
}

func newError(code uint, message string) *Error {
	return &Error{code: code, message: message}
}

func newUnknownError(message string) *Error {
	return &Error{
		code:    ErrorCodeUnkown,
		message: message,
	}
}

func newInvalidParameterError(paramName string) *Error {
	return &Error{
		code:    ErrorCodeInvalidParameters,
		message: fmt.Sprintf("%s: %s", GetError(ErrorCodeInvalidParameters).message, paramName),
	}
}
