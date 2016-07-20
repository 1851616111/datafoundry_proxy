package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func DeleteOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	orgID := ps.ByName("org")
	if !userinfo.CheckIfOrgExistByID(orgID) {
		RespError(w, ldpErrorNew(ErrCodeOrgNotFound), http.StatusNotFound)
		return
	}

	userinfo.token, _ = checkToken(r)
	org := new(Orgnazition)
	if org, err = userinfo.DeleteOrg(orgID); err != nil {

		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, org)
	}

}

func JoinOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	action := ps.ByName("action")
	orgID := ps.ByName("org")

	glog.Infof("action: %s,orgID: %s", action, orgID)

	err = user.OrgJoin(orgID)

	if err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, nil)
	}
}

func LeaveOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	action := ps.ByName("action")
	orgID := ps.ByName("org")

	glog.Infof("action: %s,orgID: %s", action, orgID)

	if !user.CheckIfOrgExistByID(orgID) {
		RespError(w, ldpErrorNew(ErrCodeOrgNotFound), http.StatusNotFound)
		return
	}

	err = user.OrgLeave(orgID)

	if err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, nil)
	}

	return
}

func ListOrganizations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	if orgs, err := userinfo.ListOrgs(); err != nil {
		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, orgs)
	}
	return

}

func CreateOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	org := new(Orgnazition)
	if err := parseRequestBody(r, org); err != nil {

		RespError(w, err, http.StatusBadRequest)
		return
	}

	if len(org.Name) < 2 {

		RespError(w, ldpErrorNew(ErrCodeOrgNameTooShort), http.StatusBadRequest)
		return
	}

	if userinfo.CheckIfOrgExist(org.Name) {
		RespError(w, ldpErrorNew(ErrCodeOrgExist), http.StatusBadRequest)
		return
	}
	userinfo.token, _ = checkToken(r)
	if org, err = userinfo.CreateOrg(org); err != nil {

		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, org)
	}
	return

}

func GetOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	orgID := ps.ByName("org")
	if !userinfo.CheckIfOrgExistByID(orgID) {
		RespError(w, ldpErrorNew(ErrCodeOrgNotFound), http.StatusNotFound)
		return
	}

	orgnazition := new(Orgnazition)
	orgnazition.ID = orgID

	if orgnazition, err = orgnazition.Get(); err != nil {
		RespError(w, err, http.StatusBadRequest)
	} else {
		RespOK(w, orgnazition)
	}
	return
}

func ManageOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		RespError(w, err, http.StatusUnauthorized)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	}

	user.token, _ = checkToken(r)
	action := ps.ByName("action")
	orgID := ps.ByName("org")

	glog.Infof("action: %s,orgID: %s", action, orgID)

	switch action {
	case "accept":
		JoinOrganization(w, r, ps)
		return
	case "leave":
		LeaveOrganization(w, r, ps)
		return
	}

	if !user.CheckIfOrgExistByID(orgID) {
		RespError(w, ldpErrorNew(ErrCodeOrgNotFound), http.StatusNotFound)
		return
	}

	member := new(OrgMember)
	if err := parseRequestBody(r, member); err != nil {
		RespError(w, err, http.StatusBadRequest)
		return
	}

	var org *Orgnazition
	switch action {
	case "remove":

		org, err = user.OrgRemoveMember(member, orgID)
	case "invite":

		org, err = user.OrgInvite(member, orgID)
		if err == nil {
			go SendOrgInviteMessage(member.MemberName, username, orgID, org.Name)
		}
	case "privileged":

		org, err = user.OrgPrivilege(member, orgID)
	default:
		RespError(w, ldpErrorNew(ErrCodeActionNotSupport), http.StatusBadRequest)
		return
	}

	if err != nil {
		glog.Error(err)
		RespError(w, err, http.StatusBadRequest)
		return
	} else {
		RespOK(w, org)
		return

	}

}

func (o *Orgnazition) Get() (org *Orgnazition, err error) {
	obj, err := dbstore.GetValue(etcdOrgPath(o.ID))
	if err != nil {
		return nil, err
	}
	org = new(Orgnazition)
	if err = json.Unmarshal([]byte(obj.(string)), org); err != nil {
		glog.Error(err)
		return nil, err
	}
	return org, nil

}
func (o *Orgnazition) Delete() (err error) {
	if err = dbstore.Delete(etcdOrgPath(o.ID), false); err != nil {
		glog.Error(err)
	}
	return
}

func (o *Orgnazition) Update() (org *Orgnazition, err error) {

	return o.Create()
}

func (o *Orgnazition) Create() (org *Orgnazition, err error) {
	if err = dbstore.SetValue(etcdOrgPath(o.ID), o, false); err != nil {
		glog.Error(err)
		return nil, err
	}
	return o, nil
}

func (o *Orgnazition) IsAdmin(username string) bool {
	for _, member := range o.MemberList {
		if member.MemberName == username && member.IsAdmin {
			return true
		}
	}
	return false
}

func (o *Orgnazition) IsLastAdmin(username string) bool {
	adminCnt := 0
	isAdmin := false
	for _, member := range o.MemberList {
		if member.MemberName == username && member.IsAdmin {
			isAdmin = true
		}
		if member.IsAdmin && member.Status == OrgMemberStatusjoined {
			adminCnt += 1
		}
	}
	return isAdmin && (adminCnt == 1)
}

func (o *Orgnazition) IsMemberExist(member *OrgMember) bool {
	for _, v := range o.MemberList {
		if v.MemberName == member.MemberName {
			return true
		}
	}
	return false
}

func (o *Orgnazition) JoinedMemberCnt() int {
	cnt := 0
	for _, v := range o.MemberList {
		if v.Status == OrgMemberStatusjoined {
			cnt += 1
		}
	}
	return cnt
}

func (o *Orgnazition) MemberStatus(member *OrgMember) MemberStatusPhase {
	for _, v := range o.MemberList {
		if v.MemberName == member.MemberName {
			return v.Status
		}
	}
	return OrgMemberStatusNone
}

func (o *Orgnazition) RemoveMember(member string) *Orgnazition {
	for idx, v := range o.MemberList {
		if v.MemberName == member {
			o.MemberList = append(o.MemberList[:idx], o.MemberList[idx+1:]...)
		}
	}
	return o
}

func (o *Orgnazition) MemberJoined(member string) *Orgnazition {
	for idx, v := range o.MemberList {
		if v.MemberName == member {
			o.MemberList[idx].Status = OrgMemberStatusjoined
			o.MemberList[idx].JoinedAt = time.Now().Format(time.RFC3339)
		}
	}
	return o
}

func (o *Orgnazition) AddMemeber(member *OrgMember) *Orgnazition { return nil }
