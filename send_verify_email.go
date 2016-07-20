package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func SendVerifyMail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}
	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	go func() {
		if err := userinfo.SendVerifyMail(); err != nil {
			glog.Error(err)
		}
	}()

	RespOK(w, userinfo)
}
