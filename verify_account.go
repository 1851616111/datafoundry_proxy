package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func VerifyAccount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	r.ParseForm()
	token := ps.ByName("token")
	if len(token) == 0 {
		http.Error(w, "invalid token.", http.StatusBadRequest)
		return
	}
	user, err := dbstore.GetValue(etcdGeneratePath(ETCDUserVerify, token))
	if err != nil {
		if checkIfNotFound(err) {
			glog.Errorf("token %s not exist.", token)
			http.Error(w, "token not exist.", http.StatusNotFound)
		} else {
			glog.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err = activeAccount(user.(string), token); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Error(w, "account verified.", http.StatusOK)
	}
}

func activeAccount(user, token string) error {
	profile, err := dbstore.GetValue(etcdProfilePath(user))
	if err != nil {
		glog.Error(err)
		return err
	}
	userinfo := new(UserInfo)
	if err = json.Unmarshal([]byte(profile.(string)), userinfo); err != nil {
		glog.Error(err)
		return err
	} else {
		userinfo.Status.Active = true
	}
	glog.Warning("TODO: INIT USER, CREATE NEW PROJECT.")

	if err = dbstore.SetValue(etcdProfilePath(userinfo.Username), userinfo, false); err != nil {
		glog.Error(err)
		return err
	} else {
		return dbstore.Delete(etcdGeneratePath(ETCDUserVerify, token), false)
	}

}
