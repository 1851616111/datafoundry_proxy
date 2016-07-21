package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	etcd "github.com/coreos/etcd/client"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
)

func httpPost(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("POST", url, body, credential...)
}

func httpPUT(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("PUT", url, body, credential...)
}

func httpPATCH(url string, body []byte, credential ...string) ([]byte, error) {
	return httpAction("PATCH", url, body, credential...)
}

func httpGet(url string, credential ...string) ([]byte, error) {
	glog.Infoln(url, credential)
	var resp *http.Response
	var err error

	if len(credential) == 2 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s", url, err)
		}
		req.Header.Set(credential[0], credential[1])

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		switch resp.StatusCode {
		case 404:
			return nil, ldpErrorNew(ErrCodeNotFound)
		case 200:
			return ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func httpGetFunc(url string, f func(resp *http.Response), credential ...string) ([]byte, error) {

	var resp *http.Response
	var err error
	if len(credential) == 2 {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("[http] err %s, %s", url, err)
		}
		req.Header.Set(credential[0], credential[1])

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}

		if f != nil {
			f(resp)
		}

		switch resp.StatusCode {
		case 404:
			return nil, ldpErrorNew(ErrCodeNotFound)
		case 200:
			return ioutil.ReadAll(resp.Body)
		}
		if resp.StatusCode < 200 || resp.StatusCode > 300 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	} else {
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("http get err:%s", err.Error())
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("[http get] status err %s, %d", url, resp.StatusCode)
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func httpAction(method, url string, body []byte, credential ...string) ([]byte, error) {
	fmt.Println(method, url, string(body), credential)
	var resp *http.Response
	var err error
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s", url, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if len(credential) == 2 {
		req.Header.Set(credential[0], credential[1])
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[http] err %s, %s", url, err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("[http] read err %s, %s", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		glog.Info("request err %s", string(b))
		return nil, fmt.Errorf("[http] status err %s, %d", url, resp.StatusCode)
	}

	return b, nil
}

func httpDelete(url string, credential ...string) ([]byte, error) {
	return httpAction("DELETE", url, nil, credential...)
}

func retHttpCodef(code, bodyCode int, w http.ResponseWriter, format string, a ...interface{}) {

	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":"%s"}`, bodyCode, fmt.Sprintf(format, a...))

	fmt.Fprintf(w, msg)
	return
}

func retHttpCode(code int, bodyCode int, w http.ResponseWriter, a ...interface{}) {
	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":"%s"}`, bodyCode, fmt.Sprint(a...))

	fmt.Fprintf(w, msg)
	return
}

func retHttpCodeJson(code int, bodyCode int, w http.ResponseWriter, a ...interface{}) {
	w.WriteHeader(code)
	msg := fmt.Sprintf(`{"code":%d,"msg":%s}`, bodyCode, fmt.Sprint(a...))

	fmt.Fprintf(w, msg)
	return
}

func RespError(w http.ResponseWriter, err error, httpCode int) {
	resp := genRespJson(httpCode, err)

	if body, err := json.MarshalIndent(resp, "", "  "); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpCode)
		w.Write(body)
	}

}

func RespOK(w http.ResponseWriter, data interface{}) {
	if data == nil {
		data = genRespJson(http.StatusOK, nil)
	}

	if body, err := json.MarshalIndent(data, "", "  "); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func genRespJson(httpCode int, err error) *APIResponse {
	resp := new(APIResponse)
	var msgCode int
	var message string

	if err == nil {
		msgCode = ErrCodeOK
		message = ErrText(msgCode)
	} else {
		if e, ok := err.(*ldap.Error); ok {
			msgCode = int(e.ResultCode) + LDAPMagicNumber
			message = ldap.LDAPResultCodeMap[e.ResultCode]
		} else if e, ok := err.(etcd.Error); ok {
			msgCode = e.Code + EtcdMagicNumber
			message = e.Message
		} else if e, ok := err.(*etcd.Error); ok {
			msgCode = e.Code + EtcdMagicNumber
			message = e.Message
		} else if e, ok := err.(LdpError); ok {
			msgCode = e.Code
			message = e.Message
		} else if e, ok := err.(*LdpError); ok {
			msgCode = e.Code
			message = e.Message
		} else {
			msgCode = ErrCodeUnknownError
			message = e.Error()
		}
	}

	resp.Code = msgCode
	resp.Message = message
	resp.Status = http.StatusText(httpCode)
	return resp
}
