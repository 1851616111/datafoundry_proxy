package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.ParseForm()
	switch r.Method {
	case "GET":
		if userProfile, err := getProfile(username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			glog.Infoln(userProfile)
			http.Error(w, userProfile.(string), http.StatusOK)
		}
	case "PUT":
		glog.Infoln("put called..")
		usr := new(UserInfo)
		if err := parseRequestBody(r, usr); err != nil {
			glog.Error("read request body error.", err)
			http.Error(w, err.Error(), 422)
		} else {
			if usr.Username != username {
				http.Error(w, "", http.StatusBadRequest)
			}
			if err = usr.Update(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, "", http.StatusOK)
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
		return nil, ErrNotFound
	}

	return dbstore.GetValue(etcdProfilePath(user))

}

func authedIdentities(r *http.Request) (string, error) {
	if token, ok := checkToken(r); !ok {
		return "", ErrAuthorizedRequired
	} else {
		glog.Info(token)

		if username, err := getDFUserame(token); err != nil {
			return "", err
		} else {
			return username, nil
		}
	}
}
