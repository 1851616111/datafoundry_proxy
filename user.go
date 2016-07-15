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
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"net/http"
	"strings"
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

func (usr *UserInfo) Create() (*UserInfo, error) {
	if err := usr.AddToLdap(); err != nil {
		if !checkIfExistldap(err) {
			glog.Fatal(err)
			return usr, err
		} else {
			glog.Infof("user %s already exist on ldap.", usr.Username)
			usr.Status.FromLdap = true
		}
	}
	return usr.AddToEtcd()
}

func (user *UserInfo) DeleteOrg(orgID string) (org *Orgnazition, err error) {
	org = new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			err = errors.New("permission denied.")
			return
		}
		if len(org.MemberList) > 1 {
			err = errors.New("members still in orgnazition.")
			return
		}
		creater := org.CreateBy
		if err = user.OpenshiftDeleteProject(org); err != nil {
			glog.Error(err)
		} else {
			user = user.DeleteOrgFromList(orgID)
			if err = user.Update(); err != nil {
				glog.Error(err)
			} else {
				if userProfile, errs := getProfile(creater); err != nil {
					glog.Error(errs)
				} else {

					createrinfo := new(UserInfo)
					if err = json.Unmarshal([]byte(userProfile.(string)), createrinfo); err != nil {
						glog.Error(err)
					} else {
						createrinfo.Status.Quota.OrgUsed -= 1
						createrinfo.Update()
						org.Delete()
					}

				}
			}
		}
	}
	return
}

func (usr *UserInfo) CreateOrg(org *Orgnazition) (neworg *Orgnazition, err error) {
	if usr.Status.Quota.OrgUsed >= usr.Status.Quota.OrgQuota {
		return nil, errors.New(fmt.Sprintf("user can only create %d orgnazition(s)", usr.Status.Quota.OrgQuota))
	}
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
	if err = usr.CreateProject(org); err != nil {
		return nil, err
	} else {
		org.RoleBinding = true
		if _, err = org.Create(); err == nil {

			orgbrief := OrgBrief{OrgID: org.ID, OrgName: org.Name}
			usr.OrgList = append(usr.OrgList, orgbrief)
			// if usr.OrgList != nil {
			// 	usr.OrgList = append(usr.OrgList, orgbrief)
			// } else {
			// 	usr.OrgList = []OrgBrief{orgbrief}
			// }
			usr.Status.Quota.OrgUsed += 1
			if err = usr.Update(); err != nil {
				glog.Error(err)
				return nil, err
			}
			return org, nil
		}
	}
	return neworg, err
}

func (user *UserInfo) CreateProject(org *Orgnazition) (err error) {
	glog.Infoln(user)
	project_url := DF_HOST + "/oapi/v1/projectrequests"

	project := new(oapi.ProjectRequest)
	project.Name = org.ID
	project.DisplayName = org.Name
	if reqbody, err := json.Marshal(project); err != nil {
		glog.Error(err)
	} else {

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		req, _ := http.NewRequest("POST", project_url, bytes.NewBuffer(reqbody))
		req.Header.Set("Authorization", user.token)
		//log.Println(req.Header, bearer)

		resp, err := client.Do(req)
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infoln(req.Host, req.Method, req.URL.RequestURI(), req.Proto, resp.StatusCode)
			b, _ := ioutil.ReadAll(resp.Body)
			glog.Infoln(string(b))
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				err = user.CreateRoleBinding(org, "admin")
				if err == nil {
					return user.CreateRoleBinding(org, "edit")
				} else {
					return err
				}
			} else {
				return errors.New(string(b))
			}
		}
	}

	return
}

func (user *UserInfo) UpdateRoleBinding(org *Orgnazition) (err error) {
	glog.Infoln("create project role...", user.token)

	glog.Infoln(user)
	AdminRoleUrl := DF_HOST + "/oapi/v1/namespaces/" + org.ID + "/rolebindings/admin-" + org.CreateBy
	EditRoleUrl := DF_HOST + "/oapi/v1/namespaces/" + org.ID + "/rolebindings/" + "edit-" + org.CreateBy

	AdminRole := new(oapi.RoleBinding)
	EditRole := new(oapi.RoleBinding)

	AdminRole.RoleRef = kapi.ObjectReference{Name: "admin"}
	AdminRole.Name = "admin-" + org.CreateBy

	EditRole.RoleRef = kapi.ObjectReference{Name: "edit"}
	EditRole.Name = "edit-" + org.CreateBy

	for _, member := range org.MemberList {
		if member.IsAdmin {
			subject := kapi.ObjectReference{Kind: "User", Name: member.MemberName}
			AdminRole.Subjects = append(AdminRole.Subjects, subject)
			AdminRole.UserNames = append(AdminRole.UserNames, member.MemberName)
		} else {
			subject := kapi.ObjectReference{Kind: "User", Name: member.MemberName}
			EditRole.Subjects = append(EditRole.Subjects, subject)
			EditRole.UserNames = append(EditRole.UserNames, member.MemberName)
		}
	}

	var e error
	reason := make(chan error, 2)
	user.OpenshiftUpdateRole(AdminRoleUrl, AdminRole, reason)
	user.OpenshiftUpdateRole(EditRoleUrl, EditRole, reason)
	e = <-reason
	if e != nil {
		err = e
	}
	e = <-reason

	if e != nil {
		err = e
	}
	return
}

func (user *UserInfo) OpenshiftDeleteProject(org *Orgnazition) (err error) {
	ProjectUrl := DF_HOST + "/oapi/v1/projects/" + org.ID
	_, err = httpDelete(ProjectUrl, "Authorization", user.token)
	return err
}
func (user *UserInfo) OpenshiftUpdateRole(url string, role *oapi.RoleBinding, reason chan error) {
	oldRole := new(oapi.RoleBinding)
	method := "PUT"
	if reqbody, err := json.Marshal(role); err != nil {
		glog.Error(err)
		reason <- err
	} else {
		if b, err := httpGet(url, "Authorization", user.token); err == nil {
			err = json.Unmarshal(b, oldRole)
			role.ResourceVersion = oldRole.ResourceVersion
			reqbody, _ = json.Marshal(role)
		} else {
			httpDelete(url, "Authorization", user.token)
			url = splitLastSlash(url)
			method = "POST"
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqbody))
		req.Header.Set("Authorization", user.token)
		//log.Println(req.Header, bearer)

		resp, err := client.Do(req)
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infoln(req.Host, req.Method, req.URL.RequestURI(), req.Proto, resp.StatusCode)
			b, _ := ioutil.ReadAll(resp.Body)
			glog.Infoln(string(b))
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			} else {
				err = errors.New(string(b))
			}
		}
		reason <- err
	}

	return
}

func splitLastSlash(url string) string {
	n := strings.LastIndex(url, "/")
	if n < 0 {
		return url
	}
	return url[:n]
}

func (user *UserInfo) CreateRoleBinding(org *Orgnazition, role string) (err error) {
	glog.Infoln("create project role...", user.token)

	glog.Infoln(user)
	rolebinding_url := DF_HOST + "/oapi/v1/namespaces/" + org.ID + "/rolebindings"

	rolebinding := new(oapi.RoleBinding)
	rolebinding.Name = role + "-" + user.Username
	rolebinding.RoleRef = kapi.ObjectReference{Name: role}
	if reqbody, err := json.Marshal(rolebinding); err != nil {
		glog.Error(err)
	} else {

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		req, _ := http.NewRequest("POST", rolebinding_url, bytes.NewBuffer(reqbody))
		req.Header.Set("Authorization", user.token)
		//log.Println(req.Header, bearer)

		resp, err := client.Do(req)
		if err != nil {
			glog.Error(err)
		} else {
			glog.Infoln(req.Host, req.Method, req.URL.RequestURI(), req.Proto, resp.StatusCode)
			b, _ := ioutil.ReadAll(resp.Body)
			glog.Infoln(string(b))
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
				return nil
			} else {
				return errors.New(string(b))
			}
		}
	}

	return
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
	glog.Infoln(usr)
	for idx, org := range usr.OrgList {
		glog.Infoln(usr)
		if org.OrgID == orgID {
			usr.OrgList = append(usr.OrgList[:idx], usr.OrgList[idx+1:]...)
			return usr
		}
	}
	return usr
}

func (usr *UserInfo) AddOrgToList(org *Orgnazition) *UserInfo {

	for _, b := range usr.OrgList {
		if org.ID == b.OrgID {
			return usr
		}
	}
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
			user = user.DeleteOrgFromList(orgID)
			return user.Update()
		}
	}

	return
}

func (user *UserInfo) OrgJoin(orgID string) (err error) {
	org := new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		// if org.IsLastAdmin(user.Username) {
		// 	return errors.New("orgnazition needs at least one admin.")
		// }
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

func (user *UserInfo) OrgInvite(member *OrgMember, orgID string) (org *Orgnazition, err error) {
	org = new(Orgnazition)
	org.ID = orgID
	var ok bool
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			err = errors.New("permission denied.")
			return
		}
		if org.IsMemberExist(member) {
			if org.MemberStatus(member) == OrgMemberStatusjoined {
				err = errors.New("user is already invited.")
			} else {
				err = errors.New("user is already in the orgnazition.")
			}
			return
		}
		minfo := new(UserInfo)
		minfo.Username = member.MemberName
		if ok, err = minfo.IfExist(); !ok {
			if err == nil {
				err = errors.New("user not registered yet.")
			}
			return
		}
		if member.IsAdmin {
			member.PrivilegedBy = user.Username
		}
		member.Status = OrgMemberStatusInvited
		org.MemberList = append(org.MemberList, *member)
	}
	if err = user.UpdateRoleBinding(org); err != nil {
		return
	}
	_, err = org.Update()
	return
}

func (user *UserInfo) OrgRemoveMember(member *OrgMember, orgID string) (org *Orgnazition, err error) {
	if user.Username == member.MemberName {
		return nil, errors.New("can't remove yourself.")
	}
	org = new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			return nil, errors.New("permission denied.")
		}
		if !org.IsMemberExist(member) {
			return nil, errors.New("no such user in the orgnazition.")
		}
		org = org.RemoveMember(member.MemberName)
		if _, err = org.Update(); err != nil {
			glog.Error(err)
			return nil, err
		} else {
			if err = user.UpdateRoleBinding(org); err != nil {
				return nil, err
			}
			minfo := new(UserInfo)
			minfo.Username = member.MemberName
			if minfo, err = minfo.Get(); err != nil {
				glog.Error(err)
				return
			} else {
				minfo = minfo.DeleteOrgFromList(orgID)
				return org, minfo.Update()
			}
		}
	}

	return
}
func (user *UserInfo) OrgPrivilege(member *OrgMember, orgID string) (org *Orgnazition, err error) {
	// if user.Username == member.MemberName {
	// 	return nil, errors.New("can't privilege yourself.")
	// }
	org = new(Orgnazition)
	org.ID = orgID
	if org, err = org.Get(); err == nil {
		if !org.IsAdmin(user.Username) {
			return nil, errors.New("permission denied.")
		}

		if org.IsLastAdmin(member.MemberName) {
			return nil, errors.New("orgnazition needs at least one admin.")
		}

		if !org.IsMemberExist(member) {
			return nil, errors.New("can't find such username in this orgnazition.")
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
				if err = user.UpdateRoleBinding(org); err != nil {
					return
				}
				if org, err = org.Update(); err == nil {
					return
				} else {
					glog.Error(err)
					return
				}
			}
		}
		return nil, errors.New("no such user.")
	}
	return
}

func (usr *UserInfo) Update() error {
	return dbstore.SetValue(etcdProfilePath(usr.Username), usr, false)
}

func (usr *UserInfo) AddToEtcd() (*UserInfo, error) {

	usr.Password = ""
	usr.Status.Phase = UserStatusInactive
	usr.Status.Active = false
	usr.Status.Initialized = false
	usr.CreateTime = time.Now().Format(time.RFC3339)
	usr.Status.Quota.OrgQuota = 1
	err := dbstore.SetValue(etcdProfilePath(usr.Username), usr, false)

	return usr, err
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
	message := fmt.Sprintf(Message, usr.Username, link, link)
	return messages.SendMail([]string{usr.Email}, []string{}, bccEmail, Subject, message, true)
}

func (user *UserInfo) InitUserProject(token string) (err error) {
	project_url := DF_HOST + "/oapi/v1/projectrequests"

	project := new(oapi.ProjectRequest)
	project.Name = user.Username
	project.DisplayName = user.Username
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
				err = user.Update()
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
