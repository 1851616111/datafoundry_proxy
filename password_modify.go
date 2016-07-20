package main

import (
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
		RespError(w, err, http.StatusUnauthorized)
		return
	}
	password := new(Password)
	if err := parseRequestBody(r, password); err != nil {
		glog.Error("read request body error.", err)
		RespError(w, err, http.StatusBadRequest)
	} else {
		if err = password.Modify(username); err != nil {
			RespError(w, err, http.StatusBadRequest)
		} else {
			RespOK(w, nil)
		}
	}
}

func (p *Password) Modify(username string) error {
	if len(p.NewPassword) == 0 || len(p.OldPassword) == 0 {
		return ldpErrorNew(ErrCodePasswordEmpty)
	}
	if len(p.NewPassword) > 12 || len(p.NewPassword) < 8 {
		return ldpErrorNew(ErrCodePasswordLengthMismatch)
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
