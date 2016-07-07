package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-ldap/ldap"
	"github.com/golang/glog"
	oapi "github.com/openshift/origin/pkg/user/api/v1"
	"io/ioutil"
	"net/http"
	"time"
	
	"github.com/asiainfoLDP/datafoundry_proxy/messages"
)

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

func (usr *UserInfo) CreateOrg(org *Orgnazition) (neworg *Orgnazition, err error) {
	org.CreateTime = time.Now().Format(time.RFC3339)
	org.CreateBy = usr.Username
	org.ID = usr.Username + "-org-" + generatedOrgName(8)
	member := OrgMember{
		MemberName:   usr.Username,
		IsAdmin:      true,
		JoinedAt:     org.CreateTime,
		PrivilegedBy: usr.Username,
		Status:       OrgMemberStatusjoined,
	}
	org.MemberList = append(org.MemberList, member)
	if err = dbstore.SetValue(etcdOrgPath(org.ID), org, false); err != nil {
		glog.Error(err)
	} else {
		orgbrief := OrgBrief{OrgID: org.ID, OrgName: org.Name}
		if usr.OrgList != nil {
			usr.OrgList = append(usr.OrgList, orgbrief)
		} else {
			usr.OrgList = []OrgBrief{orgbrief}
		}
		if err = usr.Update(); err != nil {
			glog.Error(err)
			return nil, err
		}
		return org, nil
	}
	return neworg, err
}

func (usr *UserInfo) CheckIfOrgExist(orgName string) bool {
	for _, org := range usr.OrgList {
		if org.OrgName == orgName {
			return true
		}
	}
	return false
}

func (usr *UserInfo) DeleteOrgFromList(orgID string) *UserInfo {
	for idx, org := range usr.OrgList {
		if org.OrgID == orgID {
			usr.OrgList = append(usr.OrgList[:idx], usr.OrgList[idx+1:]...)
		}
	}
	return usr
}

func (usr *UserInfo) AddOrgToList(org *Orgnazition) *UserInfo {
	orgbrief := new(OrgBrief)
	orgbrief.OrgID = org.ID
	orgbrief.OrgName = org.Name

	usr.OrgList = append(usr.OrgList, *orgbrief)

	return usr
}

func (usr *UserInfo) CheckIfOrgExistByID(id string) bool {
	for _, org := range usr.OrgList {
		if org.OrgID == id {
			return true
		}
	}
	return false
}

func (user *UserInfo) ListOrgs() (*OrgnazitionList, error) {
	orgList := new(OrgnazitionList)
	for _, orgbrief := range user.OrgList {
		org := new(Orgnazition)
		org.ID = orgbrief.OrgID
		if orgnazition, err := org.Get(); err == nil {
			orgList.Orgnazitions = append(orgList.Orgnazitions, *orgnazition)
		}

	}
	return orgList, nil
}

func (user *UserInfo) OrgLeave(orgID string) (err error) {
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if org.IsLastAdmin(user.Username) {
			return errors.New("orgnazition needs at least one admin.")
		}
		org = org.RemoveMember(user.Username)
		if _, err = org.Update(); err != nil {
			glog.Error(err)
			return err
		} else {
			if user, err = user.Get(); err != nil {
				glog.Error(err)
				return
			} else {
				user = user.DeleteOrgFromList(orgID)
				return user.Update()
			}
		}

	}

	return
}

func (user *UserInfo) OrgJoin(orgID string) (err error) {
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if org.IsLastAdmin(user.Username) {
			return errors.New("orgnazition needs at least one admin.")
		}
		org = org.MemberJoined(user.Username)
		if _, err = org.Update(); err != nil {
			glog.Error(err)
			return err
		} else {
			if user, err = user.Get(); err != nil {
				glog.Error(err)
				return
			} else {
				user = user.AddOrgToList(org)
				return user.Update()
			}
		}
	}
	return
}

func (user *UserInfo) OrgInvite(member *OrgMember, orgID string) (err error) {
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			return errors.New("permission denied.")
		}
		if org.IsMemberExist(member) {
			return errors.New("user is already in the orgnazition.")
		}
		minfo := new(UserInfo)
		minfo.Username = member.MemberName
		if ok, err := minfo.IfExist(); !ok {
			if err != nil {
				return err
			}
			return errors.New("user not registered yet.")
		}
		if member.IsAdmin {
			member.PrivilegedBy = user.Username
		}
		member.Status = OrgMemberStatusInvited
		org.MemberList = append(org.MemberList, *member)
	}
	if err = dbstore.SetValue(etcdOrgPath(org.ID), org, false); err != nil {
		glog.Error(err)
	}
	return
}

func (user *UserInfo) OrgRemoveMember(member *OrgMember, orgID string) (err error) {
	if user.Username == member.MemberName {
		return errors.New("can't remove yourself.")
	}
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			return errors.New("permission denied.")
		}
		if !org.IsMemberExist(member) {
			return errors.New("no such user in the orgnazition.")
		}
		org = org.RemoveMember(member.MemberName)
		if _, err = org.Update(); err != nil {
			glog.Error(err)
			return err
		} else {
			minfo := new(UserInfo)
			minfo.Username = member.MemberName
			if minfo, err = minfo.Get(); err != nil {
				glog.Error(err)
				return
			} else {
				minfo = minfo.DeleteOrgFromList(orgID)
				return minfo.Update()
			}
		}
	}

	return
}
func (user *UserInfo) OrgPrivilege(member *OrgMember, orgID string) (err error) {
	if user.Username == member.MemberName {
		return errors.New("can't privilege yourself.")
	}
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			return errors.New("permission denied.")
		}

		if !org.IsMemberExist(member) {
			return errors.New("can't find such username in this orgnazition.")
		}

		for idx, oldMember := range org.MemberList {
			if oldMember.MemberName == member.MemberName {
				org.MemberList[idx].IsAdmin = member.IsAdmin
				org.MemberList[idx].PrivilegedBy = user.Username
				/*
					if member.IsAdmin {
						org.MemberList[idx].PrivilegedBy = user.Username
					} else {
						org.MemberList[idx].PrivilegedBy = ""
					}
				*/
				if org, err = org.Update(); err == nil {
					return
				} else {
					glog.Error(err)
					return
				}
			}
		}
		return errors.New("no such user.")
	}
	return
}

func (usr *UserInfo) Update() error {
	return dbstore.SetValue(etcdProfilePath(usr.Username), usr, false)
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
	return messages.SendMail([]string{usr.Email}, []string{}, bccEmail, Subject, message, true)
}

func (user *UserInfo) InitUserProject(token string) (err error) {
	project_url := DF_HOST + "/oapi/v1/projectrequests"

	project := new(oapi.ProjectRequest)
	project.Name = user.Username
	if reqbody, err := json.Marshal(project); err != nil {
		glog.Error(err)
	} else {

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		req, _ := http.NewRequest("POST", project_url, bytes.NewBuffer(reqbody))
		req.Header.Set("Authorization", token)
		//log.Println(req.Header, bearer)

		resp, err := client.Do(req)
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infoln(req.Host, req.Method, req.URL.RequestURI(), req.Proto, resp.StatusCode)
			b, _ := ioutil.ReadAll(resp.Body)
			glog.Infoln(string(b))
			if resp.StatusCode == http.StatusOK {
				user.Status.Initialized = true
				err = dbstore.SetValue(etcdProfilePath(user.Username), user, false)
			}
		}
	}

	return
}
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

func (user *UserInfo) Get() (userinfo *UserInfo, err error) {
	glog.Info("user", user)
	if len(user.Username) == 0 {
		return nil, ErrNotFound
	}

	if obj, err := dbstore.GetValue(etcdProfilePath(user.Username)); err != nil {
		return nil, err
	} else {
		glog.Info(obj.(string))
		u := new(UserInfo)
		if err = json.Unmarshal([]byte(obj.(string)), u); err != nil {
			glog.Error(err)
			return nil, err
		} else {
			return u, nil
		}
	}

}
