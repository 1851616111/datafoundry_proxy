package main

import (
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func PasswordModify(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error
	if username, err = authedIdentities(r); err != nil {
		RespError(w, err.Error(), http.StatusBadRequest)
		return
	}
	password := new(Password)
	if err := parseRequestBody(r, password); err != nil {
		glog.Error("read request body error.", err)
		RespError(w, err.Error(), http.StatusBadRequest)
	} else {
		if err = password.Modify(username); err != nil {
			RespError(w, err.Error(), http.StatusBadRequest)
		} else {
			RespOK(w, nil)
		}
	}
}

func (p *Password) Modify(username string) error {
	if len(p.NewPassword) == 0 || len(p.OldPassword) == 0 {
		return errors.New("password can't be empty.")
	}
	if len(p.NewPassword) > 12 || len(p.NewPassword) < 8 {
		return errors.New("password length must be 8 to 12 characters.")
	}

	return p.ModifyPasswordLdap(username)

}

func (p *Password) ModifyPasswordLdap(username string) error {

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s", LdapEnv.Get(LDAP_HOST_ADDR)))
	if err != nil {
		glog.Fatal(err)
		return err
	}
	defer l.Close()

	err = l.Bind(ldapUser(username), p.OldPassword)
	if err != nil {
		glog.Error(err)
		return err
	} else {
		glog.Info("bind successfully.")
	}

	passwordModifyRequest := ldap.NewPasswordModifyRequest("", p.OldPassword, p.NewPassword)
	_, err = l.PasswordModify(passwordModifyRequest)

	if err != nil {
		glog.Fatalf("Password could not be changed: %s", err.Error())
		return err
	} else {
		glog.Infof("password modify successfuly [%s]", username)
	}
	return nil
}
