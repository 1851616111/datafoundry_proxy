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
		http.Error(w, "", http.StatusBadRequest)
	} else {
		glog.Infof("%+v", usr)
		if err := usr.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if exist, err := usr.IfExist(); !exist && err == nil {

			if err = usr.Create(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				if usr.Status.FromLdap {
					http.Error(w, "user already exist on ldap.", 422)
				} else {
					http.Error(w, "", http.StatusOK)

					go func() {
						if err := usr.SendVerifyMail(); err != nil {
							glog.Error(err)
						}
					}()
				}
			}
		} else if exist {
			http.Error(w, "user is already exist.", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
