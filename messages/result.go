package messages

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

//======================================================
//
//======================================================

var Json_ErrorBuildingJson []byte

func getJsonBuildingErrorJson() []byte {
	if Json_ErrorBuildingJson == nil {
		Json_ErrorBuildingJson = []byte(fmt.Sprintf(`{"code": %d, "msg": %s}`, ErrorJsonBuilding.code, ErrorJsonBuilding.message))
	}

	return Json_ErrorBuildingJson
}

type Result struct {
	Code uint        `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// if data only has one item, then the item key will be ignored.
func JsonResult(w http.ResponseWriter, statusCode int, e *Error, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if e == nil {
		e = ErrorNone
	}
	result := Result{Code: e.code, Msg: e.message, Data: data}
	jsondata, err := json.Marshal(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(getJsonBuildingErrorJson()))
	} else {
		w.WriteHeader(statusCode)
		w.Write(jsondata)
	}
}

type QueryListResult struct {
	Total   int64       `json:"total"`
	Results interface{} `json:"results"`
}

func newQueryListResult(count int64, results interface{}) *QueryListResult {
	return &QueryListResult{Total: count, Results: results}
}