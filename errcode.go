package main

import (
	"fmt"
)

const (
	ErrCodeOK                     = 1200
	ErrCodeBadRequest             = 1400
	ErrCodeOrgNameTooShort        = 14000
	ErrCodeNameTooShort           = 14001
	ErrCodeUserModifyNotAllowed   = 14002
	ErrCodeActionNotSupport       = 14003
	ErrCodeInvalidToken           = 14004
	ErrCodePasswordEmpty          = 14005
	ErrCodePasswordLengthMismatch = 14006
	ErrCodeEmailInvalid           = 14007
	ErrCodeNameInvalide           = 14008
	ErrCodeOrgNotEmpty            = 14009
	ErrCodeQuotaExceeded          = 140010
	ErrCodeLastAdminRestricted    = 140011
	ErrCodeUserInvited            = 140012
	ErrCodeUserExistsInOrg        = 140013
	ErrCodeUserNotRegistered      = 140014
	ErrCodeUnauthorized           = 1401
	ErrCodeForbidden              = 1403
	ErrCodePermissionDenied       = 14030
	ErrCodeNotFound               = 1404
	ErrCodeOrgNotFound            = 14040
	ErrCodeUserNotFound           = 14041
	ErrCodeMethodNotAllowed       = 1405
	ErrCodeOrgExist               = 14090
	ErrCodeUserExist              = 14091
	ErrCodeUserExistOnLdap        = 14092

	ErrCodeServiceUnavailable = 1503

	ErrCodeUnknownError = 1600
)

const (
	LDAPMagicNumber = 2000
	EtcdMagicNumber = 3000
)

var errText = map[int]string{
	ErrCodeOK:                     "OK",
	ErrCodeBadRequest:             "Bad request",
	ErrCodeOrgNameTooShort:        "Organization name too short",
	ErrCodeNameTooShort:           "Name too short",
	ErrCodeUserModifyNotAllowed:   "Can't modify username",
	ErrCodeActionNotSupport:       "Not supported action",
	ErrCodeInvalidToken:           "Invalid token",
	ErrCodePasswordEmpty:          "Password can't be empty",
	ErrCodePasswordLengthMismatch: "Password length must be 8 to 12 characters",
	ErrCodeEmailInvalid:           "Invalid email address",
	ErrCodeNameInvalide:           "Invalid username",
	ErrCodeOrgNotEmpty:            "Member(s) still in orgnazition",
	ErrCodeQuotaExceeded:          "Quota execeded",
	ErrCodeLastAdminRestricted:    "Last admin restricted",
	ErrCodeUserInvited:            "User already invited",
	ErrCodeUserExistsInOrg:        "User already in orgnazition",
	ErrCodeUserNotRegistered:      "User not registered",
	ErrCodeUnauthorized:           "Unauthorized",
	ErrCodeForbidden:              "Forbidden",
	ErrCodePermissionDenied:       "Permission denied",
	ErrCodeNotFound:               "Not found",
	ErrCodeOrgNotFound:            "Orgnazition not found",
	ErrCodeUserNotFound:           "User not found",
	ErrCodeMethodNotAllowed:       "Method not allowed",
	ErrCodeOrgExist:               "Organization already exists",
	ErrCodeUserExist:              "User exists",
	ErrCodeUserExistOnLdap:        "User exists on LDAP",
	ErrCodeServiceUnavailable:     "Service unavailable",

	ErrCodeUnknownError: "Unknown error",
}

func ErrText(code int) string {
	return errText[code]
}

type LdpError struct {
	Code    int
	Message string
}

func (e LdpError) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

func (e LdpError) New(code int) error {
	e.Code = code
	e.Message = ErrText(code)
	return e
}

func ldpErrorNew(code int) error {
	var e LdpError
	return e.New(code)
}
