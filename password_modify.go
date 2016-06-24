package main

import (
	_ "errors"
	_ "fmt"
	_ "github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func PasswordModify(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	if _, err := authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	password := new(Password)
	if err := parseRequestBody(r, password); err != nil {
		glog.Error("read request body error.", err)
		http.Error(w, "", http.StatusBadRequest)
	} else {
		if err := password.Modify(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (p *Password) Modify() error {
	return nil

}
