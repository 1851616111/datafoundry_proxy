package main

import (
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
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

	_, err := dbstore.GetValue(etcdProfilePath(usr.Username))
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
	if err := usr.AddToLdap(); err != nil {
		if !checkIfExistldap(err) {
			glog.Fatal(err)
			return err
		} else {
			glog.Infof("user %s already exist on ldap.", usr.Username)
		}
	}
	return usr.AddToEtcd()
}

func (usr *UserInfo) Update(username string) error {
	return dbstore.SetValue(etcdProfilePath(username), usr, false)
}

func (usr *UserInfo) AddToEtcd() error {
	pass := usr.Password
	usr.Password = ""
	usr.Status = UserStatusInactive
	err := dbstore.SetValue(etcdProfilePath(usr.Username), usr, false)
	usr.Password = pass
	return err
}

func (usr *UserInfo) AddToLdap() error {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s", LdapEnv.Get(LDAP_HOST_ADDR)))
	if err != nil {
		glog.Fatal(err)
	}
	defer l.Close()

	err = l.Bind(LdapEnv.Get(LDAP_ADMIN_USER), LdapEnv.Get(LDAP_ADMIN_PASSWORD))
	if err != nil {
		glog.Error(err)
	} else {
		glog.Info("bind successfully.")
	}

	request := ldap.NewAddRequest(fmt.Sprintf(LdapEnv.Get(LDAP_BASE_DN), usr.Username))
	request.Attribute("objectclass", []string{"inetOrgPerson"})
	request.Attribute("sn", []string{usr.Username})
	request.Attribute("uid", []string{usr.Username})
	request.Attribute("userpassword", []string{usr.Password})
	request.Attribute("mail", []string{usr.Email})
	request.Attribute("ou", []string{"DataFoundry"})

	err = l.Add(request)
	if err != nil {
		return err
		/*
			if !checkIfExistldap(err) {
				glog.Fatal(err)
				return err
			} else {
				glog.Info("user aready exist.")
				return errors.New("user already exist.")
			}*/
	} else {
		glog.Info("add to ldap successfully.")
	}
	return nil
}

func checkIfExistldap(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*ldap.Error); ok && e.ResultCode == ldap.LDAPResultEntryAlreadyExists {
		return true
	}

	return false
}
