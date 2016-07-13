package main

import (
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func SignUp(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//fmt.Println("method:",r.Method)
	//fmt.Println("scheme", r.URL.Scheme)

	usr := new(UserInfo)
	if err := parseRequestBody(r, usr); err != nil {
		glog.Error("read request body error.", err)
		RespError(w, err.Error(), http.StatusBadRequest)
	} else {
		glog.Infof("%+v", usr)
		if err := usr.Validate(); err != nil {
			RespError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if exist, err := usr.IfExist(); !exist && err == nil {

			if usr, err = usr.Create(); err != nil {
				RespError(w, err.Error(), http.StatusInternalServerError)
			} else {
				if usr.Status.FromLdap {
					RespError(w, "user already exist on ldap.", http.StatusAccepted)
				} else {
					RespOK(w, usr)

					go func() {
						if err := usr.SendVerifyMail(); err != nil {
							glog.Error(err)
						}
					}()
				}
			}
		} else if exist {
			RespError(w, "user already exist.", http.StatusConflict)
		} else {
			RespError(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
