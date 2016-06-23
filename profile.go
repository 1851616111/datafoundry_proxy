package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Profile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	if username, err := authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		if userProfile, err := getProfile(username); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			glog.Infoln(userProfile)
			http.Error(w, userProfile.(string), http.StatusOK)
		}
	}

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
