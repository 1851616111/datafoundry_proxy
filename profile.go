package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}
	r.ParseForm()
	switch r.Method {
	case "GET":
		if userProfile, err := getProfile(username); err != nil {
			RespError(w, err, http.StatusInternalServerError)
		} else {

			user := new(UserInfo)
			if err = json.Unmarshal([]byte(userProfile.(string)), user); err != nil {
				glog.Error(err)
			}
			glog.Infoln(user)
			RespOK(w, user)
		}
	case "PUT":
		glog.Infoln("put called..")
		usr := new(UserInfo)
		if err := parseRequestBody(r, usr); err != nil {
			glog.Error("read request body error.", err)
			RespError(w, err, http.StatusBadRequest)
		} else {
			if usr.Username != username {
				RespError(w, ldpErrorNew(ErrCodeUserModifyNotAllowed), http.StatusBadRequest)
			}
			if err = usr.Update(); err != nil {
				RespError(w, err, http.StatusBadRequest)
			} else {
				RespOK(w, usr)
			}
		}

	}
	return

}

func checkToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	glog.Info(auth)
	if len(auth) == 0 {
		return auth, false
	}
	return auth, true
}

func getProfile(user string) (etcdvalue interface{}, err error) {
	glog.Info("user", user)
	if len(user) == 0 {
		return nil, ldpErrorNew(ErrCodeNotFound)
	}

	return dbstore.GetValue(etcdProfilePath(user))

}

func authedIdentities(r *http.Request) (string, error) {
	if token, ok := checkToken(r); !ok {
		return "", ldpErrorNew(ErrCodeUnauthorized)
	} else {
		glog.Info(token)

		if username, err := getDFUserame(token); err != nil {
			return "", err
		} else {
			return username, nil
		}
	}
}
