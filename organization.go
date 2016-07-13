package main

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func DeleteOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	http.Error(w, "delete not implentmented.", http.StatusNotImplemented)
}

func JoinOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	action := ps.ByName("action")
	orgID := ps.ByName("org")

	glog.Infof("action: %s,orgID: %s", action, orgID)

	err = user.OrgJoin(orgID)

	if err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, "", http.StatusOK)
	}
}

func LeaveOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	action := ps.ByName("action")
	orgID := ps.ByName("org")

	glog.Infof("action: %s,orgID: %s", action, orgID)

	if !user.CheckIfOrgExistByID(orgID) {
		http.Error(w, "no such organization", http.StatusNotFound)
		return
	}

	err = user.OrgLeave(orgID)

	if err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, "", http.StatusOK)
	}

	return
}

func ListOrganizations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if orgs, err := userinfo.ListOrgs(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		resp, _ := json.Marshal(orgs)
		http.Error(w, string(resp), http.StatusOK)
	}
	return

}

func CreateOrganization(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)

	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	org := new(Orgnazition)
	if err := parseRequestBody(r, org); err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(org.Name) < 2 {

		http.Error(w, "organization name is too short.", http.StatusBadRequest)
		return
	}

	if userinfo.CheckIfOrgExist(org.Name) {
		http.Error(w, "organization name already exist.", http.StatusBadRequest)
		return
	}
	userinfo.token, _ = checkToken(r)
	if org, err = userinfo.CreateOrg(org); err != nil {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		if resp, err := json.Marshal(org); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Error(w, string(resp), http.StatusOK)
		}
		return
	}

}

func GetOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userinfo := &UserInfo{Username: username}
	if userinfo, err = userinfo.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID := ps.ByName("org")
	if !userinfo.CheckIfOrgExistByID(orgID) {
		http.Error(w, "no such organization", http.StatusNotFound)
		return
	}

	orgnazition := new(Orgnazition)
	orgnazition.ID = orgID

	if orgnazition, err = orgnazition.Get(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		resp, _ := json.Marshal(orgnazition)
		http.Error(w, string(resp), http.StatusOK)
	}
	return
}

func ManageOrganization(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	glog.Infoln("from", r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto)
	var username string
	var err error

	if username, err = authedIdentities(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &UserInfo{Username: username}
	if user, err = user.Get(); err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "no such organization", http.StatusNotFound)
		return
	}

	member := new(OrgMember)
	if err := parseRequestBody(r, member); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch action {
	case "remove":

		err = user.OrgRemoveMember(member, orgID)
	case "invite":

		var org *Orgnazition
		org, err = user.OrgInvite(member, orgID)
		if err == nil {
			go SendOrgInviteMessage(member.MemberName, username, orgID, org.Name)
		}
	case "privileged":

		err = user.OrgPrivilege(member, orgID)
	default:
		http.Error(w, "not supported action", http.StatusBadRequest)
		return
	}

	if err != nil {
		glog.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		http.Error(w, "", http.StatusOK)
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
