package messages

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	_ "github.com/go-sql-driver/mysql"

	"github.com/asiainfoLDP/datahub_commons/common"
)

//======================================================
// 
//======================================================

const (
	StringParamType_General        = 0
	StringParamType_UrlWord        = 1
	StringParamType_UnicodeUrlWord = 2
	StringParamType_Email          = 3
)

//======================================================
//
//======================================================

func MustBoolParam(params httprouter.Params, paramName string) (bool, *Error) {
	bool_str := params.ByName(paramName)
	if bool_str == "" {
		return false, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return false, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, bool_str))
	}

	return b, nil
}

func MustBoolParamInMap(m map[string]interface{}, paramName string) (bool, *Error) {
	v, ok := m [paramName]
	if ok {
		b, ok := v.(bool)
		if ok {
			return b, nil
		}
		
		return false, newInvalidParameterError(fmt.Sprintf("param %s is not bool", paramName))
	}
	
	return false, newInvalidParameterError(fmt.Sprintf("param %s is not found", paramName))
}

func MustBoolParamInQuery(r *http.Request, paramName string) (bool, *Error) {
	bool_str := r.Form.Get(paramName)
	if bool_str == "" {
		return false, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return false, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, bool_str))
	}

	return b, nil
}

func optionalBoolParamInQuery(r *http.Request, paramName string, defaultValue bool) bool {
	bool_str := r.Form.Get(paramName)
	if bool_str == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return defaultValue
	}

	return b
}

func _optionalIntParam(intStr string, defaultInt int64) int64 {
	if intStr == "" {
		return defaultInt
	}

	i, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return defaultInt
	} else {
		return i
	}
}

func optionalIntParamInQuery(r *http.Request, paramName string, defaultInt int64) int64 {
	return _optionalIntParam(r.Form.Get(paramName), defaultInt)
}

func _mustIntParam(paramName string, int_str string) (int64, *Error) {
	if int_str == "" {
		return 0, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	i, err := strconv.ParseInt(int_str, 10, 64)
	if err != nil {
		return 0, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, int_str))
	}

	return i, nil
}

func MustIntParamInQuery(r *http.Request, paramName string) (int64, *Error) {
	return _mustIntParam(paramName, r.Form.Get(paramName))
}

func MustIntParamInPath(params httprouter.Params, paramName string) (int64, *Error) {
	return _mustIntParam(paramName, params.ByName(paramName))
}

func MustIntParamInMap(m map[string]interface{}, paramName string) (int64, *Error) {
	v, ok := m[paramName]
	if ok {
		i, ok := v.(float64)
		if ok {
			return int64(i), nil
		}

		return 0, newInvalidParameterError(fmt.Sprintf("param %s is not int", paramName))
	}

	return 0, newInvalidParameterError(fmt.Sprintf("param %s is not found", paramName))
}

func optionalIntParamInMap(m map[string]interface{}, paramName string, defaultValue int64) int64 {
	v, ok := m[paramName]
	if ok {
		i, ok := v.(float64)
		if ok {
			return int64(i)
		}
	}

	return defaultValue
}

func MustFloatParam(params httprouter.Params, paramName string) (float64, *Error) {
	float_str := params.ByName(paramName)
	if float_str == "" {
		return 0.0, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	f, err := strconv.ParseFloat(float_str, 64)
	if err != nil {
		return 0.0, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, float_str))
	}

	return f, nil
}

func _mustStringParam(paramName string, str string, paramType int) (string, *Error) {
	if str == "" {
		return "", newInvalidParameterError(fmt.Sprintf("param: %s can't be blank", paramName))
	}

	if paramType == StringParamType_UrlWord {
		str2, ok := common.ValidateUrlWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("param: %s=%s", paramName, str))
		}
		str = str2
	} else if paramType == StringParamType_UnicodeUrlWord {
		str2, ok := common.ValidateUnicodeUrlWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("param: %s=%s", paramName, str))
		}
		str = str2
	} else if paramType == StringParamType_Email {
		str2, ok := common.ValidateEmail(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("param: %s=%s", paramName, str))
		}
		str = str2
	} else { // if paramType == StringParamType_General
		str2, ok := common.ValidateGeneralWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("param: %s=%s", paramName, str))
		}
		str = str2
	}

	return str, nil
}

func MustStringParamInPath(params httprouter.Params, paramName string, paramType int) (string, *Error) {
	return _mustStringParam(paramName, params.ByName(paramName), paramType)
}

func MustStringParamInQuery(r *http.Request, paramName string, paramType int) (string, *Error) {
	return _mustStringParam(paramName, r.Form.Get(paramName), paramType)
}

func MustStringParamInMap(m map[string]interface{}, paramName string, paramType int) (string, *Error) {
	v, ok := m[paramName]
	if ok {
		str, ok := v.(string)
		if ok {
			return _mustStringParam(paramName, str, paramType)
		}

		return "", newInvalidParameterError(fmt.Sprintf("param %s is not string", paramName))
	}

	return "", newInvalidParameterError(fmt.Sprintf("param %s is not found", paramName))
}

func optionalTimeParamInQuery(r *http.Request, paramName string, timeLayout string, defaultTime time.Time) time.Time {
	str := r.Form.Get(paramName)
	if str == "" {
		return defaultTime
	}

	t, err := time.Parse(timeLayout, str)
	if err != nil {
		return defaultTime
	}

	return t
}

//======================================================
//
//======================================================

//======================================================
//
//======================================================

//func MustCurrentUserName(r *http.Request) (string, *Error) {
//	username, _, ok := r.BasicAuth()
//	if !ok {
//		return "", GetError(ErrorCodeAuthFailed)
//	}
//
//	return username, nil
//}

func MustCurrentUserName(r *http.Request) (string, *Error) {
	username := r.Header.Get("User")
	if username == "" {
		return "", GetError(ErrorCodeAuthFailed)
	}

	// needed?
	//username, ok = common.ValidateEmail(username)
	//if !ok {
	//	return "", newInvalidParameterError(fmt.Sprintf("user (%s) is not valid", username))
	//}

	return username, nil
}

func getCurrentUserName(r *http.Request) string {
	return r.Header.Get("User")
}

func MustRepoName(params httprouter.Params) (string, *Error) {
	repo_name, e := MustStringParamInPath(params, "repname", StringParamType_UrlWord)
	if e != nil {
		return "", e
	}

	return repo_name, nil
}

func MustRepoAndItemName(params httprouter.Params) (repo_name string, item_name string, e *Error) {
	repo_name, e = MustStringParamInPath(params, "repname", StringParamType_UrlWord)
	if e != nil {
		return
	}

	item_name, e = MustStringParamInPath(params, "itemname", StringParamType_UrlWord)
	if e != nil {
		return
	}

	return
}

func optionalOffsetAndSize(r *http.Request, defaultSize int64, minSize int64, maxSize int64) (int64, int) {
	page := optionalIntParamInQuery(r, "page", 0)
	if page < 1 {
		page = 1
	}
	page -= 1
	
	if minSize < 1 {
		minSize = 1
	}
	if maxSize < 1 {
		maxSize = 1
	}
	if minSize > maxSize {
		minSize, maxSize = maxSize, minSize
	}
	
	size := optionalIntParamInQuery(r, "size", defaultSize)
	if size < minSize {
		size = minSize
	} else if size > maxSize {
		size = maxSize
	}
	
	return page * size, int(size)
}
