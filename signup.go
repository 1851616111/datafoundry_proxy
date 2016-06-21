package main

import (
	"errors"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	ETCDUSERPREFIX = ETCDPREFIX + "users/"
)

func SignUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//fmt.Println("method:",r.Method)
	//fmt.Println("scheme", r.URL.Scheme)

	usr := new(UserInfo)
	if err := parseRequestBody(r, usr); err != nil {
		glog.Error("read request body error.", err)
		http.Error(w, "", http.StatusBadRequest)
	} else {
		glog.Infof("%+v", usr)
		if err := usr.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if exist, err := usr.IfExist(); !exist && err == nil {

			if err = usr.Create(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				http.Error(w, "", http.StatusOK)
			}
		} else if exist {
			http.Error(w, "user is already exist.", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (usr *UserInfo) IfExist() (bool, error) {

	_, err := dbstore.GetValue(ETCDUSERPREFIX + usr.Username)
	if err != nil {
		if checkIfNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (usr *UserInfo) Validate() error {
	if len(usr.Username) == 0 ||
		len(usr.Email) == 0 ||
		len(usr.Password) == 0 {
		return errors.New("err.")
	}
	return nil
}

func (usr *UserInfo) Create() error {
	return dbstore.SetValue(ETCDUSERPREFIX+usr.Username, usr, false)
}
