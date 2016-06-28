package main

import (
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
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

func (usr *UserInfo) IfExist() (bool, error) {

	_, err := dbstore.GetValue(etcdProfilePath(usr.Username))
	if err != nil {
		glog.Infoln(err)
		if checkIfNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (usr *UserInfo) Validate() error {
	if ok, reason := ValidateUserName(usr.Username, false); !ok {
		return errors.New(reason)
	}

	if ok, reason := ValidateEmail(usr.Email); !ok {
		return errors.New(reason)
	}

	if len(usr.Password) > 12 ||
		len(usr.Password) < 8 {
		return errors.New("password must be between 8 and 12 characters long.")
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
			usr.Status.FromLdap = true
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
	usr.Status.Phase = UserStatusInactive
	usr.Status.Active = false
	usr.Status.Initialized = false
	usr.CreateTime = time.Now().Format(time.RFC3339)
	err := dbstore.SetValue(etcdProfilePath(usr.Username), usr, false)
	usr.Password = pass
	return err
}

func (usr *UserInfo) AddToLdap() error {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s", LdapEnv.Get(LDAP_HOST_ADDR)))
	if err != nil {
		glog.Infoln(err)
		return err
	}
	defer l.Close()

	err = l.Bind(LdapEnv.Get(LDAP_ADMIN_USER), LdapEnv.Get(LDAP_ADMIN_PASSWORD))
	if err != nil {
		glog.Infoln(err)
		return err
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

func (usr *UserInfo) SendVerifyMail() error {
	verifytoken, err := usr.AddToVerify()
	if err != nil {
		glog.Error(err)
		return err
	}
	link := httpAddrMaker(DataFoundryEnv.Get(DATAFOUNDRY_API_ADDR)) + "/verify_account/" + verifytoken
	message := fmt.Sprintf(Message, usr.Username, link)
	return SendMail([]string{usr.Email}, []string{}, bccEmail, Subject, message, true)
}

var Subject string = "Welcome to Datafoundry"
var Message string = `Hello %s, <br />please click <a href="%s">link</a> to verify your account, the activation link will be expire after 24 hours.`

func (user *UserInfo) AddToVerify() (verifytoken string, err error) {
	verifytoken, err = genRandomToken()
	if err != nil {
		glog.Error(err)
		return
	}
	glog.Infoln("token:", verifytoken, "user:", user.Username)
	err = dbstore.SetValuebyTTL(etcdGeneratePath(ETCDUserVerify, verifytoken), user.Username, time.Hour*24)
	return
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
