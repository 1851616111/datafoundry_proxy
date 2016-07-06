package messages

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	//"log"
	//"io/ioutil"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"time"

	"github.com/julienschmidt/httprouter"
)

var router *httprouter.Router

func init() {
	//>> params
	Debug = true
	initDB()
	//<<

	router = httprouter.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = false

	initRouter(router)
}

//==================================================================
//
//==================================================================

func ParseJsonToMap(jsonByes []byte) (map[string]interface{}, error) {
	if jsonByes == nil {
		return nil, errors.New("jsonBytes can't be nil")
	}
	var v interface{}
	err := json.Unmarshal(jsonByes, &v)
	if err != nil {
		return nil, err
	}
	json_map, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New("parse json error")
	}

	return json_map, nil
}

//==================================================================
//
//==================================================================

type HttpRequestCase struct {
	name         string
	requestInput string

	expectedStatusCode int
	expectedErrorNO    int
	expectedBody       string
}

func newHttpRequestCase(caseName string, requestHeader string, requestBody string, expectedStatusCode int, expectedErrno int, expectedBody string) *HttpRequestCase {
	a_case := &HttpRequestCase{
		name:               caseName,
		expectedStatusCode: expectedStatusCode,
		expectedErrorNO:    expectedErrno,
		expectedBody:       expectedBody,
	}

	if requestBody == "" {
		a_case.requestInput = fmt.Sprintf("%s\n\n", requestHeader)
	} else { // there must be "Content-Length %d" in requestHeader
		requestHeader = fmt.Sprintf(requestHeader, len(requestBody)+1) // +0 and +2 are both bad

		a_case.requestInput = fmt.Sprintf("%s\n\n%s", requestHeader, requestBody)
	}

	return a_case
}

func _testCases(t *testing.T, cases []*HttpRequestCase) {
	for _, cs := range cases {
		//t.Logf("\n%s\n", cs.requestInput)

		r, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer([]byte(cs.requestInput))))
		if err != nil {
			t.Errorf("[%s] error: %s", cs.name, err.Error())
			continue
		}

		handler, params, _ := router.Lookup(r.Method, r.URL.EscapedPath())
		if handler == nil {
			t.Errorf("[%s] handler == nil", cs.name)
			continue
		}
		//if params == nil {
		//	t.Errorf("[%s] params == nil", cs.name)
		//	continue
		//}
		w := httptest.NewRecorder()
		handler(w, r, params)

		if w.Code != cs.expectedStatusCode {
			t.Errorf("[%s] w.Code (%d) != %d. \n======= response.Body ====== \n%s", cs.name, w.Code, cs.expectedStatusCode, string(w.Body.Bytes()))
			continue
		}

		if w.Code != http.StatusOK && w.Code != http.StatusUnauthorized && w.Code != http.StatusBadRequest {
			continue // APIs only return the 3 possible codes
		}

		result, err := ParseJsonToMap(w.Body.Bytes())
		if err != nil {
			t.Errorf("[%s] error: %s", cs.name, err.Error())
			continue
		}

		errno, e := mustIntParamInMap(result, "code")
		if e != nil {
			t.Errorf("[%s] e=%d#%s", cs.name, e.code, e.message)
			continue
		}

		if ErrorCodeAny != cs.expectedErrorNO && int(errno) != cs.expectedErrorNO {
			t.Errorf("[%s] result.errno (%d) != %d. \n======= response.Body ====== \n%s", cs.name, errno, cs.expectedErrorNO, string(w.Body.Bytes()))
			continue
		}

		if cs.expectedBody != "" && cs.expectedBody != string(w.Body.Bytes()) {
			t.Errorf("[%s] expectedBody != response.Body\nexpectedBody=\n%s\nresponse.Body=\n%s\n====== request = :\n%s", cs.name, cs.expectedBody, string(w.Body.Bytes()), cs.requestInput)
			continue
		}
	}
}

//==================================================================
//
//==================================================================

var nowString = time.Now().Format("2006-01-02_15-04-05_999999")

// the ordr of cases are important!!!
var All_Cases = []*HttpRequestCase{

	// ChangeStarStatus

/*
	newHttpRequestCase("Create Notification 1",
		`POST /notifications HTTP/1.1
Content-Type: application/json
Content-Length: %d
Accept: application/json
User: zhang3@aa.com
`,
		`{
    "type": "apply_whitelist",
    "data": {
		"repname": "repo001",
    	"itemname": "item123"
	}
}`,
		http.StatusOK, ErrorCodeNone,
		``),

	newHttpRequestCase("Create Notification 2 (type is missing)",
		`POST /notifications HTTP/1.1
Content-Type: application/json
Content-Length: %d
Accept: application/json
User: zhang3@aa.com
`,
		`{
    "type": "apply_whitelist",
    "data": {
		"repname": "repo001",
    	"itemname": "item123"
	}
}`,
		http.StatusOK, ErrorCodeNone,
		``),
*/

	newHttpRequestCase("Get Comments 1",
		`GET /notification_stat HTTP/1.1
Content-Type: application/json
Accept: application/json
User: zhang3@aa.com

`,
		``,
		http.StatusOK, ErrorCodeNone,
		``),

	newHttpRequestCase("Get Comment Stat 1",
		`DELETE /notification_stat HTTP/1.1
Content-Type: application/json
Accept: application/json
User: zhang3@aa.com

`,
		``,
		http.StatusOK, ErrorCodeNone,
		``),
}

func TestAPIs(t *testing.T) {
	_testCases(t, All_Cases)
}
